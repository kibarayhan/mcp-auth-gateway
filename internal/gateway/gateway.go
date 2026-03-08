package gateway

import (
	"fmt"
	"log/slog"

	"github.com/akibar/mcp-auth-gateway/internal/config"
	"github.com/akibar/mcp-auth-gateway/internal/mcp"
	"github.com/akibar/mcp-auth-gateway/internal/upstream"
)

// Gateway is the core MCP proxy that routes tool calls to upstream servers.
type Gateway struct {
	Config       *config.Config
	servers      map[string]*upstream.Server
	toolToServer map[string]string
	allTools     []mcp.ToolInfo
}

// New creates a new gateway from config.
func New(cfg *config.Config) *Gateway {
	return &Gateway{
		Config:       cfg,
		servers:      make(map[string]*upstream.Server),
		toolToServer: make(map[string]string),
	}
}

// RegisterTools adds tool-to-server mappings. Called after discovering tools from upstream.
func (g *Gateway) RegisterTools(serverName string, tools []mcp.ToolInfo) {
	for _, tool := range tools {
		g.toolToServer[tool.Name] = serverName
		g.allTools = append(g.allTools, tool)
		slog.Debug("registered tool", "tool", tool.Name, "server", serverName)
	}
}

// RouteToolCall returns the server name that handles the given tool.
func (g *Gateway) RouteToolCall(toolName string) (string, bool) {
	server, ok := g.toolToServer[toolName]
	return server, ok
}

// AllTools returns all tools across all upstream servers.
func (g *Gateway) AllTools() []mcp.ToolInfo {
	return g.allTools
}

// SetServer stores a running upstream server reference.
func (g *Gateway) SetServer(name string, srv *upstream.Server) {
	g.servers[name] = srv
}

// GetServer returns a running upstream server by name.
func (g *Gateway) GetServer(name string) (*upstream.Server, error) {
	srv, ok := g.servers[name]
	if !ok {
		return nil, fmt.Errorf("server %q not found", name)
	}
	return srv, nil
}
