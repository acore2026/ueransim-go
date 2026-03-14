package gnb

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/acore2026/ueransim-go/internal/config"
	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/gnb/tasks"
	"github.com/acore2026/ueransim-go/internal/utils"
)

type Node struct {
	cfg    *config.GNBConfig
	logger logging.Logger
	group  *runtime.Group
}

func New(cfg *config.GNBConfig, logger logging.Logger) *Node {
	sctpLogger := logger.With("subsystem", "sctp")
	ngapLogger := logger.With("subsystem", "ngap")

	// 3. NGAP Task
	gnbId, _ := hex.DecodeString(cfg.GNBID)
	plmnId := utils.EncodePlmn(cfg.MCC, cfg.MNC)
	
	// Create tasks without handlers first
	ngapTask := runtime.NewTask("gnb-ngap", ngapLogger, nil, 64)
	sctpTask := runtime.NewTask("gnb-sctp", sctpLogger, nil, 64)

	// 2. RLS Task
	rlsLogger := logger.With("subsystem", "rls")
	rlsAddr := fmt.Sprintf("%s:%d", cfg.LinkIP, 38412)
	rlsHandler := NewRlsTaskHandler(rlsLogger, rlsAddr, ngapTask)
	rlsTask := runtime.NewTask("gnb-rls", rlsLogger, rlsHandler, 64)

	// Create remaining tasks
	gtpLogger := logger.With("subsystem", "gtp")
	gtpTask := runtime.NewTask("gnb-gtp", gtpLogger, newHeartbeatHandler(gtpLogger, "gnb-gtp"), 16)

	// Now create handlers with task pointers
	amfAddr := "127.0.0.1"
	amfPort := 38412
	if len(cfg.AMFConfigs) > 0 {
		amfAddr = cfg.AMFConfigs[0].Address
		amfPort = cfg.AMFConfigs[0].Port
	}

	ngapHandler := tasks.NewGnbNgapTaskHandler(ngapLogger, cfg.NodeName(), gnbId, plmnId, sctpTask, rlsTask)
	sctpHandler := tasks.NewGnbSctpTaskHandler(sctpLogger, amfAddr, amfPort, ngapTask)

	// Set handlers
	ngapTask.SetHandler(ngapHandler)
	sctpTask.SetHandler(sctpHandler)

	return &Node{
		cfg:    cfg,
		logger: logger,
		group:  runtime.NewGroup(logger, sctpTask, ngapTask, gtpTask, rlsTask),
	}
}

func (n *Node) Run(ctx context.Context) error {
	n.logger.Info("bootstrapping gNB",
		"name", n.cfg.NodeName(),
		"mcc", n.cfg.MCC,
		"mnc", n.cfg.MNC,
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
	return nil
}

func (h *heartbeatHandler) OnStop(context.Context) error {
	h.logger.Info("shutdown complete")
	return nil
}
