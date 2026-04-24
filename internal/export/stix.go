package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"

	"github.com/google/uuid"
)

// STIXBundle represents a STIX 2.1 bundle for threat intel sharing.
type STIXBundle struct {
	Type    string        `json:"type"`
	ID      string        `json:"id"`
	Objects []STIXObject  `json:"objects"`
}

// STIXObject is a generic STIX 2.1 object.
type STIXObject struct {
	Type        string                 `json:"type"`
	SpecVersion string                 `json:"spec_version"`
	ID          string                 `json:"id"`
	Created     string                 `json:"created"`
	Modified    string                 `json:"modified"`
	Name        string                 `json:"name,omitempty"`
	Value       string                 `json:"value,omitempty"`
	Labels      []string               `json:"labels,omitempty"`
	Confidence  int                    `json:"confidence,omitempty"`
	Pattern     string                 `json:"pattern,omitempty"`
	PatternType string                 `json:"pattern_type,omitempty"`
	ValidFrom   string                 `json:"valid_from,omitempty"`
	Description string                 `json:"description,omitempty"`
	SourceIP    string                 `json:"source_ip,omitempty"`
	Extensions  map[string]interface{} `json:"extensions,omitempty"`
}

// STIXExporter handles STIX/TAXII export of honeypot data.
type STIXExporter struct {
	OutputDir string
}

// NewSTIXExporter creates a new STIX exporter.
func NewSTIXExporter(outputDir string) *STIXExporter {
	return &STIXExporter{OutputDir: outputDir}
}

// ExportSessions exports session data as a STIX 2.1 bundle.
func (e *STIXExporter) ExportSessions(sessions []models.Session) (string, error) {
	bundle := STIXBundle{
		Type:    "bundle",
		ID:      fmt.Sprintf("bundle--%s", generateID()),
		Objects: make([]STIXObject, 0, len(sessions)*2+1),
	}

	// Identity object for HONEYTRAP
	bundle.Objects = append(bundle.Objects, STIXObject{
		Type:        "identity",
		SpecVersion: "2.1",
		ID:          "identity--a3e3b5c4-7c8d-4e9f-b0a1-2d3e4f5a6b7c",
		Created:     time.Now().UTC().Format(time.RFC3339),
		Modified:    time.Now().UTC().Format(time.RFC3339),
		Name:        "HONEYTRAP Deception Framework",
		Description: "AI-Powered Deception Framework — honeypot system",
		Labels:      []string{"honeypot", "deception"},
	})

	for _, s := range sessions {
		ts := s.StartedAt.UTC().Format(time.RFC3339)

		// IPv4 address object for attacker
		ipObj := STIXObject{
			Type:        "ipv4-addr",
			SpecVersion: "2.1",
			ID:          fmt.Sprintf("ipv4-addr--%s", generateID()),
			Created:     ts,
			Modified:    ts,
			Value:       s.RemoteIP,
		}
		bundle.Objects = append(bundle.Objects, ipObj)

		// Observed data for the session
		obsObj := STIXObject{
			Type:        "observed-data",
			SpecVersion: "2.1",
			ID:          fmt.Sprintf("observed-data--%s", generateID()),
			Created:     ts,
			Modified:    ts,
			Labels:      []string{"honeypot-session", s.Service},
			Confidence:  90,
			Description: fmt.Sprintf("Honeypot session on %s service from %s", s.Service, s.RemoteIP),
			Extensions: map[string]interface{}{
				"honeytrap-session": map[string]interface{}{
					"session_id":  s.ID,
					"service":     s.Service,
					"protocol":    s.Protocol,
					"remote_addr": s.RemoteAddr,
					"started_at":  s.StartedAt.UTC().Format(time.RFC3339),
				},
			},
		}
		bundle.Objects = append(bundle.Objects, obsObj)
	}

	return e.writeBundle(bundle)
}

// ExportTokens exports token access alerts as STIX indicators.
func (e *STIXExporter) ExportTokens(tokens []models.Token) (string, error) {
	bundle := STIXBundle{
		Type:    "bundle",
		ID:      fmt.Sprintf("bundle--%s", generateID()),
		Objects: make([]STIXObject, 0, len(tokens)+1),
	}

	// Identity
	bundle.Objects = append(bundle.Objects, STIXObject{
		Type:        "identity",
		SpecVersion: "2.1",
		ID:          "identity--a3e3b5c4-7c8d-4e9f-b0a1-2d3e4f5a6b7c",
		Created:     time.Now().UTC().Format(time.RFC3339),
		Modified:    time.Now().UTC().Format(time.RFC3339),
		Name:        "HONEYTRAP Deception Framework",
	})

	for _, t := range tokens {
		ts := time.Now().UTC().Format(time.RFC3339)
		if t.FirstAccessedAt != nil {
			ts = t.FirstAccessedAt.UTC().Format(time.RFC3339)
		}

		indicator := STIXObject{
			Type:        "indicator",
			SpecVersion: "2.1",
			ID:          fmt.Sprintf("indicator--%s", generateID()),
			Created:     ts,
			Modified:    ts,
			Name:        fmt.Sprintf("Honeytoken Access: %s (%s)", t.Name, t.Kind),
			Labels:      []string{"honeytoken-access", t.Kind},
			Confidence:  95,
			Pattern:     fmt.Sprintf("[file:name = '%s']", t.Value),
			PatternType: "stix",
			ValidFrom:   ts,
			Description: fmt.Sprintf("Honeytoken of kind %s was accessed. Value: %s", t.Kind, t.Value[:min(len(t.Value), 20)]),
		}
		bundle.Objects = append(bundle.Objects, indicator)
	}

	return e.writeBundle(bundle)
}

func (e *STIXExporter) writeBundle(bundle STIXBundle) (string, error) {
	if err := os.MkdirAll(e.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("cannot create output dir: %w", err)
	}

	ts := time.Now().UTC().Format("20060102_150405")
	filename := fmt.Sprintf("honeytrap_stix_%s.json", ts)
	path := filepath.Join(e.OutputDir, filename)

	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON marshal failed: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write failed: %w", err)
	}

	return path, nil
}

func generateID() string {
	return uuid.New().String()
}
