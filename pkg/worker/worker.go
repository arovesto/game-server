package worker

import (
	"context"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/arovesto/game-server/internal/log"
	"github.com/arovesto/game-server/pkg/stat"
	"github.com/arovesto/game-server/pkg/world"
)

const (
	ConnDeadline  = 5 * time.Second
	MessageLength = 64
)

type Worker struct {
	stop chan struct{}

	stat *stat.Stat

	world *world.World

	curConn net.Conn
}

func NewWorker(stat *stat.Stat, w *world.World) *Worker {
	return &Worker{
		stop:  make(chan struct{}),
		stat:  stat,
		world: w,
	}
}

func (w *Worker) Stop() {
	close(w.stop)
}

func (w *Worker) Start(ctx context.Context, incomming chan net.Conn, done chan struct{}) {
	message := make([]byte, MessageLength)

	for {
		if w.curConn == nil {
			select {
			case <-w.stop:
				done <- struct{}{}
				return
			case c := <-incomming:
				w.curConn = c
			}
		} else {
			if err := w.curConn.SetReadDeadline(time.Now().Add(ConnDeadline)); err != nil {
				log.G(ctx).Debug("error on setting conn deadline. Closing connection...", zap.Error(err))
				w.curConn = nil
				_ = w.curConn.Close()
				continue
			}
			cnt, err := w.curConn.Read(message)
			if err != nil {
				log.G(ctx).Debug("failed to read from connection. Closing...", zap.Error(err))
				w.curConn = nil
				_ = w.curConn.Close()
				continue
			}

			if cnt != MessageLength {
				log.G(ctx).Debug("message length missmatch. Closing connection...")
				w.curConn = nil
				_ = w.curConn.Close()
				continue
			}

			res := w.Process(ctx, message)

			if err := w.curConn.SetWriteDeadline(time.Now().Add(ConnDeadline)); err != nil {
				log.G(ctx).Debug("error on setting conn deadline. Closing connection...", zap.Error(err))
				w.curConn = nil
				_ = w.curConn.Close()
				continue
			}

			cnt, err = w.curConn.Write(res)
			if err != nil {
				log.G(ctx).Debug("failed to write to connection. Closing...", zap.Error(err))
				w.curConn = nil
				_ = w.curConn.Close()
				continue
			}

			if cnt != MessageLength {
				log.G(ctx).Debug("message length missmatch. Closing connection...")
				w.curConn = nil
				_ = w.curConn.Close()
				continue
			}
		}
	}
}

func (w *Worker) Process(ctx context.Context, message []byte) []byte {
	if message[0] > 10 {
		message[0] = 0
		return message
	}
	w.world.Move(ctx, &world.MoveRequest{Direction: int(message[0])})
	message[0] = 'h'
	return message
}
