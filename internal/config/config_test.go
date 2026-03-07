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
