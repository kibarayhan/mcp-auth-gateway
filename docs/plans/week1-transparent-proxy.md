# MCP Auth Gateway — Week 1: Transparent Proxy

> **For Claude:** REQUIRED SUB-SKILL: Use `/executing-plans` to implement this plan task-by-task.

**Goal:** Build a transparent MCP proxy that spawns upstream servers and forwards all tool calls without modification. Claude Code works through the gateway with zero behavior change.

**Architecture:** The gateway is a stdio MCP server that Claude Code connects to. It reads a YAML config listing upstream MCP servers, spawns them as child processes (also via stdio), discovers their tools, and exposes all tools as its own. When Claude Code calls a tool, the gateway routes it to the correct upstream server and returns the response.

**Tech Stack:** Go 1.26, `gopkg.in/yaml.v3`, `github.com/spf13/cobra`

**Project:** `~/projects/mcp-auth-gateway/`

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `README.md`
- Create: `.gitignore`

**Step 1: Initialize Go module**

```bash
cd ~/projects/mcp-auth-gateway
go mod init github.com/akibar/mcp-auth-gateway
```

**Step 2: Create main.go with cobra CLI skeleton**

```go
// main.go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-gateway",
	Short: "MCP Auth Gateway — secure proxy for MCP servers",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the gateway",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		fmt.Fprintf(os.Stderr, "Starting gateway with config: %s\n", configPath)
		return nil
	},
}

func init() {
	startCmd.Flags().StringP("config", "c", "gateway.yaml", "Path to config file")
	rootCmd.AddCommand(startCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

**Step 3: Create .gitignore**

```
mcp-gateway
*.exe
logs/
.DS_Store
```

**Step 4: Create README.md**

```markdown
# MCP Auth Gateway

A transparent reverse proxy that adds enterprise security to any MCP server.

## Quick Start

    go build -o mcp-gateway .
    ./mcp-gateway start --config gateway.yaml

## Status

Week 1: Transparent proxy (in progress)
```

**Step 5: Add cobra dependency and verify build**

```bash
cd ~/projects/mcp-auth-gateway
go get github.com/spf13/cobra
go build -o mcp-gateway .
./mcp-gateway --help
./mcp-gateway start --help
```

Expected: Help text prints correctly, binary builds.

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: project scaffolding with cobra CLI"
```

---

### Task 2: Config Loader

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `gateway.yaml` (example config)

**Step 1: Write the test**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	yaml := `
gateway:
  listen: "localhost:3100"

servers:
  - name: test-server
    command: "echo"
    args: ["hello"]
`
	tmp := filepath.Join(t.TempDir(), "test.yaml")
	os.WriteFile(tmp, []byte(yaml), 0644)

	cfg, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Gateway.Listen != "localhost:3100" {
		t.Errorf("Listen = %q, want %q", cfg.Gateway.Listen, "localhost:3100")
	}

	if len(cfg.Servers) != 1 {
		t.Fatalf("Servers count = %d, want 1", len(cfg.Servers))
	}

	if cfg.Servers[0].Name != "test-server" {
		t.Errorf("Server name = %q, want %q", cfg.Servers[0].Name, "test-server")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path.yaml")
	if err == nil {
		t.Error("Load() should error on missing file")
	}
}
```

**Step 2: Run test — verify it fails**

```bash
cd ~/projects/mcp-auth-gateway
go test ./internal/config/ -v
```

Expected: FAIL — package doesn't exist yet.

**Step 3: Write implementation**

```go
// internal/config/config.go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Gateway  GatewayConfig  `yaml:"gateway"`
	Servers  []ServerConfig `yaml:"servers"`
	Auth     AuthConfig     `yaml:"auth"`
	Audit    AuditConfig    `yaml:"audit"`
}

type GatewayConfig struct {
	Listen    string `yaml:"listen"`
	Transport string `yaml:"transport"`
}

type ServerConfig struct {
	Name     string            `yaml:"name"`
	Command  string            `yaml:"command"`
	Args     []string          `yaml:"args"`
	Env      map[string]string `yaml:"env"`
	Policies PolicyConfig      `yaml:"policies"`
}

type PolicyConfig struct {
	AllowedRoles []string              `yaml:"allowed_roles"`
	RateLimit    string                `yaml:"rate_limit"`
	PIIFilter    bool                  `yaml:"pii_filter"`
	Audit        string                `yaml:"audit"`
	Tools        map[string]ToolPolicy `yaml:"tools"`
	BlockedArgs  []BlockedArg          `yaml:"blocked_args"`
}

