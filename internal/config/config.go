package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	NodeName       string
	Environment    string
	DatabaseURL    string
	DataDir        string
	DefaultProfile string
	Services       []ServiceConfig
}

type ServiceConfig struct {
	Name     string
	Protocol string
	Address  string
	Enabled  bool
}

func Load() (Config, error) {
	dataDir := getenv("HONEYTRAP_DATA_DIR", "var")
	cfg := Config{
		NodeName:       getenv("HONEYTRAP_NODE_NAME", "local-node"),
		Environment:    getenv("HONEYTRAP_ENV", "development"),
		DatabaseURL:    os.Getenv("HONEYTRAP_DATABASE_URL"),
		DataDir:        dataDir,
		DefaultProfile: getenv("HONEYTRAP_PROFILE", "default"),
		Services: []ServiceConfig{
			{Name: "ssh", Protocol: "tcp", Address: fmt.Sprintf(":%d", getenvInt("HONEYTRAP_SSH_PORT", 2222)), Enabled: getenvBool("HONEYTRAP_ENABLE_SSH", true)},
			{Name: "ssh-enhanced", Protocol: "tcp", Address: fmt.Sprintf(":%d", getenvInt("HONEYTRAP_SSH_ENHANCED_PORT", 2223)), Enabled: getenvBool("HONEYTRAP_ENABLE_SSH_ENHANCED", true)},
			{Name: "http", Protocol: "tcp", Address: fmt.Sprintf(":%d", getenvInt("HONEYTRAP_HTTP_PORT", 8080)), Enabled: getenvBool("HONEYTRAP_ENABLE_HTTP", true)},
			{Name: "http-enhanced", Protocol: "tcp", Address: fmt.Sprintf(":%d", getenvInt("HONEYTRAP_HTTP_ENHANCED_PORT", 8443)), Enabled: getenvBool("HONEYTRAP_ENABLE_HTTP_ENHANCED", true)},
			{Name: "ftp", Protocol: "tcp", Address: fmt.Sprintf(":%d", getenvInt("HONEYTRAP_FTP_PORT", 2121)), Enabled: getenvBool("HONEYTRAP_ENABLE_FTP", true)},
			{Name: "redis", Protocol: "tcp", Address: fmt.Sprintf(":%d", getenvInt("HONEYTRAP_REDIS_PORT", 6379)), Enabled: getenvBool("HONEYTRAP_ENABLE_REDIS", true)},
			{Name: "udp-decoy", Protocol: "udp", Address: fmt.Sprintf(":%d", getenvInt("HONEYTRAP_UDP_PORT", 9161)), Enabled: getenvBool("HONEYTRAP_ENABLE_UDP", true)},
		},
	}

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return Config{}, fmt.Errorf("create data dir: %w", err)
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getenvBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func (c Config) ActiveServices() []ServiceConfig {
	services := make([]ServiceConfig, 0, len(c.Services))
	for _, svc := range c.Services {
		if svc.Enabled {
			services = append(services, svc)
		}
	}
	return services
}

func (c Config) SessionLogPath() string {
	return fmt.Sprintf("%s/sessions.jsonl", c.DataDir)
}

func (c Config) EventLogPath() string {
	return fmt.Sprintf("%s/events.jsonl", c.DataDir)
}

func (c Config) StartedAt() time.Time {
	return time.Now().UTC()
}
