# MCP Auth Gateway — Week 2: Auth + Policy Engine

> **For Claude:** REQUIRED SUB-SKILL: Use `/executing-plans` to implement this plan task-by-task.

**Goal:** Add API key authentication and a three-level policy engine so unauthorized tool calls are blocked and role-based access control works.

**Architecture:** The gateway reads API key users from `gateway.yaml`. On startup, it resolves the caller's identity from the `MCP_AUTH_TOKEN` environment variable. Every `tools/call` passes through a policy engine that checks three levels: (1) server-level — does the user's role allow access to this server? (2) tool-level — does the user's role allow this specific tool? (3) argument-level — do any arguments match blocked patterns? Denied calls return a JSON-RPC error without touching the upstream server.

**Tech Stack:** Go 1.26, `regexp` stdlib for argument pattern matching, existing `config` and `gateway` packages.

**Project:** `~/projects/mcp-auth-gateway/`

**Existing codebase (Week 1):**
- `main.go` — Cobra CLI + proxy loop (initialize → tools/list → tools/call forwarding)
- `internal/config/config.go` — YAML loader with `AuthConfig`, `PolicyConfig`, `ToolPolicy`, `BlockedArg` types (already defined but unused)
- `internal/gateway/gateway.go` — Gateway core with tool-to-server routing
- `internal/mcp/protocol.go` — JSON-RPC types, `ToolCallParams` with `Arguments map[string]string`
- `internal/transport/stdio.go` — Newline-delimited JSON reader/writer
- `internal/upstream/server.go` — Child process manager

---

### Task 1: Auth Identity Types

**Files:**
- Create: `internal/auth/auth.go`
- Create: `internal/auth/auth_test.go`

**Step 1: Write the test**

```go
// internal/auth/auth_test.go
package auth

import "testing"

func TestUserHasRole(t *testing.T) {
	u := &User{Name: "akibar", Roles: []string{"engineer", "oncall"}}

	if !u.HasRole("engineer") {
		t.Error("should have role engineer")
	}
	if u.HasRole("admin") {
		t.Error("should not have role admin")
	}
}

func TestUserHasAnyRole(t *testing.T) {
	u := &User{Name: "akibar", Roles: []string{"engineer"}}

	if !u.HasAnyRole([]string{"admin", "engineer"}) {
		t.Error("should match when any role overlaps")
	}
	if u.HasAnyRole([]string{"admin", "superadmin"}) {
		t.Error("should not match when no role overlaps")
	}
}

func TestAnonymousUser(t *testing.T) {
	u := Anonymous()
	if u.Name != "anonymous" {
		t.Errorf("Name = %q, want %q", u.Name, "anonymous")
	}
	if u.HasRole("engineer") {
		t.Error("anonymous should have no roles")
	}
	if u.Authenticated {
		t.Error("anonymous should not be authenticated")
	}
}
```

**Step 2: Run test — verify it fails**

```bash
cd ~/projects/mcp-auth-gateway
go test ./internal/auth/ -v
```

Expected: FAIL — package doesn't exist yet.

**Step 3: Write implementation**

```go
// internal/auth/auth.go
package auth

// User represents an authenticated caller.
type User struct {
	Name          string
	Roles         []string
	Authenticated bool
}

// HasRole returns true if the user has the given role.
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole returns true if the user has at least one of the given roles.
func (u *User) HasAnyRole(roles []string) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

// Anonymous returns an unauthenticated user with no roles.
func Anonymous() *User {
	return &User{Name: "anonymous", Authenticated: false}
}
```

**Step 4: Run tests**

```bash
go test ./internal/auth/ -v
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add internal/auth/
git commit -m "feat: auth identity types with role checking"
```

---

### Task 2: API Key Authenticator

**Files:**
- Create: `internal/auth/apikey.go`
- Create: `internal/auth/apikey_test.go`
- Modify: `internal/config/config.go` (add `Users` field to `AuthConfig`)

**Step 1: Update config types**

