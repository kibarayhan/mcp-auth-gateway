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
