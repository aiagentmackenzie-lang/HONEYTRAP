package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/storage"
)

type SessionManager struct {
	repo storage.Repository
}

func NewSessionManager(repo storage.Repository) *SessionManager {
	return &SessionManager{repo: repo}
}

func (s *SessionManager) Open(ctx context.Context, service, protocol string, remote net.Addr, metadata map[string]any) (models.Session, error) {
	session := models.Session{
		ID:         newID("ses"),
		Service:    service,
		Protocol:   protocol,
		RemoteAddr: remote.String(),
		RemoteIP:   extractIP(remote.String()),
		StartedAt:  time.Now().UTC(),
		Metadata:   metadata,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return models.Session{}, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

func (s *SessionManager) Close(ctx context.Context, sessionID string) error {
	if err := s.repo.CloseSession(ctx, sessionID); err != nil {
		return fmt.Errorf("close session: %w", err)
	}
	return nil
}

func (s *SessionManager) Event(ctx context.Context, session models.Session, eventType string, payload map[string]any) error {
	event := models.Event{
		ID:         newID("evt"),
		SessionID:  session.ID,
		Service:    session.Service,
		Type:       eventType,
		RemoteAddr: session.RemoteAddr,
		Payload:    payload,
		OccurredAt: time.Now().UTC(),
	}
	if err := s.repo.RecordEvent(ctx, event); err != nil {
		return fmt.Errorf("record event: %w", err)
	}
	return nil
}

func newID(prefix string) string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(raw[:])
}

func extractIP(remote string) string {
	host, _, err := net.SplitHostPort(remote)
	if err == nil {
		return host
	}
	return strings.Trim(remote, "[]")
}