Add `Users` to `AuthConfig` and add `UserConfig` type in `internal/config/config.go`:

```go
// Add to AuthConfig:
type AuthConfig struct {
	Provider      string       `yaml:"provider"`
	Issuer        string       `yaml:"issuer"`
	ClientID      string       `yaml:"client_id"`
	AllowedGroups []string     `yaml:"allowed_groups"`
	Users         []UserConfig `yaml:"users"`
}

type UserConfig struct {
	Key   string   `yaml:"key"`
	Name  string   `yaml:"name"`
	Roles []string `yaml:"roles"`
}
```

**Step 2: Write the API key authenticator test**

```go
// internal/auth/apikey_test.go
package auth

import (
	"testing"

	"github.com/akibar/mcp-auth-gateway/internal/config"
)

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	users := []config.UserConfig{
		{Key: "sk-eng-abc123", Name: "akibar", Roles: []string{"engineer"}},
		{Key: "sk-admin-xyz789", Name: "admin-bot", Roles: []string{"admin", "engineer"}},
	}
	a := NewAPIKeyAuth(users)

	user, err := a.Authenticate("sk-eng-abc123")
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if user.Name != "akibar" {
		t.Errorf("Name = %q, want %q", user.Name, "akibar")
	}
	if !user.HasRole("engineer") {
		t.Error("should have role engineer")
	}
	if !user.Authenticated {
		t.Error("should be authenticated")
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	users := []config.UserConfig{
		{Key: "sk-eng-abc123", Name: "akibar", Roles: []string{"engineer"}},
	}
	a := NewAPIKeyAuth(users)

	_, err := a.Authenticate("wrong-key")
	if err == nil {
		t.Error("should error on invalid key")
	}
}

func TestAPIKeyAuth_EmptyKey(t *testing.T) {
	users := []config.UserConfig{
		{Key: "sk-eng-abc123", Name: "akibar", Roles: []string{"engineer"}},
	}
	a := NewAPIKeyAuth(users)

	_, err := a.Authenticate("")
	if err == nil {
		t.Error("should error on empty key")
	}
}

func TestAPIKeyAuth_NoUsers(t *testing.T) {
	a := NewAPIKeyAuth(nil)

	user, err := a.Authenticate("")
	if err != nil {
		t.Fatalf("should not error when no users configured: %v", err)
	}
	if user.Name != "anonymous" {
		t.Errorf("Name = %q, want %q", user.Name, "anonymous")
	}
}
```

**Step 3: Run test — verify it fails**

```bash
go test ./internal/auth/ -v
```

Expected: FAIL — `NewAPIKeyAuth` not defined.

**Step 4: Write implementation**

```go
// internal/auth/apikey.go
package auth

import (
	"fmt"

	"github.com/akibar/mcp-auth-gateway/internal/config"
)

// APIKeyAuth authenticates users by API key lookup.
type APIKeyAuth struct {
	keys map[string]config.UserConfig
}

// NewAPIKeyAuth creates an authenticator from configured users.
// If no users are configured, all callers are treated as anonymous (auth disabled).
func NewAPIKeyAuth(users []config.UserConfig) *APIKeyAuth {
	keys := make(map[string]config.UserConfig, len(users))
	for _, u := range users {
		keys[u.Key] = u
	}
	return &APIKeyAuth{keys: keys}
}

// Authenticate resolves an API key to a User.
// If no users are configured (auth disabled), returns anonymous.
func (a *APIKeyAuth) Authenticate(token string) (*User, error) {
	if len(a.keys) == 0 {
		return Anonymous(), nil
	}

	if token == "" {
		return nil, fmt.Errorf("auth: no API key provided")
	}

	u, ok := a.keys[token]
	if !ok {
		return nil, fmt.Errorf("auth: invalid API key")
	}

	return &User{
		Name:          u.Name,
		Roles:         u.Roles,
		Authenticated: true,
	}, nil
}
```

**Step 5: Run tests**

