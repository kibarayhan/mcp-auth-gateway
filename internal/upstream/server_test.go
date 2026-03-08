package upstream

import (
	"context"
	"testing"
)

func TestStartServer_StartStop(t *testing.T) {
	srv, err := Start(context.Background(), "cat", nil, nil)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	// Close stdin to make cat exit
	srv.Stdin.Close()

	err = srv.Stop()
	if err != nil {
		t.Logf("Stop returned: %v (expected for cat)", err)
	}
}

func TestStartServer_InvalidCommand(t *testing.T) {
	_, err := Start(context.Background(), "/nonexistent/binary", nil, nil)
	if err == nil {
		t.Error("Start should error on invalid command")
	}
}
