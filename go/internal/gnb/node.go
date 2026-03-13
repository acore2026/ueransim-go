package gnb

import (
	"context"
	"encoding/hex"

	"github.com/acore2026/ueransim-go/internal/config"
	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/gnb/tasks"
)

type Node struct {
	cfg    *config.GNBConfig
	logger logging.Logger
	group  *runtime.Group
}

func New(cfg *config.GNBConfig, logger logging.Logger) *Node {
	sctpLogger := logger.With("subsystem", "sctp")
	ngapLogger := logger.With("subsystem", "ngap")

	// Placeholder for other tasks
	gtpLogger := logger.With("subsystem", "gtp")
	rlsLogger := logger.With("subsystem", "rls")
	gtpTask := runtime.NewTask("gnb-gtp", gtpLogger, newHeartbeatHandler(gtpLogger, "gnb-gtp"), 16)
	rlsTask := runtime.NewTask("gnb-rls", rlsLogger, newHeartbeatHandler(rlsLogger, "gnb-rls"), 16)

	// 1. NGAP Task
	gnbId, _ := hex.DecodeString(cfg.GNBID)
	plmnId, _ := hex.DecodeString(cfg.MCC + cfg.MNC) // Simplified
	
	// Pre-declare sctpTask so NGAP can reference it, but we need to create handlers first.
	// Actually, we need to create tasks in reverse order of dependency if they need pointers.
	
	// Create tasks without handlers first
	ngapTask := runtime.NewTask("gnb-ngap", ngapLogger, nil, 64)
	sctpTask := runtime.NewTask("gnb-sctp", sctpLogger, nil, 64)

	// Now create handlers with task pointers
	amfAddr := "127.0.0.1"
	amfPort := 38412
	if len(cfg.AMFConfigs) > 0 {
		amfAddr = cfg.AMFConfigs[0].Address
		amfPort = cfg.AMFConfigs[0].Port
	}

	ngapHandler := tasks.NewGnbNgapTaskHandler(ngapLogger, cfg.NodeName(), gnbId, plmnId, sctpTask)
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