```bash
go test ./internal/auth/ -v
```

Expected: 7 tests PASS (3 from Task 1 + 4 new).

**Step 6: Commit**

```bash
git add internal/auth/ internal/config/config.go
git commit -m "feat: API key authenticator with config-based user lookup"
```

---

### Task 3: Policy Engine — Server-Level Access

**Files:**
- Create: `internal/policy/engine.go`
- Create: `internal/policy/engine_test.go`

**Step 1: Write the test**

```go
// internal/policy/engine_test.go
package policy

import (
	"testing"

	"github.com/akibar/mcp-auth-gateway/internal/auth"
	"github.com/akibar/mcp-auth-gateway/internal/config"
)

func TestCheckServerAccess_Allowed(t *testing.T) {
	e := New()
	user := &auth.User{Name: "akibar", Roles: []string{"engineer"}, Authenticated: true}
	serverCfg := config.ServerConfig{
		Name: "sourcegraph",
		Policies: config.PolicyConfig{
			AllowedRoles: []string{"engineer", "admin"},
		},
	}

	result := e.CheckServerAccess(user, serverCfg)
	if !result.Allowed {
		t.Errorf("should allow engineer access, got: %s", result.Reason)
	}
}

func TestCheckServerAccess_Denied(t *testing.T) {
	e := New()
	user := &auth.User{Name: "intern", Roles: []string{"viewer"}, Authenticated: true}
	serverCfg := config.ServerConfig{
		Name: "database",
		Policies: config.PolicyConfig{
			AllowedRoles: []string{"admin"},
		},
	}

	result := e.CheckServerAccess(user, serverCfg)
	if result.Allowed {
		t.Error("should deny viewer access to admin-only server")
	}
	if result.Reason == "" {
		t.Error("denied result should have a reason")
	}
}

func TestCheckServerAccess_NoPolicies(t *testing.T) {
	e := New()
	user := &auth.User{Name: "anyone", Roles: []string{"viewer"}, Authenticated: true}
	serverCfg := config.ServerConfig{
		Name:     "public",
		Policies: config.PolicyConfig{},
	}

	result := e.CheckServerAccess(user, serverCfg)
	if !result.Allowed {
		t.Error("should allow access when no policies configured")
	}
}
```

**Step 2: Run test — verify it fails**

```bash
go test ./internal/policy/ -v
```

Expected: FAIL — package doesn't exist.

**Step 3: Write implementation**

```go
// internal/policy/engine.go
package policy

import (
	"fmt"

	"github.com/akibar/mcp-auth-gateway/internal/auth"
	"github.com/akibar/mcp-auth-gateway/internal/config"
)

// Decision is the result of a policy check.
type Decision struct {
	Allowed bool
	Reason  string
}

func allow() Decision {
	return Decision{Allowed: true}
}

func deny(reason string) Decision {
	return Decision{Allowed: false, Reason: reason}
}

// Engine evaluates access policies.
type Engine struct{}

// New creates a new policy engine.
func New() *Engine {
	return &Engine{}
}

// CheckServerAccess checks if a user can access a server at all.
func (e *Engine) CheckServerAccess(user *auth.User, server config.ServerConfig) Decision {
	if len(server.Policies.AllowedRoles) == 0 {
		return allow()
	}

	if !user.HasAnyRole(server.Policies.AllowedRoles) {
		return deny(fmt.Sprintf(
			"user %q (roles: %v) not authorized for server %q (requires: %v)",
			user.Name, user.Roles, server.Name, server.Policies.AllowedRoles,
		))
	}

	return allow()
}
```

**Step 4: Run tests**

```bash
go test ./internal/policy/ -v
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add internal/policy/
git commit -m "feat: policy engine with server-level access control"
```

---

### Task 4: Policy Engine — Tool-Level and Argument-Level Access

**Files:**
- Modify: `internal/policy/engine.go`
- Modify: `internal/policy/engine_test.go`

**Step 1: Add tool-level and argument-level tests**

