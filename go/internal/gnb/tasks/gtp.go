package tasks

import (
	"context"
	"hash/crc32"
	"net"
	"time"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/gnbctx"
	"github.com/acore2026/ueransim-go/internal/gtp"
	"github.com/acore2026/ueransim-go/internal/rlc"
)

const (
	MessageTypeRlsToGtp runtime.MessageType = "rls_to_gtp"
)

type GnbGtpTaskHandler struct {
	logger       logging.Logger
	listenAddr   string
	sessionStore *gnbctx.SessionStore
	rlcTask      *runtime.Task
	conn         *net.UDPConn
	lastDl       map[uint8]recentDownlink
}

type recentDownlink struct {
	teid uint32
	sum  uint32
	size int
	at   time.Time
}

func NewGnbGtpTaskHandler(logger logging.Logger, listenAddr string, sessionStore *gnbctx.SessionStore, rlcTask *runtime.Task) *GnbGtpTaskHandler {
	return &GnbGtpTaskHandler{
		logger:       logger.With("component", "gtp"),
		listenAddr:   listenAddr,
		sessionStore: sessionStore,
		rlcTask:      rlcTask,
		lastDl:       make(map[uint8]recentDownlink),
	}
}

func (h *GnbGtpTaskHandler) OnStart(ctx context.Context, task *runtime.Task) error {
	addr, err := net.ResolveUDPAddr("udp", h.listenAddr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	h.conn = conn
	h.logger.Info("GTP-U task started", "addr", conn.LocalAddr().String())

	go h.readLoop(ctx, task)
	return nil
}

func (h *GnbGtpTaskHandler) readLoop(ctx context.Context, task *runtime.Task) {
	buf := make([]byte, 64*1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_ = h.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, _, err := h.conn.ReadFromUDP(buf)
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				if ctx.Err() != nil {
					return
				}
				h.logger.Error("GTP-U read failed", "error", err)
				continue
			}
			msg, err := gtp.Decode(buf[:n])
			if err != nil {
				h.logger.Error("failed to decode GTP-U packet", "error", err)
				continue
			}
			if msg.MsgType != gtp.MT_G_PDU {
				h.logger.Info("ignoring non-G-PDU packet", "type", msg.MsgType)
				continue
			}

			session, ok := h.sessionStore.ByLocalTEID(msg.Teid)
			if !ok {
				h.logger.Info("dropping GTP-U packet for unknown TEID", "teid", msg.Teid)
				continue
			}
			if h.isDuplicateDownlink(session.SessionID, msg.Teid, msg.Payload) {
				h.logger.Info("dropping duplicate downlink GTP-U packet", "sessionID", session.SessionID, "teid", msg.Teid, "len", len(msg.Payload))
				continue
			}
			h.logger.Info("received downlink GTP-U packet", "sessionID", session.SessionID, "teid", msg.Teid, "len", len(msg.Payload))
			
			_ = h.rlcTask.Send(runtime.Message{
				Type: "upper_to_rlc",
				Payload: rlc.UpperToRlcMessage{
					Mode:      rlc.ModeUM,
					Pdu:       append([]byte(nil), msg.Payload...),
					SessionID: session.SessionID,
				},
			})
		}
	}
}

func (h *GnbGtpTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case MessageTypeRlsToGtp:
		payload := msg.Payload.(gnbctx.UplinkPacket)
		session, ok := h.sessionStore.BySessionID(payload.SessionID)
		if !ok {
			h.logger.Info("dropping uplink packet for unknown session", "sessionID", payload.SessionID)
			return nil
		}
		remoteAddr, err := session.RemoteAddr()
		if err != nil {
			return err
		}
		packet, err := (&gtp.GtpMessage{
			EFlag:   len(session.QFIs) > 0,
			MsgType: gtp.MT_G_PDU,
			Teid:    session.RemoteTEID,
			Payload: payload.Data,
			QFI:     firstQFI(session.QFIs),
		}).Encode()
		if err != nil {
			return err
		}
		if _, err := h.conn.WriteToUDP(packet, remoteAddr); err != nil {
			return err
		}
		h.logger.Info("sent uplink GTP-U packet", "sessionID", session.SessionID, "teid", session.RemoteTEID, "len", len(payload.Data))
	}
	return nil
}

func firstQFI(qfis []uint8) *uint8 {
	if len(qfis) == 0 {
		return nil
	}
	qfi := qfis[0]
	return &qfi
}

func (h *GnbGtpTaskHandler) isDuplicateDownlink(sessionID uint8, teid uint32, payload []byte) bool {
	now := time.Now()
	sum := crc32.ChecksumIEEE(payload)
	last, ok := h.lastDl[sessionID]
	h.lastDl[sessionID] = recentDownlink{
		teid: teid,
		sum:  sum,
		size: len(payload),
		at:   now,
	}
	if !ok {
		return false
	}
	if now.Sub(last.at) > 20*time.Millisecond {
		return false
	}
	return last.teid == teid && last.size == len(payload) && last.sum == sum
}

func (h *GnbGtpTaskHandler) OnStop(ctx context.Context) error {
	if h.conn != nil {
		return h.conn.Close()
	}
	return nil
}
