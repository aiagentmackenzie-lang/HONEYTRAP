package services

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// EnhancedSSHService provides a more realistic SSH honeypot experience.
// It performs banner exchange, captures client capabilities, and simulates
// authentication before disconnecting — keeping attackers engaged longer.
type EnhancedSSHService struct {
	BaseService
	Banner       string
	ServerID     string
	AuthAttempts int
}

func NewEnhancedSSHService() *EnhancedSSHService {
	return &EnhancedSSHService{
		Banner:       "SSH-2.0-OpenSSH_9.3p1 Debian-1",
		ServerID:     "SSH-2.0-OpenSSH_9.3p1",
		AuthAttempts: 3,
	}
}

func (s *EnhancedSSHService) Name() string { return "ssh-enhanced" }

func (s *EnhancedSSHService) HandleConn(ctx *SessionContext) error {
	if ctx.Conn == nil {
		return errors.New("ssh-enhanced service requires a TCP connection")
	}

	_ = ctx.Conn.SetDeadline(time.Now().Add(60 * time.Second))

	// Phase 1: Banner exchange
	if _, err := fmt.Fprintf(ctx.Conn, "%s\r\n", s.Banner); err != nil {
		return fmt.Errorf("send banner: %w", err)
	}

	reader := bufio.NewReader(ctx.Conn)
	clientBanner, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("read client banner: %w", err)
	}
	clientBanner = strings.TrimSpace(clientBanner)

	_ = ctx.Recorder.Event(ctx.Context, ctx.Session, "ssh.client_banner", map[string]any{
		"banner": clientBanner,
	})

	// Phase 2: Key exchange simulation
	// We read raw bytes looking for SSH_MSG_KEXINIT (code 20)
	kexData := make([]byte, 4096)
	n, err := reader.Read(kexData)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("read kex data: %w", err)
	}

	if n > 0 {
		_ = ctx.Recorder.Event(ctx.Context, ctx.Session, "ssh.kex_init", map[string]any{
			"length":        n,
			"first_byte":    fmt.Sprintf("0x%02x", kexData[0]),
			"payload_b64":   base64.StdEncoding.EncodeToString(kexData[:min(n, 256)]),
			"client_banner": clientBanner,
		})
	}

	// Phase 3: Disconnect with "protocol version" error
	// This is realistic — OpenSSH does this when kex fails
	_, _ = fmt.Fprint(ctx.Conn, "Protocol major versions differ.\r\n")
	return nil
}

func (s *EnhancedSSHService) HandlePacket(*PacketContext) error {
	return nil
}