Append to `internal/policy/engine_test.go`:

```go
func TestCheckToolAccess_AllowedByDefault(t *testing.T) {
	e := New()
	user := &auth.User{Name: "akibar", Roles: []string{"engineer"}, Authenticated: true}
	serverCfg := config.ServerConfig{
		Name:     "slack",
		Policies: config.PolicyConfig{},
	}

	result := e.CheckToolAccess(user, serverCfg, "read_slack_message", nil)
	if !result.Allowed {
		t.Error("should allow when no tool policy defined")
	}
}

func TestCheckToolAccess_RequiresRole(t *testing.T) {
	e := New()
	user := &auth.User{Name: "akibar", Roles: []string{"engineer"}, Authenticated: true}
	serverCfg := config.ServerConfig{
		Name: "slack",
		Policies: config.PolicyConfig{
			Tools: map[string]config.ToolPolicy{
				"post_slack_message": {AllowedRoles: []string{"admin"}},
			},
		},
	}

	result := e.CheckToolAccess(user, serverCfg, "post_slack_message", nil)
	if result.Allowed {
		t.Error("engineer should not access admin-only tool")
	}

	result = e.CheckToolAccess(user, serverCfg, "read_slack_message", nil)
	if !result.Allowed {
		t.Error("should allow tool with no specific policy")
	}
}

func TestCheckToolAccess_BlockedArgs(t *testing.T) {
	e := New()
	user := &auth.User{Name: "admin", Roles: []string{"admin"}, Authenticated: true}
	serverCfg := config.ServerConfig{
		Name: "database",
		Policies: config.PolicyConfig{
			AllowedRoles: []string{"admin"},
			BlockedArgs: []config.BlockedArg{
				{Pattern: "(?i)DROP\\s+TABLE"},
				{Pattern: "(?i)DELETE\\s+FROM.*WHERE\\s+1\\s*=\\s*1"},
			},
		},
	}

	// Safe query — should be allowed
	result := e.CheckToolAccess(user, serverCfg, "execute_query", map[string]string{
		"query": "SELECT * FROM users",
	})
	if !result.Allowed {
		t.Errorf("safe query should be allowed, got: %s", result.Reason)
	}

	// Dangerous query — should be blocked
	result = e.CheckToolAccess(user, serverCfg, "execute_query", map[string]string{
		"query": "DROP TABLE users",
	})
	if result.Allowed {
		t.Error("DROP TABLE should be blocked")
	}

	// Another dangerous query
	result = e.CheckToolAccess(user, serverCfg, "execute_query", map[string]string{
		"query": "DELETE FROM users WHERE 1=1",
	})
	if result.Allowed {
		t.Error("DELETE FROM WHERE 1=1 should be blocked")
	}
}
```

**Step 2: Run test — verify new tests fail**

```bash
go test ./internal/policy/ -v
```

Expected: FAIL — `CheckToolAccess` not defined.

**Step 3: Add tool-level and argument-level checks to engine.go**

Append to `internal/policy/engine.go`:

```go
import "regexp"

// CheckToolAccess checks if a user can call a specific tool, including argument inspection.
func (e *Engine) CheckToolAccess(user *auth.User, server config.ServerConfig, toolName string, args map[string]string) Decision {
	// Check tool-level policy
	if toolPolicy, exists := server.Policies.Tools[toolName]; exists {
		allowed := toolPolicy.AllowedRoles
		if toolPolicy.RequiresRole != "" {
			allowed = append(allowed, toolPolicy.RequiresRole)
		}
		if len(allowed) > 0 && !user.HasAnyRole(allowed) {
			return deny(fmt.Sprintf(
				"user %q not authorized for tool %q (requires: %v)",
				user.Name, toolName, allowed,
			))
		}
	}

	// Check argument-level blocked patterns
	for _, blocked := range server.Policies.BlockedArgs {
		re, err := regexp.Compile(blocked.Pattern)
		if err != nil {
			return deny(fmt.Sprintf("invalid blocked_args pattern %q: %v", blocked.Pattern, err))
		}
		for argName, argValue := range args {
			if re.MatchString(argValue) {
				return deny(fmt.Sprintf(
					"argument %q matches blocked pattern %q",
					argName, blocked.Pattern,
				))
			}
		}
	}

	return allow()
}
```

