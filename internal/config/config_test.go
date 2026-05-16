package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear all HONEYTRAP env vars to test defaults
	envVars := []string{
		"HONEYTRAP_DATA_DIR", "HONEYTRAP_NODE_NAME", "HONEYTRAP_ENV",
		"HONEYTRAP_SSH_PORT", "HONEYTRAP_HTTP_PORT", "HONEYTRAP_ENABLE_SSH",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.NodeName != "local-node" {
		t.Errorf("expected NodeName 'local-node', got %q", cfg.NodeName)
	}
	if cfg.Environment != "development" {
		t.Errorf("expected Environment 'development', got %q", cfg.Environment)
	}
	if len(cfg.Services) != 7 {
		t.Errorf("expected 7 default services, got %d", len(cfg.Services))
	}
	if cfg.SessionLogPath() != "var/sessions.jsonl" {
		t.Errorf("expected default session log path, got %q", cfg.SessionLogPath())
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	os.Setenv("HONEYTRAP_SSH_PORT", "9999")
	os.Setenv("HONEYTRAP_ENABLE_FTP", "false")
	defer os.Unsetenv("HONEYTRAP_SSH_PORT")
	defer os.Unsetenv("HONEYTRAP_ENABLE_FTP")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Find the SSH service
	var sshSvc *ServiceConfig
	for i := range cfg.Services {
		if cfg.Services[i].Name == "ssh" {
			sshSvc = &cfg.Services[i]
			break
		}
	}
	if sshSvc == nil {
		t.Fatal("ssh service not found")
	}
	if sshSvc.Address != ":9999" {
		t.Errorf("expected SSH address ':9999', got %q", sshSvc.Address)
	}

	// Find the FTP service
	var ftpSvc *ServiceConfig
	for i := range cfg.Services {
		if cfg.Services[i].Name == "ftp" {
			ftpSvc = &cfg.Services[i]
			break
		}
	}
	if ftpSvc == nil {
		t.Fatal("ftp service not found")
	}
	if ftpSvc.Enabled {
		t.Error("expected FTP to be disabled")
	}
}

func TestActiveServices(t *testing.T) {
	os.Unsetenv("HONEYTRAP_ENABLE_UDP")
	cfg, _ := Load()
	active := cfg.ActiveServices()
	if len(active) == 0 {
		t.Error("expected at least one active service")
	}
	for _, svc := range active {
		if !svc.Enabled {
			t.Errorf("active service %s should have Enabled=true", svc.Name)
		}
	}
}

func TestApplyProfile(t *testing.T) {
	os.Unsetenv("HONEYTRAP_ENABLE_HTTP")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	profile := &DeployProfile{
		Services: map[string]ServiceProfile{
			"ssh":   {Enabled: true, Port: 9999},
			"udp":   {Enabled: false, Port: 7777},
			"http_enhanced": {Enabled: true, Port: 443},
		},
		AI: AIProfile{
			Enabled:   true,
			OllamaURL: "http://custom-ollama:11434",
		},
	}

	result := ApplyProfile(&cfg, profile)

	// Check SSH port override
	var sshSvc *ServiceConfig
	for i := range result.Services {
		if result.Services[i].Name == "ssh" {
			sshSvc = &result.Services[i]
			break
		}
	}
	if sshSvc == nil {
		t.Fatal("ssh service not found")
	}
	if sshSvc.Address != ":9999" {
		t.Errorf("expected SSH port 9999, got %q", sshSvc.Address)
	}

	// Check UDP disabled
	var udpSvc *ServiceConfig
	for i := range result.Services {
		if result.Services[i].Name == "udp-decoy" {
			udpSvc = &result.Services[i]
			break
		}
	}
	if udpSvc == nil {
		t.Fatal("udp-decoy service not found")
	}
	if udpSvc.Enabled {
		t.Error("expected UDP to be disabled by profile")
	}

	// Check AI URL
	if os.Getenv("HONEYTRAP_AI_URL") != "http://custom-ollama:11434" {
		t.Errorf("expected HONEYTRAP_AI_URL to be set by profile, got %q", os.Getenv("HONEYTRAP_AI_URL"))
	}
}

func TestNormalizeServiceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ssh", "ssh"},
		{"http", "http"},
		{"ftp", "ftp"},
		{"redis", "redis"},
		{"ssh_enhanced", "ssh-enhanced"},
		{"http_enhanced", "http-enhanced"},
		{"udp", "udp-decoy"},
	}
	for _, tt := range tests {
		got := normalizeServiceName(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeServiceName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestLoadProfile(t *testing.T) {
	os.Setenv("HONEYTRAP_PROFILES_DIR", "../../profiles")
	defer os.Unsetenv("HONEYTRAP_PROFILES_DIR")

	profile, err := LoadProfile("default")
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}
	if len(profile.Services) == 0 {
		t.Error("expected services in default profile")
	}
	if !profile.AI.Enabled {
		t.Error("expected AI to be enabled in default profile")
	}
}