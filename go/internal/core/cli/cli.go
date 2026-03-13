package cli

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
)

type CliHandler struct {
	logger     logging.Logger
	targetTask *runtime.Task
}

func NewCliHandler(logger logging.Logger, targetTask *runtime.Task) *CliHandler {
	return &CliHandler{
		logger:     logger.With("component", "cli"),
		targetTask: targetTask,
	}
}

func (h *CliHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("CLI handler started")
	go h.inputLoop(ctx)
	return nil
}

func (h *CliHandler) inputLoop(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if scanner.Scan() {
				text := scanner.Text()
				h.processCommand(text)
			}
		}
	}
}

func (h *CliHandler) processCommand(line string) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	switch cmd {
	case "exit", "quit":
		h.logger.Info("exit requested via CLI")
		os.Exit(0)
	default:
		h.logger.Info("unknown command", "cmd", cmd)
	}
}

func (h *CliHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	return nil
}

func (h *CliHandler) OnStop(ctx context.Context) error {
	return nil
}
