package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/ai"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/config"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/services"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/storage"
)

const defaultMaxSessionsPerService = 500

// udpSessionTTL is how long to reuse an existing UDP session before creating a new one.
const udpSessionTTL = 5 * time.Minute

type udpSessionKey struct {
	service string
	ip      string
}

type udpSessionEntry struct {
	sessionID string
	lastSeen time.Time
}

type Engine struct {
	cfg          config.Config
	repo         storage.Repository
	sessions     *SessionManager
	services     map[string]services.Service
	aiResponder  *ai.Responder
	listeners    []io.Closer
	packetConns  []io.Closer
	activeConns  sync.WaitGroup
	connCount    map[string]*atomic.Int64
	maxSessions  map[string]int64
	shutdownOnce sync.Once

	// mu protects listeners and packetConns from concurrent append/read
	mu sync.Mutex

	// udpSessions tracks active UDP sessions for deduplication
	udpSessions sync.Map // udpSessionKey -> udpSessionEntry
}

func New(cfg config.Config, repo storage.Repository) *Engine {
	connCount := make(map[string]*atomic.Int64)
	maxSessions := make(map[string]int64)
	servicesMap := map[string]services.Service{
		"ssh":           services.NewSSHService(),
		"ssh-enhanced":  services.NewEnhancedSSHService(),
		"http":          services.NewHTTPService(),
		"http-enhanced": services.NewEnhancedHTTPService(),
		"ftp":           services.NewFTPService(),
		"redis":         services.NewRedisService(),
		"udp-decoy":    services.NewUDPDecoyService(),
	}
	for _, svc := range cfg.Services {
		connCount[svc.Name] = &atomic.Int64{}
		limit := int64(defaultMaxSessionsPerService)
		maxSessions[svc.Name] = limit
	}

	// Create AI responder if emulator URL is configured
	var aiResponder *ai.Responder
	aiURL := os.Getenv("HONEYTRAP_AI_URL")
	if aiURL == "" {
		aiURL = "http://localhost:8443"
	}
	aiClient := ai.NewClient(aiURL)
	aiResponder = ai.NewResponder(aiClient)

	// Check AI availability in background (non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := aiClient.Health(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "honeytrap: AI emulator unavailable at %s (static responses only): %v\n", aiURL, err)
		} else {
			aiResponder.MarkAvailable()
			fmt.Fprintf(os.Stderr, "honeytrap: AI emulator connected at %s\n", aiURL)
		}
	}()

	return &Engine{
		cfg:         cfg,
		repo:        repo,
		sessions:    NewSessionManager(repo),
		services:    servicesMap,
		connCount:   connCount,
		maxSessions: maxSessions,
	}
}

func (e *Engine) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(e.cfg.ActiveServices()))

	for _, svcCfg := range e.cfg.ActiveServices() {
		service, ok := e.services[svcCfg.Name]
		if !ok {
			return fmt.Errorf("service %s is not registered", svcCfg.Name)
		}

		wg.Add(1)
		go func(cfg config.ServiceConfig, service services.Service) {
			defer wg.Done()
			switch cfg.Protocol {
			case "tcp":
				errCh <- e.serveTCP(ctx, cfg, service)
			case "udp":
				errCh <- e.serveUDP(ctx, cfg, service)
			default:
				errCh <- fmt.Errorf("unsupported protocol %s", cfg.Protocol)
			}
		}(svcCfg, service)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errCount int
	for err := range errCh {
		if err != nil {
			errCount++
			fmt.Fprintf(os.Stderr, "honeytrap: service error: %v\n", err)
		}
	}

	// Wait for in-flight connections to drain (up to 10 seconds)
	drainDone := make(chan struct{})
	go func() {
		e.activeConns.Wait()
		close(drainDone)
	}()

	select {
	case <-drainDone:
		// All connections drained cleanly
	case <-time.After(10 * time.Second):
		fmt.Fprintln(os.Stderr, "honeytrap: drain timeout, forcing shutdown")
	}

	if errCount > 0 {
		return fmt.Errorf("%d service(s) failed", errCount)
	}
	return nil
}

// Shutdown gracefully stops all listeners and waits for connections to drain.
func (e *Engine) Shutdown(ctx context.Context) error {
	var err error
	e.shutdownOnce.Do(func() {
		e.mu.Lock()
		listeners := make([]io.Closer, len(e.listeners))
		copy(listeners, e.listeners)
		packetConns := make([]io.Closer, len(e.packetConns))
		copy(packetConns, e.packetConns)
		e.mu.Unlock()

		// Close all listeners first (stop accepting new connections)
		for _, l := range listeners {
			_ = l.Close()
		}
		for _, pc := range packetConns {
			_ = pc.Close()
		}

		// Wait for in-flight connections with timeout
		drainDone := make(chan struct{})
		go func() {
			e.activeConns.Wait()
			close(drainDone)
		}()

		select {
		case <-drainDone:
			err = nil
		case <-ctx.Done():
			err = ctx.Err()
		case <-time.After(15 * time.Second):
			err = errors.New("drain timeout: some connections did not close")
		}
	})
	return err
}

