package exec

import (
	"context"
	"io"
)

// Cmd represents an executable command, abstracting os/exec.Cmd.
type Cmd interface {
	SetStdout(io.Writer)
	SetStderr(io.Writer)
	Run() error
	Start() error
	Wait() error
	StderrPipe() (io.ReadCloser, error)
}

// Commander defines a port for creating executable commands.
// This allows us to mock exec.CommandContext in tests.
type Commander interface {
	CommandContext(ctx context.Context, name string, args ...string) Cmd
}
