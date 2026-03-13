package gnb

import (
	"context"
	"fmt"
	"time"

	"github.com/acore2026/ueransim-go/internal/config"
	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
)

type Node struct {
	cfg    *config.GNBConfig
	logger logging.Logger
	group  *runtime.Group
}

func New(cfg *config.GNBConfig, logger logging.Logger) *Node {
	appLogger := logger.With("subsystem", "app")
	sctpLogger := logger.With("subsystem", "sctp")
	ngapLogger := logger.With("subsystem", "ngap")
	rrcLogger := logger.With("subsystem", "rrc")
	gtpLogger := logger.With("subsystem", "gtp")
	rlsLogger := logger.With("subsystem", "rls")

	appTask := runtime.NewTask("gnb-app", appLogger, newHeartbeatHandler(appLogger, "gnb-app"), 16)
	sctpTask := runtime.NewTask("gnb-sctp", sctpLogger, newHeartbeatHandler(sctpLogger, "gnb-sctp"), 16)
	ngapTask := runtime.NewTask("gnb-ngap", ngapLogger, newHeartbeatHandler(ngapLogger, "gnb-ngap"), 16)
	rrcTask := runtime.NewTask("gnb-rrc", rrcLogger, newHeartbeatHandler(rrcLogger, "gnb-rrc"), 16)
	gtpTask := runtime.NewTask("gnb-gtp", gtpLogger, newHeartbeatHandler(gtpLogger, "gnb-gtp"), 16)
	rlsTask := runtime.NewTask("gnb-rls", rlsLogger, newHeartbeatHandler(rlsLogger, "gnb-rls"), 16)

	return &Node{
		cfg:    cfg,
		logger: logger,
		group:  runtime.NewGroup(logger, appTask, sctpTask, ngapTask, rrcTask, gtpTask, rlsTask),
	}
}

func (n *Node) Run(ctx context.Context) error {
	n.logger.Info("bootstrapping gNB",
		"name", n.cfg.NodeName(),
		"mcc", n.cfg.MCC,
		"mnc", n.cfg.MNC,
		"amfCount", len(n.cfg.AMFConfigs),
		"sliceCount", len(n.cfg.Slices),
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
