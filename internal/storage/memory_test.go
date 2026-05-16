package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
)

func TestMemoryRepository_CreateAndListSessions(t *testing.T) {
	repo, err := NewMemoryRepository(t.TempDir()+"/sessions.jsonl", t.TempDir()+"/events.jsonl")
	if err != nil {
		t.Fatalf("NewMemoryRepository() error: %v", err)
	}

	session := models.Session{
		ID:         "test-1",
		Service:    "ssh",
		Protocol:   "tcp",
		RemoteAddr: "10.0.0.1:43210",
		RemoteIP:   "10.0.0.1",
	}

	err = repo.CreateSession(context.Background(), session)
	if err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	sessions, err := repo.ListSessions(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListSessions() error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != "test-1" {
		t.Errorf("expected session ID 'test-1', got %q", sessions[0].ID)
	}
}

func TestMemoryRepository_CloseSession(t *testing.T) {
	repo, _ := NewMemoryRepository(t.TempDir()+"/sessions.jsonl", t.TempDir()+"/events.jsonl")

	session := models.Session{
		ID:      "test-close-1",
		Service: "ssh",
	}

	repo.CreateSession(context.Background(), session)

	err := repo.CloseSession(context.Background(), "test-close-1")
	if err != nil {
		t.Fatalf("CloseSession() error: %v", err)
	}

	sessions, _ := repo.ListSessions(context.Background(), 10)
	if sessions[0].EndedAt == nil {
		t.Error("expected EndedAt to be set after close")
	}
}

func TestMemoryRepository_RecordAndListEvents(t *testing.T) {
	repo, _ := NewMemoryRepository(t.TempDir()+"/sessions.jsonl", t.TempDir()+"/events.jsonl")

	event := models.Event{
		ID:         "evt-1",
		SessionID:  "ses-1",
		Service:    "ssh",
		Type:       "command",
		RemoteAddr: "10.0.0.1:43210",
		Payload:    map[string]any{"data": "whoami"},
	}

	err := repo.RecordEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("RecordEvent() error: %v", err)
	}

	events, err := repo.ListEvents(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListEvents() error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != "command" {
		t.Errorf("expected event type 'command', got %q", events[0].Type)
	}
}

func TestMemoryRepository_EvictionCap(t *testing.T) {
	repo, _ := NewMemoryRepository(t.TempDir()+"/sessions.jsonl", t.TempDir()+"/events.jsonl")

	// Add more than MaxEventsInMemory events
	for i := 0; i < MaxEventsInMemory+100; i++ {
		event := models.Event{
			ID:        fmt.Sprintf("evt-%d", i),
			SessionID: "ses-1",
			Type:      "test",
		}
		repo.RecordEvent(context.Background(), event)
	}

	events, _ := repo.ListEvents(context.Background(), MaxEventsInMemory+200)
	if len(events) > MaxEventsInMemory {
		t.Errorf("expected at most %d events after eviction, got %d", MaxEventsInMemory, len(events))
	}
}

func TestMemoryRepository_Health(t *testing.T) {
	repo, _ := NewMemoryRepository(t.TempDir()+"/sessions.jsonl", t.TempDir()+"/events.jsonl")
	if err := repo.Health(context.Background()); err != nil {
		t.Errorf("Health() error: %v", err)
	}
}