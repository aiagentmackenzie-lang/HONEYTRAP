package analysis

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
)

// IsScripted detects automated tool usage by analyzing command timing patterns.
// Scripted attacks show: rapid uniform intervals, no typos, no exploration pauses.
func IsScripted(events []models.Event) bool {
	if len(events) < 3 {
		return false
	}

	// Calculate inter-event intervals
	intervals := eventIntervals(events)
	if len(intervals) == 0 {
		return false
	}

	// Check for uniform timing (low coefficient of variation = scripted)
	mean := meanFloat(intervals)
	if mean == 0 {
		return true // All events at same time = scripted
	}

	cv := stdDev(intervals) / mean // Coefficient of variation

	// Scripts typically have CV < 0.3 (very uniform timing)
	// Humans have CV > 0.5 (highly variable timing)
	return cv < 0.3
}

// IsHuman detects human-like behavior: variable timing, typos, exploration.
func IsHuman(events []models.Event) bool {
	if len(events) < 3 {
		return true // Single events could be human
	}

	intervals := eventIntervals(events)
	if len(intervals) == 0 {
		return false
	}

	mean := meanFloat(intervals)
	if mean == 0 {
		return false
	}

	cv := stdDev(intervals) / mean

	// Human behavior has high variability
	if cv < 0.3 {
		return false
	}

	// Check for long pauses (thinking time)
	for _, interval := range intervals {
		if interval > 30 && interval < 600 { // 30s - 10min pause = human thinking
			return true
		}
	}

	return cv > 0.5
}

// ClassifyTool attempts to identify the attack tool based on command patterns.
func ClassifyTool(events []models.Event) string {
	signatures := map[string][]string{
		"nmap":       {"nmap", "-sv", "-ss", "-o ", "--script", "-p-"},
		"hydra":      {"hydra", "-l ", "-p ", "-t ", "password"},
		"metasploit": {"msfconsole", "use exploit", "set payload", "meterpreter", "sessions -l"},
		"nikto":      {"nikto", "-h ", "-c ", "all"},
		"sqlmap":     {"sqlmap", "--dbs", "--tables", "--dump", "--batch"},
		"curl":       {"curl", "-x ", "-h ", "--data", "-o "},
		"wget":       {"wget", "-o ", "--no-check-certificate"},
		"nuclei":     {"nuclei", "-t ", "-severity", "-target"},
	}

	commandText := ""
	for _, e := range events {
		// Extract data from Payload map if present
		if data, ok := e.Payload["data"].(string); ok {
			commandText += " " + strings.ToLower(data)
		}
		if cmd, ok := e.Payload["command"].(string); ok {
			commandText += " " + strings.ToLower(cmd)
		}
		// Also check Type as fallback
		commandText += " " + strings.ToLower(e.Type)
	}

	bestMatch := "unknown"
	bestScore := 0

	for tool, patterns := range signatures {
		score := 0
		for _, pattern := range patterns {
			if strings.Contains(commandText, pattern) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestMatch = tool
		}
	}

	if bestScore >= 2 {
		return bestMatch
	}

	return "custom"
}

// RiskScore calculates a 0-1 risk score for a session based on multiple factors.
func RiskScore(session models.Session, events []models.Event) float64 {
	if len(events) == 0 {
		return 0.1 // Minimal risk with no events
	}

	var score float64

	// Factor 1: Number of events (more = more activity = higher risk)
	eventFactor := math.Min(float64(len(events))/50.0, 1.0) * 0.2
	score += eventFactor

	// Factor 2: Tool classification
	tool := ClassifyTool(events)
	toolWeights := map[string]float64{
		"metasploit": 0.25,
		"hydra":      0.25,
		"nmap":       0.15,
		"sqlmap":     0.20,
		"nuclei":     0.15,
		"nikto":      0.10,
		"custom":     0.20,
		"unknown":    0.05,
		"curl":       0.05,
		"wget":       0.05,
	}
	if w, ok := toolWeights[tool]; ok {
		score += w
	}

	// Factor 3: Scripted attacks are often more dangerous (automated)
	if IsScripted(events) {
		score += 0.15
	}

	// Factor 4: Session duration (longer = more persistent = higher risk)
	durationMin := time.Since(session.StartedAt).Minutes()
	if session.EndedAt != nil {
		durationMin = session.EndedAt.Sub(session.StartedAt).Minutes()
	}
	durationFactor := math.Min(durationMin/60.0, 1.0) * 0.1
	score += durationFactor

	// Factor 5: Dangerous commands
	commandText := ""
	for _, e := range events {
		if data, ok := e.Payload["data"].(string); ok {
			commandText += " " + strings.ToLower(data)
		}
		if cmd, ok := e.Payload["command"].(string); ok {
			commandText += " " + strings.ToLower(cmd)
		}
	}
	dangerPatterns := []string{"rm -rf", "chmod 777", "/etc/shadow", "passwd", "sudo", "wget", "curl", "reverse shell", "/bin/bash -i"}
	dangerCount := 0
	for _, pattern := range dangerPatterns {
		if strings.Contains(commandText, pattern) {
			dangerCount++
		}
	}
	dangerFactor := math.Min(float64(dangerCount)/3.0, 1.0) * 0.2
	score += dangerFactor

	// Factor 6: Login attempts
	loginCount := 0
	for _, e := range events {
		if strings.ToLower(e.Type) == "login" {
			loginCount++
		}
	}
	loginFactor := math.Min(float64(loginCount)/5.0, 1.0) * 0.1
	score += loginFactor

	// Clamp to [0, 1]
	return math.Max(0, math.Min(1, score))
}

// --- Helper functions ---

func eventIntervals(events []models.Event) []float64 {
	intervals := make([]float64, 0, len(events)-1)
	for i := 1; i < len(events); i++ {
		d := events[i].OccurredAt.Sub(events[i-1].OccurredAt).Seconds()
		intervals = append(intervals, d)
	}
	return intervals
}

func meanFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func stdDev(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	m := meanFloat(vals)
	sum := 0.0
	for _, v := range vals {
		diff := v - m
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(vals)))
}

// SortedIntervals returns sorted inter-event intervals for analysis.
func SortedIntervals(events []models.Event) []float64 {
	intervals := eventIntervals(events)
	sort.Float64s(intervals)
	return intervals
}