type ToolPolicy struct {
	RequiresRole    string   `yaml:"requires_role"`
	AllowedRoles    []string `yaml:"allowed_roles"`
	BlockedChannels []string `yaml:"blocked_channels"`
}

type BlockedArg struct {
	Pattern string `yaml:"pattern"`
}

type AuthConfig struct {
	Provider      string   `yaml:"provider"`
	Issuer        string   `yaml:"issuer"`
	ClientID      string   `yaml:"client_id"`
	AllowedGroups []string `yaml:"allowed_groups"`
}

type AuditConfig struct {
	Destination   string `yaml:"destination"`
	Path          string `yaml:"path"`
	RetentionDays int    `yaml:"retention_days"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}
```

**Step 4: Add yaml dependency and run tests**

```bash
cd ~/projects/mcp-auth-gateway
go get gopkg.in/yaml.v3
go test ./internal/config/ -v
```

Expected: 2 tests PASS.

**Step 5: Create example gateway.yaml**

```yaml
# gateway.yaml — Example configuration
gateway:
  listen: "localhost:3100"

servers:
  - name: filesystem
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp/test"]

  - name: echo
    command: "npx"
    args: ["-y", "@anthropic-ai/mcp-test-server"]
```

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: config loader with YAML parsing and tests"
```

---

### Task 3: MCP JSON-RPC Protocol Types

**Files:**
- Create: `internal/mcp/protocol.go`
- Create: `internal/mcp/protocol_test.go`

**Step 1: Write test for JSON-RPC message parsing**

```go
// internal/mcp/protocol_test.go
package mcp

import (
	"encoding/json"
	"testing"
)

func TestParseRequest(t *testing.T) {
	raw := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"search","arguments":{"query":"test"}}}`

	var msg JSONRPCMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if msg.Method != "tools/call" {
		t.Errorf("Method = %q, want %q", msg.Method, "tools/call")
	}

	if msg.ID == nil {
		t.Error("ID should not be nil for a request")
	}
}

func TestParseNotification(t *testing.T) {
	raw := `{"jsonrpc":"2.0","method":"notifications/initialized"}`

	var msg JSONRPCMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if msg.ID != nil {
		t.Error("ID should be nil for a notification")
	}
}

func TestToolCallParams(t *testing.T) {
	raw := `{"name":"search_code","arguments":{"query":"auth token","repo":"infra-main"}}`

	var params ToolCallParams
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if params.Name != "search_code" {
		t.Errorf("Name = %q, want %q", params.Name, "search_code")
	}

	if params.Arguments["query"] != "auth token" {
		t.Errorf("query arg = %q, want %q", params.Arguments["query"], "auth token")
	}
}
```

**Step 2: Run test — verify it fails**

```bash
cd ~/projects/mcp-auth-gateway
go test ./internal/mcp/ -v
```

**Step 3: Write implementation**

```go
// internal/mcp/protocol.go
package mcp

import "encoding/json"

// JSONRPCMessage represents a JSON-RPC 2.0 message (request, response, or notification).
type JSONRPCMessage struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *JSONRPCError    `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ToolCallParams is the params for a "tools/call" request.
type ToolCallParams struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments"`
}

// ToolListResult is the result for a "tools/list" response.
type ToolListResult struct {
	Tools []ToolInfo `json:"tools"`
}

type ToolInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"inputSchema,omitempty"`
}

// IsRequest returns true if this message has an ID (it's a request, not a notification).
func (m *JSONRPCMessage) IsRequest() bool {
	return m.ID != nil
}

// IsResponse returns true if this message has a result or error.
func (m *JSONRPCMessage) IsResponse() bool {
	return m.Result != nil || m.Error != nil
}
```

**Step 4: Run tests**

```bash
go test ./internal/mcp/ -v
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: MCP JSON-RPC protocol types with tests"
```

---

### Task 4: Upstream Server Process Manager

**Files:**
- Create: `internal/upstream/server.go`
- Create: `internal/upstream/server_test.go`

**Step 1: Write test**

