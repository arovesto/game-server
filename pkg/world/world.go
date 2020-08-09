package world

import (
	"context"
)

type MoveRequest struct {
	Direction int
}

type MoveResponce struct {
}

type World struct {
}

func NewWorld() *World {
	return &World{}
}

func (w *World) Move(ctx context.Context, rq *MoveRequest) *MoveResponce {
	//log.G(ctx).Info("World: move", zap.Int("direction", rq.Direction))
	return nil
}