**Step 4: Run tests**

```bash
go test ./internal/policy/ -v
```

Expected: 6 tests PASS.

**Step 5: Commit**

```bash
git add internal/policy/
git commit -m "feat: tool-level and argument-level policy checks"
```

---

### Task 5: Wire Auth + Policy into the Proxy Loop

**Files:**
- Modify: `main.go`
- Modify: `internal/gateway/gateway.go`

**Step 1: Add policy engine and user to Gateway**

In `internal/gateway/gateway.go`, add fields and a method for the policy check:

```go
// Add imports: "github.com/akibar/mcp-auth-gateway/internal/auth" and "github.com/akibar/mcp-auth-gateway/internal/policy"

// Update Gateway struct — add:
//   Policy *policy.Engine
//   User   *auth.User

// Update New() to initialize Policy:
//   Policy: policy.New(),

// Add method:
// ServerConfigByName returns the config for a named server.
func (g *Gateway) ServerConfigByName(name string) (config.ServerConfig, bool) {
	for _, s := range g.Config.Servers {
		if s.Name == name {
			return s, true
		}
	}
	return config.ServerConfig{}, false
}
```

**Step 2: Update main.go — authenticate on startup, check policies on tools/call**

In `runStart`, after loading config and before the proxy loop:

```go
// After: gw := gateway.New(cfg)
// Add auth setup:
authenticator := auth.NewAPIKeyAuth(cfg.Auth.Users)
token := os.Getenv("MCP_AUTH_TOKEN")
user, err := authenticator.Authenticate(token)
if err != nil {
    return fmt.Errorf("auth: %w", err)
}
gw.User = user
slog.Info("authenticated", "user", user.Name, "roles", user.Roles, "authenticated", user.Authenticated)
```

In the `tools/call` case, after routing to the server name but before forwarding:

```go
// After: serverName, ok := gw.RouteToolCall(params.Name)
// After: srv, err := gw.GetServer(serverName)

// Add policy checks:
serverCfg, _ := gw.ServerConfigByName(serverName)

// Check server-level access
decision := gw.Policy.CheckServerAccess(gw.User, serverCfg)
if !decision.Allowed {
    slog.Warn("policy denied", "user", gw.User.Name, "server", serverName, "tool", params.Name, "reason", decision.Reason)
    errResp := &mcp.JSONRPCMessage{
        JSONRPC: "2.0",
        ID:      msg.ID,
        Error:   &mcp.JSONRPCError{Code: -32600, Message: fmt.Sprintf("access denied: %s", decision.Reason)},
    }
    transport.WriteMessage(os.Stdout, errResp)
    continue
}

// Check tool-level + argument-level access
decision = gw.Policy.CheckToolAccess(gw.User, serverCfg, params.Name, params.Arguments)
if !decision.Allowed {
    slog.Warn("policy denied", "user", gw.User.Name, "server", serverName, "tool", params.Name, "reason", decision.Reason)
    errResp := &mcp.JSONRPCMessage{
        JSONRPC: "2.0",
        ID:      msg.ID,
        Error:   &mcp.JSONRPCError{Code: -32600, Message: fmt.Sprintf("access denied: %s", decision.Reason)},
    }
    transport.WriteMessage(os.Stdout, errResp)
    continue
}

slog.Info("policy allowed", "user", gw.User.Name, "server", serverName, "tool", params.Name)
```

**Step 3: Build and run all tests**

```bash
cd ~/projects/mcp-auth-gateway
go build -o mcp-gateway .
go test ./... -v
```

Expected: Builds without errors. All existing + new tests pass.

**Step 4: Commit**

