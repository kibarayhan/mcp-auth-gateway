package gateway

import (
	"testing"

	"github.com/akibar/mcp-auth-gateway/internal/config"
)

func TestNewGateway(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{
			{Name: "test", Command: "echo", Args: []string{"hello"}},
		},
	}

	gw := New(cfg)
	if gw == nil {
		t.Fatal("New() returned nil")
	}

	if len(gw.Config.Servers) != 1 {
		t.Errorf("Servers count = %d, want 1", len(gw.Config.Servers))
	}
}

func TestRouteToolToServer(t *testing.T) {
	gw := &Gateway{
		toolToServer: map[string]string{
			"search_code":        "sourcegraph",
			"read_slack_message": "slack",
		},
	}

	server, ok := gw.RouteToolCall("search_code")
	if !ok {
		t.Fatal("RouteToolCall should find search_code")
	}
	if server != "sourcegraph" {
		t.Errorf("server = %q, want %q", server, "sourcegraph")
	}

	_, ok = gw.RouteToolCall("unknown_tool")
	if ok {
		t.Error("RouteToolCall should not find unknown_tool")
	}
}
