package engine

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/config"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/services"
)

func (e *Engine) serveTCP(ctx context.Context, cfg config.ServiceConfig, service services.Service) error {
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("listen %s %s: %w", cfg.Name, cfg.Address, err)
	}
	e.listeners = append(e.listeners, listener)

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

		go e.handleTCP(ctx, cfg, service, conn)
	}
}

func (e *Engine) handleTCP(ctx context.Context, cfg config.ServiceConfig, service services.Service, conn net.Conn) {
	defer conn.Close()

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
	e.packetConns = append(e.packetConns, packetConn)

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
	session, err := e.sessions.Open(ctx, cfg.Name, cfg.Protocol, addr, map[string]any{"listener": cfg.Address, "size": len(payload)})
	if err != nil {
		return
	}
	defer func() { _ = e.sessions.Close(context.Background(), session.ID) }()

	_ = e.sessions.Event(ctx, session, "udp.received", map[string]any{"size": len(payload), "payload_preview": string(payload)})
	_ = service.HandlePacket(&services.PacketContext{
		Context:    ctx,
		Session:    session,
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
