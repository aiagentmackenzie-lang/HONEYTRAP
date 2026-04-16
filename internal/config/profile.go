package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DeployProfile represents a YAML deploy profile configuration.
type DeployProfile struct {
	Services map[string]ServiceProfile `yaml:"services"`
	AI       AIProfile                 `yaml:"ai"`
	Alerts   AlertsProfile             `yaml:"alerts"`
	Export   ExportProfile             `yaml:"export"`
	Logging  LoggingProfile            `yaml:"logging"`
}

type ServiceProfile struct {
	Enabled       bool   `yaml:"enabled"`
	Port          int    `yaml:"port"`
	Banner        string `yaml:"banner,omitempty"`
	AIEmulation   bool   `yaml:"ai_emulation,omitempty"`
	FakeLogin     bool   `yaml:"fake_login,omitempty"`
	FakeDashboard bool   `yaml:"fake_dashboard,omitempty"`
	FakeAPI       bool   `yaml:"fake_api,omitempty"`
	MaxSessions   int    `yaml:"max_sessions"`
}

type AIProfile struct {
	Enabled    bool   `yaml:"enabled"`
	OllamaURL  string `yaml:"ollama_url,omitempty"`
	Model      string `yaml:"model,omitempty"`
	CacheSize  int    `yaml:"cache_size,omitempty"`
	CacheTTL   int    `yaml:"cache_ttl,omitempty"`
	Fallback   bool   `yaml:"fallback"`
}

type AlertsProfile struct {
	Slack    SlackProfile    `yaml:"slack"`
	Telegram TelegramProfile `yaml:"telegram"`
	Email    EmailProfile    `yaml:"email"`
}

type SlackProfile struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url,omitempty"`
}

type TelegramProfile struct {
	Enabled  bool   `yaml:"enabled"`
	BotToken string `yaml:"bot_token,omitempty"`
	ChatID   string `yaml:"chat_id,omitempty"`
}

type EmailProfile struct {
	Enabled  bool   `yaml:"enabled"`
	SMTPHost string `yaml:"smtp_host,omitempty"`
	SMTPPort int    `yaml:"smtp_port,omitempty"`
	From     string `yaml:"from,omitempty"`
	To       string `yaml:"to,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type ExportProfile struct {
	STIXEnabled    bool `yaml:"stix_enabled"`
	AutoExport     bool `yaml:"auto_export"`
	ExportInterval int  `yaml:"export_interval,omitempty"`
}

type LoggingProfile struct {
	Level       string `yaml:"level"`
	PCAPCapture bool   `yaml:"pcap_capture"`
	PCAPDir     string `yaml:"pcap_dir,omitempty"`
}

// LoadProfile reads a YAML profile from the profiles/ directory.
func LoadProfile(name string) (*DeployProfile, error) {
	profileDir := "profiles"
	if envDir := os.Getenv("HONEYTRAP_PROFILES_DIR"); envDir != "" {
		profileDir = envDir
	}

	path := filepath.Join(profileDir, name+".yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("profile %q not found: %w", name, err)
	}

	var profile DeployProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("invalid profile %q: %w", name, err)
	}

	// Expand environment variables in alert configs
	expandEnv(&profile.Alerts.Slack.WebhookURL)
	expandEnv(&profile.Alerts.Telegram.BotToken)
	expandEnv(&profile.Alerts.Telegram.ChatID)
	expandEnv(&profile.Alerts.Email.SMTPHost)
	expandEnv(&profile.Alerts.Email.To)
	expandEnv(&profile.Alerts.Email.Username)
	expandEnv(&profile.Alerts.Email.Password)

	return &profile, nil
}

// ListProfiles returns available profile names.
func ListProfiles() ([]string, error) {
	profileDir := "profiles"
	if envDir := os.Getenv("HONEYTRAP_PROFILES_DIR"); envDir != "" {
		profileDir = envDir
	}

	entries, err := os.ReadDir(profileDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read profiles directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yml" {
			names = append(names, entry.Name()[:len(entry.Name())-4])
		}
	}
	return names, nil
}

func expandEnv(s *string) {
	if s != nil && *s != "" {
		*s = os.ExpandEnv(*s)
	}
}