```bash
git add main.go internal/gateway/gateway.go
git commit -m "feat: wire auth + policy engine into proxy loop"
```

---

### Task 6: Update Example Config with Auth + Policies

**Files:**
- Modify: `gateway.yaml`

**Step 1: Update gateway.yaml with auth and policy examples**

```yaml
# gateway.yaml — Configuration with auth + policies
gateway:
  listen: "localhost:3100"

auth:
  provider: apikey
  users:
    - key: "sk-eng-abc123"
      name: "akibar"
      roles: ["engineer"]
    - key: "sk-admin-xyz789"
      name: "admin-bot"
      roles: ["admin", "engineer"]

servers:
  - name: filesystem
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/private/tmp/mcp-gateway-test"]
    policies:
      allowed_roles: ["engineer", "admin"]
      tools:
        write_file:
          allowed_roles: ["admin"]
        edit_file:
          allowed_roles: ["admin"]
      blocked_args:
        - pattern: "(?i)\\.env$"
        - pattern: "(?i)credentials"
```

**Step 2: Update .mcp.json to pass auth token via env**

```json
{
  "mcpServers": {
    "mcp-gateway": {
      "command": "/Users/akibar/projects/mcp-auth-gateway/mcp-gateway",
      "args": ["start", "--config", "/Users/akibar/projects/mcp-auth-gateway/gateway.yaml"],
      "env": {
        "MCP_AUTH_TOKEN": "sk-eng-abc123"
      }
    }
  }
}
```

**Step 3: Commit**

```bash
git add gateway.yaml .mcp.json
git commit -m "feat: example config with auth users and policy rules"
```

---

### Task 7: Config Validation Tests

**Files:**
- Modify: `internal/config/config_test.go`

**Step 1: Add test for auth + policy config parsing**

Append to `internal/config/config_test.go`:

```go
func TestLoadConfig_WithAuth(t *testing.T) {
	yaml := `
auth:
  provider: apikey
  users:
    - key: "sk-test-123"
      name: "tester"
      roles: ["engineer", "oncall"]

servers:
  - name: database
    command: "echo"
    args: ["hello"]
    policies:
      allowed_roles: ["admin"]
      tools:
        execute_query:
          allowed_roles: ["admin"]
      blocked_args:
        - pattern: "DROP TABLE"
`
	tmp := filepath.Join(t.TempDir(), "auth.yaml")
	os.WriteFile(tmp, []byte(yaml), 0644)

	cfg, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Auth.Provider != "apikey" {
		t.Errorf("Provider = %q, want %q", cfg.Auth.Provider, "apikey")
	}

	if len(cfg.Auth.Users) != 1 {
		t.Fatalf("Users count = %d, want 1", len(cfg.Auth.Users))
	}

	if cfg.Auth.Users[0].Name != "tester" {
		t.Errorf("User name = %q, want %q", cfg.Auth.Users[0].Name, "tester")
	}

	if len(cfg.Auth.Users[0].Roles) != 2 {
		t.Errorf("Roles count = %d, want 2", len(cfg.Auth.Users[0].Roles))
	}

	server := cfg.Servers[0]
	if len(server.Policies.AllowedRoles) != 1 || server.Policies.AllowedRoles[0] != "admin" {
		t.Errorf("AllowedRoles = %v, want [admin]", server.Policies.AllowedRoles)
	}

	if len(server.Policies.BlockedArgs) != 1 {
		t.Fatalf("BlockedArgs count = %d, want 1", len(server.Policies.BlockedArgs))
	}

	if server.Policies.BlockedArgs[0].Pattern != "DROP TABLE" {
		t.Errorf("Pattern = %q, want %q", server.Policies.BlockedArgs[0].Pattern, "DROP TABLE")
	}
}
```

**Step 2: Run tests**

```bash
go test ./internal/config/ -v
```

Expected: 3 tests PASS.

**Step 3: Commit**

```bash
git add internal/config/config_test.go
git commit -m "test: config parsing for auth users and policy rules"
```

---

