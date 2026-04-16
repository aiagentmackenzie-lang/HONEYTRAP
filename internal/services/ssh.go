package services

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type SSHService struct{}

func NewSSHService() *SSHService {
	return &SSHService{}
}

func (s *SSHService) Name() string {
	return "ssh"
}

func (s *SSHService) HandleConn(ctx *SessionContext) error {
	if ctx.Conn == nil {
		return errors.New("ssh service requires a TCP connection")
	}

	_ = ctx.Conn.SetDeadline(time.Now().Add(ctx.Deadline))
	reader := bufio.NewReader(ctx.Conn)
	if _, err := fmt.Fprint(ctx.Conn, "SSH-2.0-OpenSSH_9.3p1 Debian-1\r\n"); err != nil {
		return err
	}

	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	clientBanner := strings.TrimSpace(line)
	if clientBanner != "" {
		_ = ctx.Recorder.Event(ctx.Context, ctx.Session, "ssh.client_banner", map[string]any{"banner": clientBanner})
	}

	_, _ = fmt.Fprint(ctx.Conn, "Protocol mismatch.\r\n")
	return nil
}

func (s *SSHService) HandlePacket(*PacketContext) error {
	return nil
}
