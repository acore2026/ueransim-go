package gnb

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/acore2026/ueransim-go/internal/config"
	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/gnb/tasks"
	"github.com/acore2026/ueransim-go/internal/gnbctx"
	"github.com/acore2026/ueransim-go/internal/ngap"
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
	sessionStore := gnbctx.NewSessionStore()
	rlsTask := runtime.NewTask("gnb-rls", rlsLogger, nil, 64)

	gtpLogger := logger.With("subsystem", "gtp")
	gtpAddr := fmt.Sprintf("%s:%d", cfg.GTPIP, gnbctx.GTPUPort)
	gtpHandler := tasks.NewGnbGtpTaskHandler(gtpLogger, gtpAddr, sessionStore, rlsTask)
	gtpTask := runtime.NewTask("gnb-gtp", gtpLogger, gtpHandler, 64)

	rlsHandler := NewRlsTaskHandler(rlsLogger, rlsAddr, ngapTask, gtpTask)
	rlsTask.SetHandler(rlsHandler)

	// Now create handlers with task pointers
	amfAddr := "127.0.0.1"
	amfPort := 38412
	if len(cfg.AMFConfigs) > 0 {
		amfAddr = cfg.AMFConfigs[0].Address
		amfPort = cfg.AMFConfigs[0].Port
	}

	tac := []byte{byte(cfg.TAC >> 16), byte(cfg.TAC >> 8), byte(cfg.TAC)}
	nrCellID := buildNrCellIdentity(cfg.NCI)
	uli := ngap.BuildUserLocationInformationNR(plmnId, tac, nrCellID)
	ngapHandler := tasks.NewGnbNgapTaskHandler(ngapLogger, cfg.NodeName(), gnbId, plmnId, uli, cfg.GTPIP, sctpTask, rlsTask, sessionStore)
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

func buildNrCellIdentity(nci string) []byte {
	trimmed := strings.TrimPrefix(strings.TrimPrefix(nci, "0x"), "0X")
	if trimmed == "" {
		return []byte{0x00, 0x00, 0x00, 0x00, 0x10}
	}
	value, err := strconv.ParseUint(trimmed, 16, 64)
	if err != nil {
		return []byte{0x00, 0x00, 0x00, 0x00, 0x10}
	}
	return []byte{
		byte(value >> 32),
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
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
