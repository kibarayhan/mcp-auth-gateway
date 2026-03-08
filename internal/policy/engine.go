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
