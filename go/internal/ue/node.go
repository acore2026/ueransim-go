package ue

import (
	"context"

	"github.com/acore2026/ueransim-go/internal/config"
	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/ue/tun"
)

type Node struct {
	cfg    *config.UEConfig
	logger logging.Logger
	group  *runtime.Group
}

func New(cfg *config.UEConfig, logger logging.Logger) *Node {
	rlsLogger := logger.With("subsystem", "rls")
	rrcLogger := logger.With("subsystem", "rrc")
	nasLogger := logger.With("subsystem", "nas")
	tunLogger := logger.With("subsystem", "tun")

	// 1. Setup RLS Task
	// Assuming first gNB from search list for now
	gnbAddr := "127.0.0.1:38412"
	if len(cfg.GNBSearchList) > 0 {
		gnbAddr = cfg.GNBSearchList[0]
	}

	rlsHandler, err := NewRlsTaskHandler(rlsLogger, gnbAddr, 1, nil, nil) // tasks will be set later
	if err != nil {
		logger.Error("failed to create RLS handler", "error", err)
		return nil
	}
	rlsTask := runtime.NewTask("ue-rls", rlsLogger, rlsHandler, 64)

	// 1.5 Setup TUN Task with delayed configuration from session accept.
	tunHandler := tun.NewTaskHandler("uesimtun%d", "", cfg.TUNNetmask, 1400, true, 1, rlsTask, tunLogger)
	tunTask := runtime.NewTask("ue-tun", tunLogger, tunHandler, 64)

	// 2. Setup RRC Task
	rrcHandler := NewRrcTaskHandler(rrcLogger, rlsTask, nil)
	rrcTask := runtime.NewTask("ue-rrc", rrcLogger, rrcHandler, 64)

	rlsHandler.SetRrcTask(rrcTask)
	rlsHandler.SetTunTask(tunTask)

	// 3. Setup NAS Task
	nasHandler := NewNasTaskHandler(nasLogger, cfg, rrcTask, rlsTask, tunTask)
	nasTask := runtime.NewTask("ue-nas", nasLogger, nasHandler, 64)

	rrcHandler.SetNasTask(nasTask)

	return &Node{
		cfg:    cfg,
		logger: logger,
		group:  runtime.NewGroup(logger, nasTask, rrcTask, rlsTask, tunTask),
	}
}

func (n *Node) Run(ctx context.Context) error {
	n.logger.Info("bootstrapping UE",
		"supi", n.cfg.SUPI,
		"mcc", n.cfg.MCC,
		"mnc", n.cfg.MNC,
	)
	return n.group.Run(ctx)
}
