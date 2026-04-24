package ai

import (
	"context"
	"fmt"
	"sync"
)

// Responder implements services.AIResponder using the AI emulator service.
// It provides dynamic response generation with fallback to static responses
// when the AI service is unavailable.
type Responder struct {
	client    *Client
	available bool
	mu        sync.RWMutex
}

// NewResponder creates an AI responder that wraps the AI emulator client.
func NewResponder(client *Client) *Responder {
	return &Responder{
		client:    client,
		available: client != nil,
	}
}

// Generate sends a request to the AI emulator and returns the dynamic response.
// Falls back to an error if the AI service is unavailable.
func (r *Responder) Generate(ctx context.Context, service string, contextData map[string]any) (string, error) {
	r.mu.RLock()
	if !r.available || r.client == nil {
		r.mu.RUnlock()
		return "", fmt.Errorf("AI emulation unavailable")
	}
	r.mu.RUnlock()

	resp, err := r.client.Generate(ctx, EmulationRequest{
		Service:  service,
		Context:  contextData,
	})
	if err != nil {
		r.mu.Lock()
		r.available = false
		r.mu.Unlock()
		return "", fmt.Errorf("AI emulation failed: %w", err)
	}

	return resp.Response, nil
}

// IsAvailable returns whether the AI service is currently reachable.
func (r *Responder) IsAvailable() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.available
}

// MarkAvailable marks the AI service as available (e.g., after a health check succeeds).
func (r *Responder) MarkAvailable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.available = true
}