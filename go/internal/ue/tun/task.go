package tun

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
)

const (
	MessageTypeAppToTun  runtime.MessageType = "app_to_tun"
	MessageTypeTunToApp  runtime.MessageType = "tun_to_app"
	MessageTypeTunError  runtime.MessageType = "tun_error"
	MessageTypeConfigure runtime.MessageType = "configure_tun"
)

type AppToTunMessage struct {
	Data []byte
}

type TunToAppMessage struct {
	Psi  int
	Data []byte
}

type TunErrorMessage struct {
	Error string
}

type ConfigureMessage struct {
	IPAddress string
	Netmask   string
	Route     bool
}

type TaskHandler struct {
	deviceName string
	ipAddr     string
	netmask    string
	mtu        int
	route      bool
	psi        int

	targetTask      *runtime.Task
	logger          logging.Logger
	device          *Device
	configured      bool
	readLoopStarted bool
	ctx             context.Context
	wg              sync.WaitGroup
}

// NewTaskHandler creates a new TaskHandler for the TUN interface.
func NewTaskHandler(deviceName, ipAddr, netmask string, mtu int, route bool, psi int, targetTask *runtime.Task, logger logging.Logger) *TaskHandler {
	return &TaskHandler{
		deviceName: deviceName,
		ipAddr:     ipAddr,
		netmask:    netmask,
		mtu:        mtu,
		route:      route,
		psi:        psi,
		targetTask: targetTask,
		logger:     logger.With("component", "tun"),
	}
}

func (h *TaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.ctx = ctx
	h.logger.Info("allocating TUN device", "prefix", h.deviceName)
	dev, err := Allocate(h.deviceName)
	if err != nil {
		return fmt.Errorf("failed to allocate TUN: %w", err)
	}
	h.device = dev
	h.logger.Info("TUN allocated", "name", dev.Name())

	if h.ipAddr != "" {
		if err := h.configure(); err != nil {
			h.device.Close()
			return err
		}
	}

	return nil
}

func (h *TaskHandler) readLoop(ctx context.Context) {
	defer h.wg.Done()

	buffer := make([]byte, 8000)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := h.device.Read(buffer)
			if err != nil {
				if errors.Is(err, os.ErrClosed) {
					return
				}
				h.logger.Error("TUN read failed", "error", err)
				h.sendError(fmt.Sprintf("TUN read failed: %v", err))
				return
			}
			if n > 0 {
				dataCopy := make([]byte, n)
				copy(dataCopy, buffer[:n])

				msg := runtime.Message{
					Type: MessageTypeTunToApp,
					Payload: TunToAppMessage{
						Psi:  h.psi,
						Data: dataCopy,
					},
				}
				if err := h.targetTask.Send(msg); err != nil {
					h.logger.Error("failed to send to target task", "error", err)
				}
			}
		}
	}
}

func (h *TaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case MessageTypeAppToTun:
		payload, ok := msg.Payload.(AppToTunMessage)
		if !ok {
			return errors.New("invalid payload for MessageTypeAppToTun")
		}

		n, err := h.device.Write(payload.Data)
		if err != nil {
			h.logger.Error("TUN write failed", "error", err)
			h.sendError(fmt.Sprintf("TUN write failed: %v", err))
		} else if n != len(payload.Data) {
			h.logger.Error("TUN write partial")
			h.sendError("TUN device partially written")
		}
	case MessageTypeConfigure:
		payload, ok := msg.Payload.(ConfigureMessage)
		if !ok {
			return errors.New("invalid payload for MessageTypeConfigure")
		}
		h.ipAddr = payload.IPAddress
		h.netmask = payload.Netmask
		h.route = payload.Route
		if err := h.configure(); err != nil {
			return err
		}
	}
	return nil
}

func (h *TaskHandler) OnStop(ctx context.Context) error {
	h.logger.Info("stopping TUN device")
	var err error
	if h.device != nil {
		err = h.device.Close()
	}
	// Note: the read loop will error out and finish because the file is closed.
	// But it might be blocking on Read. Closing the file interrupts Read.
	h.wg.Wait()
	return err
}

func (h *TaskHandler) sendError(errStr string) {
	_ = h.targetTask.Send(runtime.Message{
		Type: MessageTypeTunError,
		Payload: TunErrorMessage{
			Error: errStr,
		},
	})
}

func (h *TaskHandler) configure() error {
	if h.configured {
		return nil
	}
	h.logger.Info("configuring TUN device", "ip", h.ipAddr)
	if err := h.device.Configure(h.ipAddr, h.netmask, h.mtu, h.route); err != nil {
		return fmt.Errorf("failed to configure TUN: %w", err)
	}
	h.configured = true
	if !h.readLoopStarted {
		h.wg.Add(1)
		h.readLoopStarted = true
		go h.readLoop(h.ctx)
	}
	return nil
}
