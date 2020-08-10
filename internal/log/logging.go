package log

import (
	"context"

	"go.uber.org/zap"

	"github.com/arovesto/game-server/internal/config"
)

type logKey struct{}

// TODO support some fast logging type
func SetupLogger(cfg *config.LoggingConfig) (*zap.Logger, error) {
	switch cfg.Type {
	case config.LogDebug:
		return zap.NewDevelopment()
	case config.LogProd:
		return zap.NewProduction()
	default:
		panic("Failed to create logger")
	}
}

func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, logKey{}, l)
}

func G(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(logKey{}).(*zap.Logger)
	if !ok {
		return zap.NewNop()
	}
	return logger
}
