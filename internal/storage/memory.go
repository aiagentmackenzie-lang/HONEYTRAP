package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
)

const (
	// MaxEventsInMemory caps the in-memory event list to prevent unbounded growth.
	// Oldest events are evicted when the cap is hit.
	MaxEventsInMemory = 10000
	// MaxSessionsInMemory caps the in-memory session list.
	MaxSessionsInMemory = 5000
)

type MemoryRepository struct {
	mu          sync.RWMutex
	sessions    []models.Session
	events      []models.Event
	sessionPath string
	eventPath   string
}

func NewMemoryRepository(sessionPath, eventPath string) (*MemoryRepository, error) {
	repo := &MemoryRepository{
		sessions:    make([]models.Session, 0, 128),
		events:      make([]models.Event, 0, 512),
		sessionPath: sessionPath,
		eventPath:   eventPath,
	}

	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *MemoryRepository) CreateSession(_ context.Context, session models.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions = append(r.sessions, session)

	// Evict oldest sessions if over cap
	if len(r.sessions) > MaxSessionsInMemory {
		r.sessions = r.sessions[len(r.sessions)-MaxSessionsInMemory:]
	}

	return appendJSONL(r.sessionPath, session)
}

func (r *MemoryRepository) CloseSession(_ context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	for i := range r.sessions {
		if r.sessions[i].ID == sessionID {
			r.sessions[i].EndedAt = &now
			return rewriteJSONL(r.sessionPath, r.sessions)
		}
	}
	return fmt.Errorf("session %s not found", sessionID)
}

func (r *MemoryRepository) RecordEvent(_ context.Context, event models.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, event)

	// Evict oldest events if over cap
	if len(r.events) > MaxEventsInMemory {
		r.events = r.events[len(r.events)-MaxEventsInMemory:]
	}

	return appendJSONL(r.eventPath, event)
}

func (r *MemoryRepository) ListSessions(_ context.Context, limit int) ([]models.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return reverseSessions(r.sessions, limit), nil
}

func (r *MemoryRepository) ListEvents(_ context.Context, limit int) ([]models.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return reverseEvents(r.events, limit), nil
}

func (r *MemoryRepository) Health(context.Context) error {
	return nil
}

func (r *MemoryRepository) load() error {
	if err := loadJSONL(r.sessionPath, &r.sessions); err != nil {
		return err
	}
	if err := loadJSONL(r.eventPath, &r.events); err != nil {
		return err
	}

	// Trim to caps on load (in case the JSONL files grew beyond limits)
	if len(r.sessions) > MaxSessionsInMemory {
		r.sessions = r.sessions[len(r.sessions)-MaxSessionsInMemory:]
	}
	if len(r.events) > MaxEventsInMemory {
		r.events = r.events[len(r.events)-MaxEventsInMemory:]
	}

	return nil
}

func appendJSONL(path string, value any) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	return nil
}

func rewriteJSONL[T any](path string, values []T) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("rewrite %s: %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, value := range values {
		if err := encoder.Encode(value); err != nil {
			return fmt.Errorf("encode %s: %w", path, err)
		}
	}
	return nil
}

func loadJSONL[T any](path string, target *[]T) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		var item T
		if err := decoder.Decode(&item); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("decode %s: %w", path, err)
		}
		*target = append(*target, item)
	}
	return nil
}

func reverseSessions(source []models.Session, limit int) []models.Session {
	if limit <= 0 || limit > len(source) {
		limit = len(source)
	}
	result := make([]models.Session, 0, limit)
	for i := len(source) - 1; i >= 0 && len(result) < limit; i-- {
		result = append(result, source[i])
	}
	return result
}

func reverseEvents(source []models.Event, limit int) []models.Event {
	if limit <= 0 || limit > len(source) {
		limit = len(source)
	}
	result := make([]models.Event, 0, limit)
	for i := len(source) - 1; i >= 0 && len(result) < limit; i-- {
		result = append(result, source[i])
	}
	return result
}