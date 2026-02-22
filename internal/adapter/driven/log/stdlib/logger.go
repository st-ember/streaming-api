package stdlib

import (
	"log"

	logPort "github.com/st-ember/streaming-api/internal/application/ports/log"
)

type StdLogger struct{}

func NewStdLogger() logPort.Logger { return &StdLogger{} }

func (l *StdLogger) Errorf(format string, args ...any) { log.Printf("ERROR: "+format, args...) }

func (l *StdLogger) Warnf(format string, args ...any) { log.Printf("WARN: "+format, args...) }

func (l *StdLogger) Infof(format string, args ...any) { log.Printf("INFO: "+format, args...) }
