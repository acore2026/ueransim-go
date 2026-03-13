package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/acore2026/ueransim-go/internal/config"
	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/ue"
)

func main() {
	configPath := flag.String("config", "../config/free5gc-ue.yaml", "path to UE config file")
	flag.Parse()

	logger := logging.New("nr-ue")

	cfg, err := config.LoadUE(*configPath)
	if err != nil {
		logger.Error("load config", "path", *configPath, "error", err)
		panic(err)
	}

	node := ue.New(cfg, logger.With("node", cfg.NodeName()))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := node.Run(ctx); err != nil {
		logger.Error("node exited", "error", err)
		panic(err)
	}
}
