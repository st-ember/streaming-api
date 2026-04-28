package log

import (
	"context"
)

type Logger interface {
	Errorf(ctx context.Context, category LogCategory, sourceID string, format string, args ...any)
	Warnf(ctx context.Context, category LogCategory, sourceID string, format string, args ...any)
	Infof(ctx context.Context, category LogCategory, sourceID string, format string, args ...any)
}