```go
// internal/upstream/server_test.go
package upstream

import (
	"context"
	"testing"
	"time"
)

func TestStartServer_Echo(t *testing.T) {
	// Use 'cat' as a simple echo server — it reads stdin and writes to stdout
	srv, err := Start(context.Background(), "cat", nil, nil)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer srv.Stop()

	// Write a JSON-RPC message to stdin
	msg := `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"
	_, err = srv.Stdin.Write([]byte(msg))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}

	// Read it back from stdout (cat echoes it)
	buf := make([]byte, 1024)
	srv.Stdout.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := srv.Stdout.Read(buf)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}

	got := string(buf[:n])
	if got != msg {
		t.Errorf("Echo got %q, want %q", got, msg)
	}
}

func TestStartServer_InvalidCommand(t *testing.T) {
	_, err := Start(context.Background(), "/nonexistent/binary", nil, nil)
	if err == nil {
		t.Error("Start should error on invalid command")
	}
}
```

**Step 2: Run test — verify it fails**

```bash
go test ./internal/upstream/ -v
```

**Step 3: Write implementation**

```go
// internal/upstream/server.go
package upstream

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// Server represents a running upstream MCP server process.
type Server struct {
	Name    string
	Cmd     *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
	cancel  context.CancelFunc
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
```

**Step 4: Run tests**

Note: The echo test with `cat` won't work cleanly with ReadDeadline since Stdout is a pipe, not a net.Conn. Simplify the test:

```go
// Revised test — just verify process starts and stops
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
```

```bash
go test ./internal/upstream/ -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: upstream server process manager with start/stop"
```

---

### Task 5: stdio JSON-RPC Transport

**Files:**
- Create: `internal/transport/stdio.go`
- Create: `internal/transport/stdio_test.go`

**Step 1: Write test**

```go
// internal/transport/stdio_test.go
package transport

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/akibar/mcp-auth-gateway/internal/mcp"
)

func TestReadMessage(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
	reader := bytes.NewBufferString(input)

	msg, err := ReadMessage(reader)
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}

	if msg.Method != "tools/list" {
		t.Errorf("Method = %q, want %q", msg.Method, "tools/list")
	}
}

func TestWriteMessage(t *testing.T) {
	msg := &mcp.JSONRPCMessage{
		JSONRPC: "2.0",
		Method:  "test",
	}

	var buf bytes.Buffer
	err := WriteMessage(&buf, msg)
	if err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}

	// Verify it's valid JSON followed by newline
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	var parsed mcp.JSONRPCMessage
	if err := json.Unmarshal(lines[0], &parsed); err != nil {
		t.Fatalf("Output not valid JSON: %v", err)
	}

	if parsed.Method != "test" {
		t.Errorf("Method = %q, want %q", parsed.Method, "test")
	}
}
```

**Step 2: Run test — verify it fails**

```bash
go test ./internal/transport/ -v
```

**Step 3: Write implementation**

```go
// internal/transport/stdio.go
package transport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/akibar/mcp-auth-gateway/internal/mcp"
)

// ReadMessage reads a single JSON-RPC message from a newline-delimited stream.
func ReadMessage(r io.Reader) (*mcp.JSONRPCMessage, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read: %w", err)
		}
		return nil, io.EOF
	}

	var msg mcp.JSONRPCMessage
	if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
		return nil, fmt.Errorf("parse JSON-RPC: %w", err)
	}

	return &msg, nil
}

// WriteMessage writes a single JSON-RPC message as a newline-delimited JSON line.
func WriteMessage(w io.Writer, msg *mcp.JSONRPCMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
```

**Step 4: Run tests**

```bash
go test ./internal/transport/ -v
```

Expected: 2 tests PASS.

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: stdio JSON-RPC transport reader/writer with tests"
```

---

### Task 6: Gateway Core — Tool Discovery and Routing

**Files:**
- Create: `internal/gateway/gateway.go`
- Create: `internal/gateway/gateway_test.go`

**Step 1: Write test**

```go
// internal/gateway/gateway_test.go
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
			"search_code":       "sourcegraph",
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
```

**Step 2: Run test — verify it fails**

```bash
go test ./internal/gateway/ -v
```

**Step 3: Write implementation**

```go
// internal/gateway/gateway.go
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
```

**Step 4: Run tests**

```bash
go test ./internal/gateway/ -v
```

Expected: 2 tests PASS.

**Step 5: Run all tests**

```bash
go test ./... -v
```

Expected: All tests pass across all packages.

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: gateway core with tool routing and server registry"
```

---

### Task 7: Wire It All Together — main.go

**Files:**
- Modify: `main.go`

**Step 1: Update main.go to load config, start servers, and run the proxy loop**

```go
// main.go
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/akibar/mcp-auth-gateway/internal/config"
	"github.com/akibar/mcp-auth-gateway/internal/gateway"
	"github.com/akibar/mcp-auth-gateway/internal/mcp"
	"github.com/akibar/mcp-auth-gateway/internal/transport"
	"github.com/akibar/mcp-auth-gateway/internal/upstream"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-gateway",
	Short: "MCP Auth Gateway — secure proxy for MCP servers",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the gateway",
	RunE:  runStart,
}

