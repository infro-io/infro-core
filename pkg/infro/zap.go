package infro

import (
	"context"

	"go.uber.org/zap"

	"github.com/infro-io/infro-core/internal/xzap"
)

// NewZapContext sets the logger that Infro will use.
func NewZapContext(ctx context.Context, log *zap.Logger) context.Context {
	return xzap.NewContext(ctx, log)
}
