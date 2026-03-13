package sctp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/ishidawataru/sctp"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
)

const (
	MessageTypeSctpReceive runtime.MessageType = "sctp_receive"
	MessageTypeSctpSend    runtime.MessageType = "sctp_send"
	MessageTypeSctpError   runtime.MessageType = "sctp_error"
)

type ReceiveMessage struct {
	Stream uint16
	Ppid   uint32
	Data   []byte
}

type SendMessage struct {
	Stream uint16
	Ppid   uint32
	Data   []byte
}

type ErrorMessage struct {
	Error string
}

// ClientTaskHandler manages an SCTP connection to a remote server.
type ClientTaskHandler struct {
	localAddr  string
	remoteAddr string
	remotePort int
	
	conn       *sctp.SCTPConn
	targetTask *runtime.Task
	logger     logging.Logger
	wg         sync.WaitGroup
}

// NewClientTaskHandler creates a new SCTP client handler.
func NewClientTaskHandler(localAddr, remoteAddr string, remotePort int, targetTask *runtime.Task, logger logging.Logger) *ClientTaskHandler {
	return &ClientTaskHandler{
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
		remotePort: remotePort,
		targetTask: targetTask,
		logger:     logger.With("component", "sctp_client"),
	}
}

func (h *ClientTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	var laddr *sctp.SCTPAddr
	if h.localAddr != "" {
		ip := net.ParseIP(h.localAddr)
		if ip == nil {
			return fmt.Errorf("invalid local address: %s", h.localAddr)
		}
		laddr = &sctp.SCTPAddr{
			IPAddrs: []net.IPAddr{{IP: ip}},
			Port:    0,
		}
	}

	remoteIP := net.ParseIP(h.remoteAddr)
	if remoteIP == nil {
		return fmt.Errorf("invalid remote address: %s", h.remoteAddr)
	}
	raddr := &sctp.SCTPAddr{
		IPAddrs: []net.IPAddr{{IP: remoteIP}},
		Port:    h.remotePort,
	}

	h.logger.Info("connecting to SCTP server", "remote", raddr.String())
	
	conn, err := sctp.DialSCTP("sctp", laddr, raddr)
	if err != nil {
		h.logger.Error("SCTP connection failed", "error", err)
		return fmt.Errorf("failed to connect: %w", err)
	}
	h.logger.Info("SCTP connected")

	h.conn = conn
	
	// Setup InitMsg similar to C++ (10, 10, 10, 10000)
	// This library sets defaults internally, but we can set InitMsg if needed
	// using sctp.InitMsg. For now, rely on standard OS defaults for simplicity.
	
	// Enable receiving SCTP_SNDRCV events
	if err := conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO); err != nil {
		h.logger.Error("failed to subscribe to SCTP events", "error", err)
		// continue anyway as some kernels handle it differently
	}

	h.wg.Add(1)
	go h.readLoop(ctx)

	return nil
}

func (h *ClientTaskHandler) readLoop(ctx context.Context) {
	defer h.wg.Done()
	
	buffer := make([]byte, 65535)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read with a deadline to allow checking for ctx cancellation
			h.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			
			n, info, err := h.conn.SCTPRead(buffer)
			if err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
					return
				}
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				h.logger.Error("SCTP read failed", "error", err)
				h.sendError(fmt.Sprintf("SCTP read failed: %v", err))
				return // Terminate read loop on fatal errors
			}

			if n > 0 && info != nil {
				dataCopy := make([]byte, n)
				copy(dataCopy, buffer[:n])
				
				msg := runtime.Message{
					Type: MessageTypeSctpReceive,
					Payload: ReceiveMessage{
						Stream: info.Stream,
						Ppid:   info.PPID,
						Data:   dataCopy,
					},
				}
				if err := h.targetTask.Send(msg); err != nil {
					h.logger.Error("failed to send to target task", "error", err)
				}
			} else if n > 0 && info == nil {
				h.logger.Debug("SCTP received non-data message", "size", n)
			}
		}
	}
}

func (h *ClientTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case MessageTypeSctpSend:
		h.logger.Info("received SCTP send request", "len", len(msg.Payload.(SendMessage).Data))
		payload, ok := msg.Payload.(SendMessage)
		if !ok {
			return errors.New("invalid payload for MessageTypeSctpSend")
		}
		
		if h.conn == nil {
			return errors.New("connection not established")
		}

		info := &sctp.SndRcvInfo{
			Stream: uint16(payload.Stream),
			PPID:   uint32(payload.Ppid),
		}
		
		_, err := h.conn.SCTPWrite(payload.Data, info)
		if err != nil {
			h.logger.Error("SCTP write failed", "error", err)
			h.sendError(fmt.Sprintf("SCTP write failed: %v", err))
		}
	}
	return nil
}

func (h *ClientTaskHandler) OnStop(ctx context.Context) error {
	h.logger.Info("closing SCTP client connection")
	var err error
	if h.conn != nil {
		err = h.conn.Close()
	}
	h.wg.Wait()
	return err
}

func (h *ClientTaskHandler) sendError(errStr string) {
	_ = h.targetTask.Send(runtime.Message{
		Type: MessageTypeSctpError,
		Payload: ErrorMessage{
			Error: errStr,
		},
	})
}
