package policy

import (
	"fmt"
	"regexp"

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
