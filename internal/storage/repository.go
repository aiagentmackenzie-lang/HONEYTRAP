package storage

import (
	"context"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
)

type Repository interface {
	CreateSession(ctx context.Context, session models.Session) error
	CloseSession(ctx context.Context, sessionID string) error
	RecordEvent(ctx context.Context, event models.Event) error
	ListSessions(ctx context.Context, limit int) ([]models.Session, error)
	ListEvents(ctx context.Context, limit int) ([]models.Event, error)
	Health(ctx context.Context) error
}
