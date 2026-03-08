package upstream

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// Server represents a running upstream MCP server process.
type Server struct {
	Name   string
	Cmd    *exec.Cmd
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	cancel context.CancelFunc
}

// Start spawns an upstream MCP server as a child process.
func Start(ctx context.Context, command string, args []string, env []string) (*Server, error) {
	ctx, cancel := context.WithCancel(ctx)

	cmd := exec.CommandContext(ctx, command, args...)
	if len(env) > 0 {
		cmd.Env = env
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start %s: %w", command, err)
	}

	return &Server{
		Cmd:    cmd,
		Stdin:  stdin,
		Stdout: stdout,
		cancel: cancel,
	}, nil
}

// Stop terminates the server process.
func (s *Server) Stop() error {
	s.cancel()
	return s.Cmd.Wait()
}
