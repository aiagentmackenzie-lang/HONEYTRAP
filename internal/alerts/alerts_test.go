package alerts

import (
	"testing"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/config"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
)

func TestSessionAlert_Severity(t *testing.T) {
	mgr := NewManager(config.AlertsProfile{})

	tests := []struct {
		service  string
		expected string
	}{
		{"ssh-enhanced", "high"},
		{"http-enhanced", "high"},
		{"ssh", "medium"},
		{"http", "medium"},
		{"ftp", "medium"},
		{"redis", "medium"},
	}

	for _, tt := range tests {
		session := models.Session{
			ID:         "s-1",
			Service:    tt.service,
			RemoteIP:   "185.220.101.1",
			StartedAt:  time.Now(),
		}
		alert := mgr.SessionAlert(session)
		if alert.Type != "session" {
			t.Errorf("expected type 'session', got %q", alert.Type)
		}
		if alert.Severity != tt.expected {
			t.Errorf("service %s: expected severity %q, got %q", tt.service, tt.expected, alert.Severity)
		}
	}
}

func TestSessionAlert_Fields(t *testing.T) {
	mgr := NewManager(config.AlertsProfile{})
	session := models.Session{
		ID:        "s-test",
		Service:   "ssh",
		RemoteIP:  "10.0.0.1",
		StartedAt: time.Now(),
	}

	alert := mgr.SessionAlert(session)
	if alert.SourceIP != "10.0.0.1" {
		t.Errorf("expected source IP '10.0.0.1', got %q", alert.SourceIP)
	}
	if alert.Service != "ssh" {
		t.Errorf("expected service 'ssh', got %q", alert.Service)
	}
	if alert.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}
}

func TestTokenAccessAlert(t *testing.T) {
	mgr := NewManager(config.AlertsProfile{})
	token := models.Token{
		ID:   "t-1",
		Kind: "aws_credentials",
	}

	alert := mgr.TokenAccessAlert(token, "103.224.1.1")
	if alert.Type != "token_access" {
		t.Errorf("expected type 'token_access', got %q", alert.Type)
	}
	if alert.Severity != "critical" {
		t.Errorf("expected critical severity, got %q", alert.Severity)
	}
	if alert.SourceIP != "103.224.1.1" {
		t.Errorf("expected source IP '103.224.1.1', got %q", alert.SourceIP)
	}
}

func TestCredentialAlert(t *testing.T) {
	mgr := NewManager(config.AlertsProfile{})
	alert := mgr.CredentialAlert("HTTP+", "admin", "10.0.0.5")

	if alert.Type != "credential" {
		t.Errorf("expected type 'credential', got %q", alert.Type)
	}
	if alert.Severity != "high" {
		t.Errorf("expected high severity, got %q", alert.Severity)
	}
	if alert.Service != "HTTP+" {
		t.Errorf("expected service 'HTTP+', got %q", alert.Service)
	}
}

func TestSendAlert_Disabled(t *testing.T) {
	mgr := NewManager(config.AlertsProfile{
		Slack:    config.SlackProfile{Enabled: false},
		Telegram: config.TelegramProfile{Enabled: false},
		Email:    config.EmailProfile{Enabled: false},
	})

	// SendAlert with all integrations disabled should return nil
	alert := Alert{Type: "session", Severity: "high", Message: "test"}
	err := mgr.SendAlert(alert)
	if err != nil {
		t.Errorf("expected nil error with all integrations disabled, got %v", err)
	}
}