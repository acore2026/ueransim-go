package ue

import (
	"context"
	"fmt"
	"time"

	"github.com/acore2026/ueransim-go/internal/config"
	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
)

type Node struct {
	cfg    *config.UEConfig
	logger logging.Logger
	group  *runtime.Group
}

func New(cfg *config.UEConfig, logger logging.Logger) *Node {
	appLogger := logger.With("subsystem", "app")
	nasLogger := logger.With("subsystem", "nas")
	rrcLogger := logger.With("subsystem", "rrc")
	rlsLogger := logger.With("subsystem", "rls")

	appTask := runtime.NewTask("ue-app", appLogger, newHeartbeatHandler(appLogger, "ue-app"), 16)
	nasTask := runtime.NewTask("ue-nas", nasLogger, newHeartbeatHandler(nasLogger, "ue-nas"), 16)
	rrcTask := runtime.NewTask("ue-rrc", rrcLogger, newHeartbeatHandler(rrcLogger, "ue-rrc"), 16)
	rlsTask := runtime.NewTask("ue-rls", rlsLogger, newHeartbeatHandler(rlsLogger, "ue-rls"), 16)

	return &Node{
		cfg:    cfg,
		logger: logger,
		group:  runtime.NewGroup(logger, appTask, nasTask, rrcTask, rlsTask),
	}
}

func (n *Node) Run(ctx context.Context) error {
	n.logger.Info("bootstrapping UE",
		"supi", n.cfg.SUPI,
		"mcc", n.cfg.MCC,
		"mnc", n.cfg.MNC,
		"gnbSearchCount", len(n.cfg.GNBSearchList),
		"sessionCount", len(n.cfg.Sessions),
	)
	return n.group.Run(ctx)
}

type heartbeatHandler struct {
	name   string
	logger logging.Logger
	tick   runtime.PeriodicTask
}

func newHeartbeatHandler(logger logging.Logger, name string) *heartbeatHandler {
	return &heartbeatHandler{
		name:   name,
		logger: logger,
		tick:   runtime.NewPeriodicTask(10*time.Second, logger),
	}
}

func (h *heartbeatHandler) OnStart(ctx context.Context, task *runtime.Task) error {
	h.tick.Start(ctx, task)
	h.logger.Info("initialized")
	return nil
}

func (h *heartbeatHandler) OnMessage(_ context.Context, msg runtime.Message) error {
	if msg.Type == runtime.MessageTypeTick {
		h.logger.Info("heartbeat", "task", h.name)
		return nil
	}
	return fmt.Errorf("unexpected message type: %s", msg.Type)
}

func (h *heartbeatHandler) OnStop(context.Context) error {
	h.logger.Info("shutdown complete")
	return nil
}
