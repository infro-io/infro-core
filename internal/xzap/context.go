package xzap

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

// NewContext returns a new Context that carries the log *zap.Logger.
func NewContext(ctx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, log)
}

// FromContext returns the *zap.Logger stored in ctx if any, a no-op one otherwise.
func FromContext(ctx context.Context) *zap.Logger {
	if log, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return log
	}
	return zap.NewNop()
}
