package tasks

import (
	"context"
	"net"
	"time"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/gnbctx"
	"github.com/acore2026/ueransim-go/internal/gtp"
)

const (
	MessageTypeRlsToGtp runtime.MessageType = "rls_to_gtp"
	MessageTypeGtpToRls runtime.MessageType = "gtp_to_rls"
)

type GnbGtpTaskHandler struct {
	logger       logging.Logger
	listenAddr   string
	sessionStore *gnbctx.SessionStore
	rlsTask      *runtime.Task
	conn         *net.UDPConn
}

func NewGnbGtpTaskHandler(logger logging.Logger, listenAddr string, sessionStore *gnbctx.SessionStore, rlsTask *runtime.Task) *GnbGtpTaskHandler {
	return &GnbGtpTaskHandler{
		logger:       logger.With("component", "gtp"),
		listenAddr:   listenAddr,
		sessionStore: sessionStore,
		rlsTask:      rlsTask,
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
			h.logger.Info("received downlink GTP-U packet", "sessionID", session.SessionID, "teid", msg.Teid, "len", len(msg.Payload))
			_ = h.rlsTask.Send(runtime.Message{
				Type: MessageTypeGtpToRls,
				Payload: gnbctx.DownlinkPacket{
					SessionID: session.SessionID,
					Data:      append([]byte(nil), msg.Payload...),
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

func (h *GnbGtpTaskHandler) OnStop(ctx context.Context) error {
	if h.conn != nil {
		return h.conn.Close()
	}
	return nil
}
