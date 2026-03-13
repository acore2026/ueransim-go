package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/acore2026/ueransim-go/internal/config"
	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/gnb"
)

func main() {
	configPath := flag.String("config", "../config/free5gc-gnb.yaml", "path to gNB config file")
	flag.Parse()

	logger := logging.New("nr-gnb")

	cfg, err := config.LoadGNB(*configPath)
	if err != nil {
		logger.Error("load config", "path", *configPath, "error", err)
		panic(err)
	}

	node := gnb.New(cfg, logger.With("node", cfg.NodeName()))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := node.Run(ctx); err != nil {
		logger.Error("node exited", "error", err)
		panic(err)
	}
}
