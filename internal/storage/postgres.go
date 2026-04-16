package storage

import (
	"context"
	"errors"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
)

var ErrPostgresDriverUnavailable = errors.New("postgres driver unavailable in offline build; use schema.sql with the Fastify API runtime")

type PostgresRepository struct{}

func NewPostgresRepository(string) (*PostgresRepository, error) {
	return nil, ErrPostgresDriverUnavailable
}

func (p *PostgresRepository) CreateSession(context.Context, models.Session) error {
	return ErrPostgresDriverUnavailable
}
func (p *PostgresRepository) CloseSession(context.Context, string) error {
	return ErrPostgresDriverUnavailable
}
func (p *PostgresRepository) RecordEvent(context.Context, models.Event) error {
	return ErrPostgresDriverUnavailable
}
func (p *PostgresRepository) ListSessions(context.Context, int) ([]models.Session, error) {
	return nil, ErrPostgresDriverUnavailable
}
func (p *PostgresRepository) ListEvents(context.Context, int) ([]models.Event, error) {
	return nil, ErrPostgresDriverUnavailable
}
func (p *PostgresRepository) Health(context.Context) error { return ErrPostgresDriverUnavailable }