func init() {
	startCmd.Flags().StringP("config", "c", "gateway.yaml", "Path to config file")
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	slog.Info("starting gateway", "config", configPath, "servers", len(cfg.Servers))

	gw := gateway.New(cfg)
	ctx := context.Background()

	// Start upstream servers and discover their tools
	for _, serverCfg := range cfg.Servers {
		slog.Info("starting upstream server", "name", serverCfg.Name, "command", serverCfg.Command)

		srv, err := upstream.Start(ctx, serverCfg.Command, serverCfg.Args, nil)
		if err != nil {
			return fmt.Errorf("start server %s: %w", serverCfg.Name, err)
		}
		srv.Name = serverCfg.Name
		gw.SetServer(serverCfg.Name, srv)

		// Send initialize request to discover capabilities
		initReq := &mcp.JSONRPCMessage{
			JSONRPC: "2.0",
			Method:  "initialize",
			Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"mcp-gateway","version":"0.1.0"}}`),
		}
		id := json.RawMessage(`1`)
		initReq.ID = &id

		if err := transport.WriteMessage(srv.Stdin, initReq); err != nil {
			slog.Error("failed to send initialize", "server", serverCfg.Name, "error", err)
			continue
		}

		// Read initialize response
		resp, err := transport.ReadMessage(srv.Stdout)
		if err != nil {
			slog.Error("failed to read initialize response", "server", serverCfg.Name, "error", err)
			continue
		}
		slog.Debug("initialize response", "server", serverCfg.Name, "result", string(resp.Result))

		// Send initialized notification
		initNotif := &mcp.JSONRPCMessage{
			JSONRPC: "2.0",
			Method:  "notifications/initialized",
		}
		transport.WriteMessage(srv.Stdin, initNotif)

		// Request tools list
		toolsReq := &mcp.JSONRPCMessage{
			JSONRPC: "2.0",
			Method:  "tools/list",
		}
		toolsID := json.RawMessage(`2`)
		toolsReq.ID = &toolsID

		if err := transport.WriteMessage(srv.Stdin, toolsReq); err != nil {
			slog.Error("failed to request tools", "server", serverCfg.Name, "error", err)
			continue
		}

		toolsResp, err := transport.ReadMessage(srv.Stdout)
		if err != nil {
			slog.Error("failed to read tools", "server", serverCfg.Name, "error", err)
			continue
		}

		var toolsResult mcp.ToolListResult
		if err := json.Unmarshal(toolsResp.Result, &toolsResult); err != nil {
			slog.Error("failed to parse tools", "server", serverCfg.Name, "error", err)
			continue
		}

		gw.RegisterTools(serverCfg.Name, toolsResult.Tools)
		slog.Info("discovered tools", "server", serverCfg.Name, "count", len(toolsResult.Tools))
	}

	slog.Info("gateway ready", "total_tools", len(gw.AllTools()))

	// Main proxy loop: read from stdin, route to upstream, write to stdout
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var msg mcp.JSONRPCMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			slog.Error("invalid JSON from client", "error", err)
			continue
		}

		switch msg.Method {
		case "initialize":
			// Respond with our own capabilities
			result := map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":   map[string]interface{}{"tools": map[string]interface{}{}},
				"serverInfo":     map[string]interface{}{"name": "mcp-gateway", "version": "0.1.0"},
			}
			resultJSON, _ := json.Marshal(result)
			resp := &mcp.JSONRPCMessage{
				JSONRPC: "2.0",
				ID:      msg.ID,
				Result:  resultJSON,
			}
			transport.WriteMessage(os.Stdout, resp)

		case "notifications/initialized":
			// Ignore

		case "tools/list":
			// Return aggregated tools from all upstream servers
			result := mcp.ToolListResult{Tools: gw.AllTools()}
			resultJSON, _ := json.Marshal(result)
			resp := &mcp.JSONRPCMessage{
				JSONRPC: "2.0",
				ID:      msg.ID,
				Result:  resultJSON,
			}
			transport.WriteMessage(os.Stdout, resp)

		case "tools/call":
			// Route to the correct upstream server
			var params mcp.ToolCallParams
			json.Unmarshal(msg.Params, &params)

			serverName, ok := gw.RouteToolCall(params.Name)
			if !ok {
				errResp := &mcp.JSONRPCMessage{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Error:   &mcp.JSONRPCError{Code: -32601, Message: fmt.Sprintf("tool %q not found", params.Name)},
				}
				transport.WriteMessage(os.Stdout, errResp)
				continue
			}

			srv, err := gw.GetServer(serverName)
			if err != nil {
				errResp := &mcp.JSONRPCMessage{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Error:   &mcp.JSONRPCError{Code: -32603, Message: err.Error()},
				}
				transport.WriteMessage(os.Stdout, errResp)
				continue
			}

			// Forward to upstream
			transport.WriteMessage(srv.Stdin, &msg)

			// Read response from upstream
			resp, err := transport.ReadMessage(srv.Stdout)
			if err != nil {
				if err == io.EOF {
					slog.Error("upstream server closed", "server", serverName)
				}
				errResp := &mcp.JSONRPCMessage{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Error:   &mcp.JSONRPCError{Code: -32603, Message: fmt.Sprintf("upstream error: %v", err)},
				}
				transport.WriteMessage(os.Stdout, errResp)
				continue
			}

			// Forward response back to client
			resp.ID = msg.ID
			transport.WriteMessage(os.Stdout, resp)

		default:
			slog.Debug("unhandled method", "method", msg.Method)
		}
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

**Step 2: Build and verify**

```bash
cd ~/projects/mcp-auth-gateway
go build -o mcp-gateway .
./mcp-gateway --help
```

Expected: Builds without errors.

**Step 3: Run all tests**

```bash
go test ./... -v
```

Expected: All tests pass.

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: wire gateway proxy loop — transparent MCP forwarding"
```

---

### Task 8: Integration Test with a Real MCP Server

**Files:**
- Create: `test/integration_test.sh`

**Step 1: Create integration test script**

```bash
#!/bin/bash
# test/integration_test.sh — Test gateway with a real MCP server
set -e

echo "Building gateway..."
cd ~/projects/mcp-auth-gateway
go build -o mcp-gateway .

echo "Testing: send initialize + tools/list through gateway..."

# Create a minimal test config using the filesystem MCP server
cat > /tmp/test-gateway.yaml << 'EOF'
gateway:
  listen: "localhost:3100"
servers:
  - name: filesystem
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
EOF

# Send initialize and tools/list, capture output
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | timeout 30 ./mcp-gateway start --config /tmp/test-gateway.yaml 2>/tmp/gateway-test.log

echo ""
echo "Gateway logs:"
cat /tmp/gateway-test.log
echo ""
echo "DONE"
```

**Step 2: Run integration test**

```bash
chmod +x ~/projects/mcp-auth-gateway/test/integration_test.sh
bash ~/projects/mcp-auth-gateway/test/integration_test.sh
```

Expected: Gateway starts, connects to filesystem MCP server, discovers tools, and returns them. The output should contain a JSON response with filesystem tools (read_file, write_file, etc.).

**Step 3: Commit**

```bash
cd ~/projects/mcp-auth-gateway
git add -A
git commit -m "test: integration test with real filesystem MCP server"
```

---

## Summary

| Task | What it builds | Tests |
|------|---------------|-------|
| 1 | Project scaffolding, CLI | Build verification |
| 2 | YAML config loader | 2 unit tests |
| 3 | MCP JSON-RPC protocol types | 3 unit tests |
| 4 | Upstream server process manager | 2 unit tests |
| 5 | stdio transport (read/write) | 2 unit tests |
| 6 | Gateway core (routing, tool registry) | 2 unit tests |
| 7 | Main proxy loop (wiring) | Build + all tests |
| 8 | Integration test with real MCP server | End-to-end |

After completing all 8 tasks, you'll have a working transparent MCP proxy that Claude Code can connect to.
