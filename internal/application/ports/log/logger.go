package log

import "context"

type Logger interface {
	Errorf(ctx context.Context, format string, args ...any)
	Warnf(ctx context.Context, format string, args ...any)
	Infof(ctx context.Context, format string, args ...any)
}
