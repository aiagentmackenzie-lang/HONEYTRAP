package services

import (
	"context"
	"net"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
)

type EventRecorder interface {
	Event(ctx context.Context, session models.Session, eventType string, payload map[string]any) error
}

type SessionContext struct {
	Context  context.Context
	Session  models.Session
	Conn     net.Conn
	Recorder EventRecorder
	Deadline time.Duration
}

type PacketContext struct {
	Context    context.Context
	Session    models.Session
	Service    string
	RemoteAddr net.Addr
	Payload    []byte
	Recorder   EventRecorder
	Write      func([]byte) error
}

type Service interface {
	Name() string
	HandleConn(*SessionContext) error
	HandlePacket(*PacketContext) error
}
