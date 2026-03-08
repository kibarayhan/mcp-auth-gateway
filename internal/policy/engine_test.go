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
