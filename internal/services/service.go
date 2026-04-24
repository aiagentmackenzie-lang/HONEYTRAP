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

// AIResponder is implemented by the AI client. Services can optionally use it
// for dynamic response generation, falling back to static responses if nil.
type AIResponder interface {
	Generate(ctx context.Context, service string, contextData map[string]any) (string, error)
	IsAvailable() bool
}

type SessionContext struct {
	Context  context.Context
	Session  models.Session
	Conn     net.Conn
	Recorder EventRecorder
	Deadline time.Duration
	AI       AIResponder // nil when AI emulation is disabled
}

type PacketContext struct {
	Context    context.Context
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

// BaseService provides a no-op base that embedded services can compose.
type BaseService struct{}