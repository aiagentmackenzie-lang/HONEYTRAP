package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/config"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
)

// Alert represents an alert to be sent.
type Alert struct {
	Type      string `json:"type"`     // "session", "token_access", "credential"
	Severity  string `json:"severity"` // "low", "medium", "high", "critical"
	Service   string `json:"service"`
	SourceIP  string `json:"source_ip"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// Manager handles alert routing to configured integrations.
type Manager struct {
	config config.AlertsProfile
	client *http.Client
}

// NewManager creates a new alert manager.
func NewManager(cfg config.AlertsProfile) *Manager {
	return &Manager{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendAlert dispatches an alert to all enabled integrations.
func (m *Manager) SendAlert(alert Alert) error {
	var errs []error

	if m.config.Slack.Enabled {
		if err := m.sendSlack(alert); err != nil {
			errs = append(errs, fmt.Errorf("slack: %w", err))
		}
	}

	if m.config.Telegram.Enabled {
		if err := m.sendTelegram(alert); err != nil {
			errs = append(errs, fmt.Errorf("telegram: %w", err))
		}
	}

	if m.config.Email.Enabled {
		if err := m.sendEmail(alert); err != nil {
			errs = append(errs, fmt.Errorf("email: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("alert partial failure: %v", errs)
	}
	return nil
}

// SessionAlert creates an alert from a new session.
func (m *Manager) SessionAlert(session models.Session) Alert {
	severity := "medium"
	if session.Service == "ssh_enhanced" || session.Service == "http_enhanced" {
		severity = "high"
	}

	return Alert{
		Type:      "session",
		Severity:  severity,
		Service:   session.Service,
		SourceIP:  session.RemoteIP,
		Message:   fmt.Sprintf("New honeypot session on %s from %s", session.Service, session.RemoteIP),
		Timestamp: session.StartedAt.UTC().Format(time.RFC3339),
	}
}

// TokenAccessAlert creates an alert from a token access.
func (m *Manager) TokenAccessAlert(token models.Token, accessorIP string) Alert {
	return Alert{
		Type:      "token_access",
		Severity:  "critical",
		Service:   "honeytoken",
		SourceIP:  accessorIP,
		Message:   fmt.Sprintf("⚠️ CRITICAL: Honeytoken accessed! Kind=%s, IP=%s", token.Kind, accessorIP),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// CredentialAlert creates an alert from captured credentials.
func (m *Manager) CredentialAlert(service, username, sourceIP string) Alert {
	return Alert{
		Type:      "credential",
		Severity:  "high",
		Service:   service,
		SourceIP:  sourceIP,
		Message:   fmt.Sprintf("Credentials captured on %s: username=%s, IP=%s", service, username, sourceIP),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func (m *Manager) sendSlack(alert Alert) error {
	webhookURL := m.config.Slack.WebhookURL
	if webhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	emoji := "🔔"
	switch alert.Severity {
	case "critical":
		emoji = "🚨"
	case "high":
		emoji = "🔴"
	case "medium":
		emoji = "🟡"
	case "low":
		emoji = "🟢"
	}

	payload := map[string]interface{}{
		"text": fmt.Sprintf("%s *HONEYTRAP Alert* [%s]\n• *Type:* %s\n• *Service:* %s\n• *Source IP:* `%s`\n• *Message:* %s",
			emoji, alert.Severity, alert.Type, alert.Service, alert.SourceIP, alert.Message),
	}

	body, _ := json.Marshal(payload)
	resp, err := m.client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func (m *Manager) sendTelegram(alert Alert) error {
	botToken := m.config.Telegram.BotToken
	chatID := m.config.Telegram.ChatID
	if botToken == "" || chatID == "" {
		return fmt.Errorf("telegram bot token or chat ID not configured")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	severityEmoji := "🟡"
	switch alert.Severity {
	case "critical":
		severityEmoji = "🚨"
	case "high":
		severityEmoji = "🔴"
	case "low":
		severityEmoji = "🟢"
	}

	text := fmt.Sprintf("%s *HONEYTRAP Alert* [%s]\nType: %s\nService: %s\nIP: `%s`\n%s",
		severityEmoji, alert.Severity, alert.Type, alert.Service, alert.SourceIP, alert.Message)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	body, _ := json.Marshal(payload)
	resp, err := m.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func (m *Manager) sendEmail(alert Alert) error {
	cfg := m.config.Email
	if cfg.SMTPHost == "" {
		return fmt.Errorf("email SMTP host not configured")
	}

	// SMTP sending placeholder — use agentmail skill for full email integration
	// Structure preserved for future net/smtp or agentmail API integration
	return fmt.Errorf("email alerts: use agentmail skill for SMTP relay")
}