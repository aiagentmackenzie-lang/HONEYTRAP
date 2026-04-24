package models

import "time"

type Session struct {
	ID         string
	Service    string
	Protocol   string
	RemoteAddr string
	RemoteIP   string
	StartedAt  time.Time
	EndedAt    *time.Time
	Metadata   map[string]any
}

type Event struct {
	ID         string
	SessionID  string
	Service    string
	Type       string
	RemoteAddr string
	Payload    map[string]any
	OccurredAt time.Time
}

type Token struct {
	ID              string
	Name            string
	Kind            string
	Value           string
	Description     string
	FirstAccessedAt *time.Time
	LastAccessedAt  *time.Time
	Metadata        map[string]any
}

type ServiceStatus struct {
	Name        string
	Protocol    string
	Address     string
	Enabled     bool
	ActiveConns int64
}
