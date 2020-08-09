package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/arovesto/game-server/internal/config"
	"github.com/arovesto/game-server/internal/log"
	"github.com/arovesto/game-server/pkg/stat"
	"github.com/arovesto/game-server/pkg/worker"
	"github.com/arovesto/game-server/pkg/world"
)

const (
	ShutdownDuration = 10 * time.Second
)

type statistics struct {
	curConn   int
	curErrors int
}

type server struct {
	workers []*worker.Worker

	allDone chan struct{}
	tasks   chan net.Conn

	stat *stat.Stat

	world *world.World
}

func NewServer(cfg *config.Config) (*server, error) {
	st := stat.NewStat()
	w := world.NewWorld()

	workers := make([]*worker.Worker, cfg.General.ClientsCap)
	for i := range workers {
		workers[i] = worker.NewWorker(st, w)
	}

	return &server{
		workers: workers,
		allDone: make(chan struct{}, len(workers)),
		tasks:   make(chan net.Conn, len(workers)),
		stat:    st,
		world:   w,
	}, nil
}

func (s *server) Start(ctx context.Context, l net.Listener) error {
	log.G(ctx).Info("creating workers...")
	for _, w := range s.workers {
		go w.Start(ctx, s.tasks, s.allDone)
	}

	defer s.GraceStop(ctx)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.G(ctx).Error("failed to obtain a connection", zap.Error(err))
			s.stat.Inc(ctx, stat.ErrorsStat)
		}

		log.G(ctx).Info("new connection")
		s.stat.Inc(ctx, stat.CurConnStat)
		if len(s.tasks) == len(s.workers) {
			log.G(ctx).Info("too much connections. Dumping incoming...", zap.String("address", conn.LocalAddr().String()))
			_ = conn.Close()
			s.stat.Dec(ctx, stat.CurConnStat)
		}
		s.tasks <- conn
	}
}

func (s *server) StartExec(ctx context.Context, scanner *bufio.Scanner) error {
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.G(ctx).Error("error occured on scanner on command reading", zap.Error(err))
			return err
		}
		fmt.Println("[Server]", scanner.Text())
		if scanner.Text() == "shutdown" {
			s.GraceStop(ctx)
			return nil
		}
	}
	return nil
}

func (s *server) GraceStop(ctx context.Context) {
	log.G(ctx).Info("shutting down...")

	for _, w := range s.workers {
		w.Stop()
	}

	timeout := time.NewTimer(ShutdownDuration)

	log.G(ctx).Info("Waiting for workers to stop...")
	for range s.workers {
		select {
		case <-s.allDone:
		case <-timeout.C:
			log.G(ctx).Debug("failed to shutdown gracefully, exiting...")
			return
		}
	}
	log.G(ctx).Info("All workers are down")
}
