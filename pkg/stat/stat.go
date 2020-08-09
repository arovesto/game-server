package stat

import (
	"context"
	"github.com/arovesto/game-server/internal/log"
	"go.uber.org/zap"
	"sync"
)

const (
	CurConnStat = iota
	ErrorsStat
	StatLen
)

type Stat struct {
	mtx  sync.RWMutex
	stat []int
}

func NewStat() *Stat {
	return &Stat{
		mtx:  sync.RWMutex{},
		stat: make([]int, StatLen),
	}
}

func (s *Stat) Inc(ctx context.Context, stat int) {
	if stat >= StatLen {
		log.G(ctx).Error("stat out of enum", zap.Int("stat", stat))
		return
	}
	s.mtx.Lock()
	s.stat[stat]++
	s.mtx.Unlock()
}

func (s *Stat) Dec(ctx context.Context, stat int) {
	if stat >= StatLen {
		log.G(ctx).Error("stat out of enum", zap.Int("stat", stat))
		return
	}

	s.mtx.Lock()
	s.stat[stat]--
	s.mtx.Unlock()
}

func (s *Stat) GetStat(ctx context.Context, stat int) int {
	if stat >= StatLen {
		log.G(ctx).Error("stat out of enum", zap.Int("stat", stat))
		return -1
	}
	return s.stat[stat]
}
