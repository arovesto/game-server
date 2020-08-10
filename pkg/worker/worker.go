package worker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/arovesto/game-server/pkg/stat"
	"github.com/arovesto/game-server/pkg/world"
)

const (
	ConnDeadline  = 5 * time.Second
	// TODO this should be taked carefull
	MessageLength = 1024
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
			select {
			case <-w.stop:
				done <- struct{}{}
				return
			default:
				w.ProcessConnection(ctx)
			}
		}
	}
}

// TODO support short read message with long write message
func (w *Worker) ProcessConnection(ctx context.Context) {
	message := make([]byte, MessageLength)

	var err error
	defer func() {
		if err != nil {
			var e net.Error
			if ok := errors.As(err, &e); ok && e.Timeout() {
				return
			}
			_ = w.curConn.Close()
			w.curConn = nil
		}
	}()

	if err = w.curConn.SetReadDeadline(time.Now().Add(ConnDeadline)); err != nil {
		err = fmt.Errorf("error on setting conn deadline. Closing connection...: %w", err)
		return
	}
	cnt, err := w.curConn.Read(message)
	if err != nil {
		err = fmt.Errorf("failed to read from connection: %w", err)
		return
	}

	if cnt != MessageLength {
		err = fmt.Errorf("message length missmatch: %v not %v", cnt, MessageLength)
		return
	}

	res := w.Process(ctx, message)

	if err = w.curConn.SetWriteDeadline(time.Now().Add(ConnDeadline)); err != nil {
		err = fmt.Errorf("error on setting write conn deadline: %w", err)
		return
	}

	cnt, err = w.curConn.Write(res)
	if err != nil {
		err = fmt.Errorf("failed to write from connection: %w", err)
		return
	}

	if cnt != MessageLength {
		err = fmt.Errorf("message length missmatch: %v not %v", cnt, MessageLength)
		return
	}
}

func (w *Worker) Process(ctx context.Context, message []byte) []byte {
	if message[0] > '9' {
		message[0] = '0'
		return message
	}
	w.world.Move(ctx, &world.MoveRequest{Direction: int(message[0])})
	message[0] = 'h'
	return message
}
