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
