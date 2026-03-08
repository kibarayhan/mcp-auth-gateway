package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Gateway GatewayConfig  `yaml:"gateway"`
	Servers []ServerConfig `yaml:"servers"`
	Auth    AuthConfig     `yaml:"auth"`
	Audit   AuditConfig    `yaml:"audit"`
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
