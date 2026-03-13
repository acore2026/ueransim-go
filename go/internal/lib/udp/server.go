package udp

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
)

const (
	// BufferSize matches the C++ BUFFER_SIZE definition.
	BufferSize = 65536
	// MessageTypeUdpReceive is the type of message sent to the target task.
	MessageTypeUdpReceive runtime.MessageType = "udp_receive"
)

// ReceiveMessage is the payload sent to the target task when data is received.
type ReceiveMessage struct {
	Data []byte
	From *net.UDPAddr
}

// ServerTaskHandler implements runtime.TaskHandler for a UDP server.
type ServerTaskHandler struct {
	addr       *net.UDPAddr
	conn       *net.UDPConn
	targetTask *runtime.Task
	logger     logging.Logger
	wg         sync.WaitGroup
}

// NewServerTaskHandler creates a new ServerTaskHandler.
// If addr is nil, it will listen on an ephemeral port.
func NewServerTaskHandler(addr *net.UDPAddr, targetTask *runtime.Task, logger logging.Logger) *ServerTaskHandler {
	return &ServerTaskHandler{
		addr:       addr,
		targetTask: targetTask,
		logger:     logger.With("component", "udp_server"),
	}
}

// OnStart initializes the UDP socket and starts the read loop.
func (h *ServerTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	addrStr := "ephemeral"
	if h.addr != nil {
		addrStr = h.addr.String()
	}
	h.logger.Info("starting UDP server", "address", addrStr)

	conn, err := net.ListenUDP("udp", h.addr)
	if err != nil {
		return err
	}
	h.conn = conn

	// Start the read loop in a background goroutine.
	h.wg.Add(1)
	go h.readLoop(ctx)

	return nil
}

func (h *ServerTaskHandler) readLoop(ctx context.Context) {
	defer h.wg.Done()

	buffer := make([]byte, BufferSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Set a read deadline to allow checking for ctx cancellation
			// Mimics the 500ms timeout in the C++ version.
			_ = h.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, addr, err := h.conn.ReadFromUDP(buffer)
			if err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, context.Canceled) {
					return
				}
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				h.logger.Error("UDP read failed", "error", err)
				// Continue loop on other errors for resilience, similar to C++.
				continue
			}

			if n > 0 {
				dataCopy := make([]byte, n)
				copy(dataCopy, buffer[:n])
				msg := runtime.Message{
					Type: MessageTypeUdpReceive,
					Payload: ReceiveMessage{
						Data: dataCopy,
						From: addr,
					},
				}
				if err := h.targetTask.Send(msg); err != nil {
					h.logger.Error("failed to send UDP message to target task", "error", err)
				}
			}
		}
	}
}

// OnMessage processes incoming messages to the UDP server task.
func (h *ServerTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	return nil
}

// OnStop closes the UDP connection and waits for the read loop to exit.
func (h *ServerTaskHandler) OnStop(ctx context.Context) error {
	h.logger.Info("stopping UDP server")
	var err error
	if h.conn != nil {
		err = h.conn.Close()
	}
	// Wait for the read loop to gracefully exit.
	h.wg.Wait()
	return err
}

// Send transmits data to the specified UDP address.
func (h *ServerTaskHandler) Send(to *net.UDPAddr, data []byte) error {
	if h.conn == nil {
		return errors.New("UDP server not started")
	}
	_, err := h.conn.WriteToUDP(data, to)
	return err
}
