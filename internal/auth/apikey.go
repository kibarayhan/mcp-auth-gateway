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
