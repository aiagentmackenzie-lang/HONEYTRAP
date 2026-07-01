package engine

import (
	"context"
	"testing"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/config"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/storage"
)

func TestNew(t *testing.T) {
	cfg := config.Config{
		NodeName:    "test-node",
		Environment: "test",
		DataDir:     t.TempDir(),
		Services: []config.ServiceConfig{
			{Name: "ssh", Protocol: "tcp", Address: ":0", Enabled: true},
		},
	}

	repo, err := storage.NewMemoryRepository(cfg.SessionLogPath(), cfg.EventLogPath())
	if err != nil {
		t.Fatalf("NewMemoryRepository() error: %v", err)
	}

	e := New(cfg, repo)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestStatus(t *testing.T) {
	cfg := config.Config{
		NodeName:    "test-node",
		Environment: "test",
		DataDir:     t.TempDir(),
		Services: []config.ServiceConfig{
			{Name: "ssh", Protocol: "tcp", Address: ":2222", Enabled: true},
			{Name: "http", Protocol: "tcp", Address: ":8080", Enabled: false},
		},
	}

	repo, _ := storage.NewMemoryRepository(cfg.SessionLogPath(), cfg.EventLogPath())
	e := New(cfg, repo)

	statuses := e.Status()
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}

	found := make(map[string]bool)
	for _, s := range statuses {
		found[s.Name] = s.Enabled
	}
	if !found["ssh"] {
		t.Error("expected ssh to be enabled")
	}
	if found["http"] {
		t.Error("expected http to be disabled")
	}
}

func TestCanAccept(t *testing.T) {
	cfg := config.Config{
		DataDir: t.TempDir(),
		Services: []config.ServiceConfig{
			{Name: "ssh", Protocol: "tcp", Address: ":0", Enabled: true},
		},
	}

	repo, _ := storage.NewMemoryRepository(cfg.SessionLogPath(), cfg.EventLogPath())
	e := New(cfg, repo)

	if !e.canAccept("ssh") {
		t.Error("expected canAccept to return true with no connections")
	}
	if !e.canAccept("unknown") {
		t.Error("expected canAccept to return true for unknown services")
	}
}

func TestRepository(t *testing.T) {
	cfg := config.Config{
		DataDir: t.TempDir(),
		Services: []config.ServiceConfig{
			{Name: "ssh", Protocol: "tcp", Address: ":0", Enabled: true},
		},
	}

	repo, _ := storage.NewMemoryRepository(cfg.SessionLogPath(), cfg.EventLogPath())
	e := New(cfg, repo)

	if e.Repository() == nil {
		t.Error("expected non-nil repository")
	}
}

func TestShutdown(t *testing.T) {
	cfg := config.Config{
		DataDir: t.TempDir(),
		Services: []config.ServiceConfig{
			{Name: "ssh", Protocol: "tcp", Address: ":0", Enabled: true},
		},
	}

	repo, _ := storage.NewMemoryRepository(cfg.SessionLogPath(), cfg.EventLogPath())
	e := New(cfg, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := e.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() error: %v", err)
	}

	// Second shutdown should be no-op (sync.Once)
	err = e.Shutdown(ctx)
	if err != nil {
		t.Errorf("Second Shutdown() error: %v", err)
	}
}
