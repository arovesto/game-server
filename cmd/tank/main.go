package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

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

	t := time.Now()

	res := make(chan struct{}, 10)

	for j := 0; j < 10; j++ {
		go func() {
			l, err := net.Dial("tcp", cfg.General.Address)
			if err != nil {
				log.G(ctx).Panic("failed to call address", zap.Error(err))
			}

			for i := 0; i < 10000; i++ {
				message := []byte("1111111111111111111111111111111111111111111111111111111111111111")

				cnt, err := l.Write(message)
				if err != nil || cnt != worker.MessageLength {
					log.G(ctx).Panic("error from connection", zap.Error(err), zap.Int("message-lengtth", cnt))
				}
				cnt, err = l.Read(message)

				if err != nil || cnt != worker.MessageLength {
					log.G(ctx).Panic("error from connection", zap.Error(err), zap.Int("message-lengtth", cnt))
				}
			}
			res <- struct{}{}
		}()
	}
	for j := 0; j < 10; j++ {
		<-res
	}
	fmt.Println(time.Now().Sub(t))
}
