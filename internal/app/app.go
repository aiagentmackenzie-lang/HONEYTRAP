package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/cli"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/config"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/engine"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/storage"
)

type App struct {
	runner *cli.Runner
	engine *engine.Engine
	cfg    config.Config
}

func New(profileName string) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if profileName != "" {
		profile, err := config.LoadProfile(profileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "honeytrap: warning: %v\n", err)
		} else {
			cfg = *config.ApplyProfile(&cfg, profile)
		}
	}

	repo, err := storage.NewMemoryRepository(cfg.SessionLogPath(), cfg.EventLogPath())
	if err != nil {
		return nil, err
	}

	core := engine.New(cfg, repo)
	return &App{runner: cli.NewRunner(core), engine: core, cfg: cfg}, nil
}

func (a *App) Run(ctx context.Context, args []string) error {
	// Start JSONL ingestion bridge if API URL is configured
	apiURL := os.Getenv("HONEYTRAP_API_URL")
	apiToken := os.Getenv("API_TOKEN")

	// Also try the default API URL
	if apiURL == "" {
		apiURL = "http://localhost:3000"
	}

	go a.ingestLoop(ctx, apiURL, apiToken, a.cfg.SessionLogPath(), a.cfg.EventLogPath())

	return a.runner.Run(ctx, args)
}

// ingestLoop periodically reads new entries from JSONL files and POSTs them to the API.
// It tracks the byte offset of each file to avoid re-ingesting old entries.
func (a *App) ingestLoop(ctx context.Context, apiURL, apiToken, sessionPath, eventPath string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	sessionOffset := int64(0)
	eventOffset := int64(0)

	// Initialize offsets to end of existing files
	if info, err := os.Stat(sessionPath); err == nil {
		sessionOffset = info.Size()
	}
	if info, err := os.Stat(eventPath); err == nil {
		eventOffset = info.Size()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sessionOffset = ingestFile(ctx, apiURL, apiToken, sessionPath, sessionOffset, "sessions")
			eventOffset = ingestFile(ctx, apiURL, apiToken, eventPath, eventOffset, "events")
		}
	}
}

// ingestFile reads new lines from a JSONL file starting at offset and POSTs them to the API.
// Returns the new offset after ingestion.
func ingestFile(ctx context.Context, apiURL, apiToken, filePath string, offset int64, dataType string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		return offset
	}

	if info.Size() <= offset {
		return offset
	}

	f, err := os.Open(filePath)
	if err != nil {
		return offset
	}
	defer f.Close()

	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return offset
	}

	scanner := bufio.NewScanner(f)
	var records []json.RawMessage

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		records = append(records, json.RawMessage(line))
	}

	if len(records) == 0 {
		return info.Size()
	}

	// POST each record individually to the ingest endpoint
	client := &http.Client{Timeout: 30 * time.Second}
	ingested := 0

	for _, record := range records {
		body := fmt.Sprintf(`{"path":"","type":"%s","record":%s}`, dataType, string(record))

		req, err := http.NewRequestWithContext(ctx, "POST", apiURL+"/ingest-record", strings.NewReader(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if apiToken != "" {
			req.Header.Set("Authorization", "Bearer "+apiToken)
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode < 300 {
			ingested++
		}
	}

	if ingested > 0 {
		fmt.Fprintf(os.Stderr, "honeytrap: ingested %d/%d %s records to API\n", ingested, len(records), dataType)
	}

	return info.Size()
}