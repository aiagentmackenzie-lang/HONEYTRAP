package services

import (
	"bufio"
	"context"
	"net"
	"testing"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
)

// mockRecorder captures events for testing
type mockRecorder struct {
	events []string
}

func (m *mockRecorder) Event(_ context.Context, _ models.Session, eventType string, _ map[string]any) error {
	m.events = append(m.events, eventType)
	return nil
}

func TestSSHService_Name(t *testing.T) {
	svc := NewSSHService()
	if svc.Name() != "ssh" {
		t.Errorf("expected name 'ssh', got %q", svc.Name())
	}
}

func TestEnhancedSSHService_Name(t *testing.T) {
	svc := NewEnhancedSSHService()
	if svc.Name() != "ssh-enhanced" {
		t.Errorf("expected name 'ssh-enhanced', got %q", svc.Name())
	}
}

func TestHTTPService_Name(t *testing.T) {
	svc := NewHTTPService()
	if svc.Name() != "http" {
		t.Errorf("expected name 'http', got %q", svc.Name())
	}
}

func TestEnhancedHTTPService_Name(t *testing.T) {
	svc := NewEnhancedHTTPService()
	if svc.Name() != "http-enhanced" {
		t.Errorf("expected name 'http-enhanced', got %q", svc.Name())
	}
}

func TestFTPService_Name(t *testing.T) {
	svc := NewFTPService()
	if svc.Name() != "ftp" {
		t.Errorf("expected name 'ftp', got %q", svc.Name())
	}
}

func TestRedisService_Name(t *testing.T) {
	svc := NewRedisService()
	if svc.Name() != "redis" {
		t.Errorf("expected name 'redis', got %q", svc.Name())
	}
}

func TestUDPDecoyService_Name(t *testing.T) {
	svc := NewUDPDecoyService()
	if svc.Name() != "udp-decoy" {
		t.Errorf("expected name 'udp-decoy', got %q", svc.Name())
	}
}

func TestEnhancedHTTPService_HandleConn_ConnectionClose(t *testing.T) {
	svc := NewEnhancedHTTPService()
	recorder := &mockRecorder{}

	client, server := net.Pipe()

	ctx := context.Background()
	session := models.Session{ID: "test-session", Service: "http-enhanced"}

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Send a GET request with Connection: close and close immediately
		client.Write([]byte("GET / HTTP/1.1\r\nHost: test\r\nConnection: close\r\n\r\n"))
		// Read the response
		reader := bufio.NewReader(client)
		client.SetReadDeadline(time.Now().Add(3 * time.Second))
		reader.ReadString(0) // read until EOF
		client.Close()
	}()

	sessCtx := &SessionContext{
		Context:  ctx,
		Session:  session,
		Conn:     server,
		Recorder: recorder,
		Deadline: 5 * time.Second,
		AI:       nil,
	}

	_ = svc.HandleConn(sessCtx)
	<-done

	// Verify at least one event was recorded (http.request)
	if len(recorder.events) == 0 {
		t.Error("expected at least one event to be recorded")
	}
}

func TestRedisService_HandleConn_QUIT(t *testing.T) {
	svc := NewRedisService()
	recorder := &mockRecorder{}

	client, server := net.Pipe()

	ctx := context.Background()
	session := models.Session{ID: "test-redis", Service: "redis"}

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Send QUIT command
		client.Write([]byte("QUIT\r\n"))
		reader := bufio.NewReader(client)
		client.SetReadDeadline(time.Now().Add(3 * time.Second))
		reader.ReadString('\n')
		client.Close()
	}()

	sessCtx := &SessionContext{
		Context:  ctx,
		Session:  session,
		Conn:     server,
		Recorder: recorder,
		Deadline: 5 * time.Second,
		AI:       nil,
	}

	_ = svc.HandleConn(sessCtx)
	<-done

	if len(recorder.events) == 0 {
		t.Error("expected at least one event to be recorded")
	}
}