### Task 8: E2E Integration Test — Auth Allowed + Denied

**Files:**
- Create: `test/e2e_auth_test.sh`

**Step 1: Create integration test script**

```bash
#!/bin/bash
# test/e2e_auth_test.sh — Test auth + policy engine end-to-end
set -e

cd ~/projects/mcp-auth-gateway
go build -o mcp-gateway .

mkdir -p /private/tmp/mcp-gateway-test
echo "secret data" > /private/tmp/mcp-gateway-test/hello.txt

echo "=== Test 1: Authenticated engineer — read_file ALLOWED ==="
OUTPUT=$(MCP_AUTH_TOKEN=sk-eng-abc123 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"read_file","arguments":{"path":"/private/tmp/mcp-gateway-test/hello.txt"}}}' \
  | MCP_AUTH_TOKEN=sk-eng-abc123 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/tmp/auth-test.log)

if echo "$OUTPUT" | grep -q "secret data"; then
  echo "PASS: Engineer can read files"
else
  echo "FAIL: Engineer should be able to read files"
  echo "$OUTPUT"
fi

echo ""
echo "=== Test 2: Authenticated engineer — write_file DENIED (admin only) ==="
OUTPUT=$(MCP_AUTH_TOKEN=sk-eng-abc123 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"write_file","arguments":{"path":"/private/tmp/mcp-gateway-test/hack.txt","content":"pwned"}}}' \
  | MCP_AUTH_TOKEN=sk-eng-abc123 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/tmp/auth-test2.log)

if echo "$OUTPUT" | grep -q "access denied"; then
  echo "PASS: Engineer blocked from write_file (admin only)"
else
  echo "FAIL: Engineer should be blocked from write_file"
  echo "$OUTPUT"
fi

echo ""
echo "=== Test 3: No auth token — DENIED ==="
OUTPUT=$(printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  | perl -e 'alarm 10; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>&1 || true)

if echo "$OUTPUT" | grep -qi "auth"; then
  echo "PASS: No token rejected"
else
  echo "FAIL: Should reject missing token"
  echo "$OUTPUT"
fi

echo ""
echo "=== Test 4: Admin — write_file ALLOWED ==="
OUTPUT=$(MCP_AUTH_TOKEN=sk-admin-xyz789 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"write_file","arguments":{"path":"/private/tmp/mcp-gateway-test/admin-wrote.txt","content":"admin was here"}}}' \
  | MCP_AUTH_TOKEN=sk-admin-xyz789 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/tmp/auth-test4.log)

if echo "$OUTPUT" | grep -qv "access denied"; then
  echo "PASS: Admin can write files"
else
  echo "FAIL: Admin should be able to write files"
  echo "$OUTPUT"
fi

echo ""
echo "=== All auth tests complete ==="
```

**Step 2: Run integration test**

```bash
chmod +x ~/projects/mcp-auth-gateway/test/e2e_auth_test.sh
bash ~/projects/mcp-auth-gateway/test/e2e_auth_test.sh
```

Expected: All 4 tests PASS.

**Step 3: Commit**

```bash
cd ~/projects/mcp-auth-gateway
git add test/e2e_auth_test.sh
git commit -m "test: e2e auth + policy integration tests (allowed/denied scenarios)"
```

---

## Summary

| Task | What it builds | Tests |
|------|---------------|-------|
| 1 | Auth identity types (User, roles) | 3 unit tests |
| 2 | API key authenticator | 4 unit tests |
| 3 | Policy engine — server-level | 3 unit tests |
| 4 | Policy engine — tool + argument level | 3 unit tests |
| 5 | Wire auth + policy into proxy loop | Build + all tests |
| 6 | Example config with auth + policies | Config update |
| 7 | Config validation tests | 1 unit test |
| 8 | E2E integration test | 4 scenarios (allow/deny) |

After completing all 8 tasks: unauthorized calls are blocked, role-based tool access works, dangerous arguments are rejected, and all of it is tested end-to-end.
