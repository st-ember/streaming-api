package os

import (
	"context"
	"io"
	"os/exec"

	execPort "github.com/st-ember/streaming-api/internal/application/ports/exec"
)

type osCommander struct{}

func NewOsCommander() execPort.Commander {
	return &osCommander{}
}

func (c *osCommander) CommandContext(ctx context.Context, name string, args ...string) execPort.Cmd {
	return &osCmd{exec.CommandContext(ctx, name, args...)}
}

type osCmd struct {
	cmd *exec.Cmd
}

func (c *osCmd) SetStdout(w io.Writer)              { c.cmd.Stdout = w }
func (c *osCmd) SetStderr(w io.Writer)              { c.cmd.Stderr = w }
func (c *osCmd) Run() error                         { return c.cmd.Run() }
func (c *osCmd) Start() error                       { return c.cmd.Start() }
func (c *osCmd) Wait() error                        { return c.cmd.Wait() }
func (c *osCmd) StderrPipe() (io.ReadCloser, error) { return c.cmd.StderrPipe() }
