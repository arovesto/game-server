package world

import (
	"context"
)

// TODO move to separated package, together with encoder decoder from raw message
type MoveRequest struct {
	Direction int
}

type MoveResponce struct {
}

// TODO It is should be "Game" class based on logic
type World struct {
}

func NewWorld() *World {
	return &World{}
}

// TODO methhod to move based on request
func (w *World) Move(ctx context.Context, rq *MoveRequest) *MoveResponce {
	//log.G(ctx).Info("World: move", zap.Int("direction", rq.Direction))
	return nil
}
