package stdlib

import (
	"context"
	"log"

	logPort "github.com/st-ember/streaming-api/internal/application/ports/log"
)

type StdLogger struct{}

func NewStdLogger() logPort.Logger { return &StdLogger{} }

func (l *StdLogger) Errorf(ctx context.Context, format string, args ...any) {
	log.Printf("ERROR: "+format, args...)
}

func (l *StdLogger) Warnf(ctx context.Context, format string, args ...any) {
	log.Printf("WARN: "+format, args...)
}

func (l *StdLogger) Infof(ctx context.Context, format string, args ...any) {
	log.Printf("INFO: "+format, args...)
}
