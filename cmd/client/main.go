package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"go.uber.org/zap"

	"github.com/arovesto/game-server/internal/config"
	"github.com/arovesto/game-server/internal/log"
	"github.com/arovesto/game-server/pkg/worker"
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

	log.G(ctx).Info("Starting client...")

	l, err := net.Dial("tcp", cfg.General.Address)
	if err != nil {
		log.G(ctx).Panic("failed to call address", zap.Error(err))
	}
	scanner := bufio.NewScanner(os.Stdin)
	log.G(ctx).Info("starting client console...")
	fmt.Println("================================================================")
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.G(ctx).Panic("scanner error", zap.Error(err))
		}
		message := scanner.Bytes()

		fmt.Println("[SEND]", message, len(message))
		cnt, err := l.Write(message)
		if err != nil || cnt != worker.MessageLength {
			log.G(ctx).Panic("error from connection", zap.Error(err), zap.Int("message-lengtth", cnt))
		}
		cnt, err = l.Read(message)

		if err != nil || cnt != worker.MessageLength {
			log.G(ctx).Panic("error from connection", zap.Error(err), zap.Int("message-lengtth", cnt))
		}

		fmt.Println("[RECV]", message, len(message))
	}
}
