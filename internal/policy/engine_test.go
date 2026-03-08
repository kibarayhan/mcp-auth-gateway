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
				{Pattern: `(?i)DROP\s+TABLE`},
				{Pattern: `(?i)DELETE\s+FROM.*WHERE\s+1\s*=\s*1`},
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