func (e *Engine) Status() []models.ServiceStatus {
	statuses := make([]models.ServiceStatus, 0, len(e.cfg.Services))
	for _, svc := range e.cfg.Services {
		status := models.ServiceStatus{
			Name:     svc.Name,
			Protocol: svc.Protocol,
			Address:  svc.Address,
			Enabled:  svc.Enabled,
		}
		if counter, ok := e.connCount[svc.Name]; ok {
			status.ActiveConns = counter.Load()
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func (e *Engine) Repository() storage.Repository {
	return e.repo
}

func (e *Engine) canAccept(serviceName string) bool {
	counter, ok := e.connCount[serviceName]
	if !ok {
		return true
	}
	limit, ok := e.maxSessions[serviceName]
	if !ok {
		return true
	}
	return counter.Load() < limit
}

func (e *Engine) serveTCP(ctx context.Context, cfg config.ServiceConfig, service services.Service) error {
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("listen %s %s: %w", cfg.Name, cfg.Address, err)
	}
	e.mu.Lock()
	e.listeners = append(e.listeners, listener)
	e.mu.Unlock()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			continue
		}

		// Rate limiting: reject if over max sessions
		if !e.canAccept(cfg.Name) {
			_ = conn.Close()
			continue
		}

		e.activeConns.Add(1)
		if counter, ok := e.connCount[cfg.Name]; ok {
			counter.Add(1)
		}
		go e.handleTCP(ctx, cfg, service, conn)
	}
}

func (e *Engine) handleTCP(ctx context.Context, cfg config.ServiceConfig, service services.Service, conn net.Conn) {
	defer func() {
		conn.Close()
		e.activeConns.Done()
		if counter, ok := e.connCount[cfg.Name]; ok {
			counter.Add(-1)
		}
	}()

	session, err := e.sessions.Open(ctx, cfg.Name, cfg.Protocol, conn.RemoteAddr(), map[string]any{"listener": cfg.Address})
	if err != nil {
		return
	}
	defer func() { _ = e.sessions.Close(context.Background(), session.ID) }()

	_ = e.sessions.Event(ctx, session, "session.opened", map[string]any{"service": cfg.Name})
	err = service.HandleConn(&services.SessionContext{
		Context:  ctx,
		Session:  session,
		Conn:     conn,
		Recorder: e.sessions,
		Deadline: 30 * time.Second,
		AI:       e.aiResponder,
	})
	if err != nil {
		_ = e.sessions.Event(ctx, session, "session.error", map[string]any{"error": err.Error()})
		return
	}
	_ = e.sessions.Event(ctx, session, "session.closed", map[string]any{"service": cfg.Name})
}

func (e *Engine) serveUDP(ctx context.Context, cfg config.ServiceConfig, service services.Service) error {
	packetConn, err := net.ListenPacket("udp", cfg.Address)
	if err != nil {
		return fmt.Errorf("listen %s %s: %w", cfg.Name, cfg.Address, err)
	}
	e.mu.Lock()
	e.packetConns = append(e.packetConns, packetConn)
	e.mu.Unlock()

	go func() {
		<-ctx.Done()
		_ = packetConn.Close()
	}()

	buffer := make([]byte, 2048)
	for {
		n, addr, err := packetConn.ReadFrom(buffer)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			continue
		}

		payload := append([]byte(nil), buffer[:n]...)
		go e.handleUDP(ctx, cfg, service, packetConn, addr, payload)
	}
}

func (e *Engine) handleUDP(ctx context.Context, cfg config.ServiceConfig, service services.Service, packetConn net.PacketConn, addr net.Addr, payload []byte) {
	ip := extractIPFromAddr(addr)
	key := udpSessionKey{service: cfg.Name, ip: ip}

	// Check for existing session to deduplicate
	var session models.Session
	now := time.Now()
	if entry, ok := e.udpSessions.Load(key); ok {
		entry := entry.(udpSessionEntry)
		if now.Sub(entry.lastSeen) < udpSessionTTL {
			// Reuse existing session
			session = models.Session{ID: entry.sessionID, Service: cfg.Name, Protocol: "udp", RemoteAddr: addr.String(), RemoteIP: ip}
			_ = e.sessions.Event(ctx, session, "udp.datagram", map[string]any{"size": len(payload), "payload_preview": string(payload)})
			e.udpSessions.Store(key, udpSessionEntry{sessionID: entry.sessionID, lastSeen: now})
			_ = service.HandlePacket(&services.PacketContext{
				Context:    ctx,
				Service:    cfg.Name,
				RemoteAddr: addr,
				Payload:    payload,
				Recorder:   e.sessions,
				Write: func(data []byte) error {
					_, err := packetConn.WriteTo(data, addr)
					return err
				},
			})
			return
		}
	}

	// No active session or expired — create new one
	var err error
	session, err = e.sessions.Open(ctx, cfg.Name, "udp", addr, map[string]any{"listener": cfg.Address, "size": len(payload)})
	if err != nil {
		return
	}

	e.udpSessions.Store(key, udpSessionEntry{sessionID: session.ID, lastSeen: now})

	_ = e.sessions.Event(ctx, session, "udp.received", map[string]any{"size": len(payload), "payload_preview": string(payload)})
	_ = service.HandlePacket(&services.PacketContext{
		Context:    ctx,
		Service:    cfg.Name,
		RemoteAddr: addr,
		Payload:    payload,
		Recorder:   e.sessions,
		Write: func(data []byte) error {
			_, err := packetConn.WriteTo(data, addr)
			return err
		},
	})
	_ = e.sessions.Event(ctx, session, "session.closed", map[string]any{"service": cfg.Name})
}

func extractIPFromAddr(addr net.Addr) string {
	host, _, err := net.SplitHostPort(addr.String())
	if err == nil {
		return host
	}
	return addr.String()
}