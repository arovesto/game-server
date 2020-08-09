package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"net"
	"os"

	"github.com/arovesto/game-server/internal/config"
	"github.com/arovesto/game-server/internal/log"
	"github.com/arovesto/game-server/pkg/server"
)

var cfgPath = flag.String("config", "config.toml", "path to config")

func init() {
	flag.Parse()
}

func main() {
	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	logger, err := log.SetupLogger(&cfg.Logging)
	if err != nil {
		panic(fmt.Sprintf("failed to start logger: %v", err))
	}

	ctx := log.WithLogger(context.Background(), logger)

	log.G(ctx).Info("Starting server...")

	srv, err := server.NewServer(cfg)
	if err != nil {
		log.G(ctx).Panic("failed to create server", zap.Error(err))
	}
	listener, err := net.Listen("tcp", cfg.General.Address)
	if err != nil {
		log.G(ctx).Panic("failed to start listener: %v", zap.Error(err))
	}

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		log.G(ctx).Info("starting server console...")
		if err := srv.StartExec(ctx, scanner); err != nil {
			log.G(ctx).Panic("failed to start server console: %v", zap.Error(err))
		}
	}()

	if err := srv.Start(ctx, listener); err != nil {
		log.G(ctx).Panic("failed to start server: %v", zap.Error(err))
	}
}
