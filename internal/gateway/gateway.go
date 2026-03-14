package gateway

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/akibar/mcp-auth-gateway/internal/audit"
	"github.com/akibar/mcp-auth-gateway/internal/auth"
	"github.com/akibar/mcp-auth-gateway/internal/config"
	"github.com/akibar/mcp-auth-gateway/internal/mcp"
	"github.com/akibar/mcp-auth-gateway/internal/pii"
	"github.com/akibar/mcp-auth-gateway/internal/policy"
	"github.com/akibar/mcp-auth-gateway/internal/ratelimit"
	"github.com/akibar/mcp-auth-gateway/internal/upstream"
)

// Gateway is the core MCP proxy that routes tool calls to upstream servers.
type Gateway struct {
	Config      *config.Config
	Policy      *policy.Engine
	User        *auth.User
	Audit       *audit.Logger
	RateLimiter *ratelimit.Limiter
	PIIFilter   *pii.Filter

	mu           sync.RWMutex
	servers      map[string]*upstream.Server
	toolToServer map[string]string
	allTools     []mcp.ToolInfo

	// Ready is closed when all upstream servers have been initialized.
	Ready chan struct{}
}

// New creates a new gateway from config.
func New(cfg *config.Config) *Gateway {
	return &Gateway{
		Config:       cfg,
		Policy:       policy.New(),
		RateLimiter:  ratelimit.New(),
		servers:      make(map[string]*upstream.Server),
		toolToServer: make(map[string]string),
		Ready:        make(chan struct{}),
	}
}

// RegisterTools adds tool-to-server mappings. Safe for concurrent use.
func (g *Gateway) RegisterTools(serverName string, tools []mcp.ToolInfo) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, tool := range tools {
		g.toolToServer[tool.Name] = serverName
		g.allTools = append(g.allTools, tool)
		slog.Debug("registered tool", "tool", tool.Name, "server", serverName)
	}
}

// RouteToolCall returns the server name that handles the given tool.
func (g *Gateway) RouteToolCall(toolName string) (string, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	server, ok := g.toolToServer[toolName]
	return server, ok
}

// AllTools returns all tools across all upstream servers.
func (g *Gateway) AllTools() []mcp.ToolInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]mcp.ToolInfo, len(g.allTools))
	copy(out, g.allTools)
	return out
}

// SetServer stores a running upstream server reference. Safe for concurrent use.
func (g *Gateway) SetServer(name string, srv *upstream.Server) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.servers[name] = srv
}

// GetServer returns a running upstream server by name.
func (g *Gateway) GetServer(name string) (*upstream.Server, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	srv, ok := g.servers[name]
	if !ok {
		return nil, fmt.Errorf("server %q not found", name)
	}
	return srv, nil
}

// ServerConfigByName returns the config for a named server.
func (g *Gateway) ServerConfigByName(name string) (config.ServerConfig, bool) {
	for _, s := range g.Config.Servers {
		if s.Name == name {
			return s, true
		}
	}
	return config.ServerConfig{}, false
}
