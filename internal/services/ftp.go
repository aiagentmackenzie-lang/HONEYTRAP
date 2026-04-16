package services

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type FTPService struct{}

func NewFTPService() *FTPService {
	return &FTPService{}
}

func (s *FTPService) Name() string {
	return "ftp"
}

func (s *FTPService) HandleConn(ctx *SessionContext) error {
	if ctx.Conn == nil {
		return errors.New("ftp service requires a TCP connection")
	}

	_ = ctx.Conn.SetDeadline(time.Now().Add(ctx.Deadline))
	reader := bufio.NewReader(ctx.Conn)
	if _, err := fmt.Fprint(ctx.Conn, "220 HONEYTRAP FTP Service ready\r\n"); err != nil {
		return err
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}
		_ = ctx.Recorder.Event(ctx.Context, ctx.Session, "ftp.command", map[string]any{"command": command})

		upper := strings.ToUpper(command)
		switch {
		case strings.HasPrefix(upper, "USER "):
			_, _ = fmt.Fprint(ctx.Conn, "331 Username OK, need password\r\n")
		case strings.HasPrefix(upper, "PASS "):
			_, _ = fmt.Fprint(ctx.Conn, "230 Login successful\r\n")
		case upper == "PWD":
			_, _ = fmt.Fprint(ctx.Conn, `257 "/srv/backups" is the current directory`+"\r\n")
		case upper == "SYST":
			_, _ = fmt.Fprint(ctx.Conn, "215 UNIX Type: L8\r\n")
		case upper == "TYPE I":
			_, _ = fmt.Fprint(ctx.Conn, "200 Type set to I\r\n")
		case upper == "PASV":
			_, _ = fmt.Fprint(ctx.Conn, "502 Passive mode unavailable\r\n")
		case upper == "LIST":
			_, _ = fmt.Fprint(ctx.Conn, "150 Opening ASCII mode data connection for file list\r\n-rw-r--r-- 1 root root 4096 Apr 16 12:00 payroll.csv\r\n226 Transfer complete\r\n")
		case upper == "QUIT":
			_, _ = fmt.Fprint(ctx.Conn, "221 Goodbye\r\n")
			return nil
		default:
			_, _ = fmt.Fprint(ctx.Conn, "502 Command not implemented\r\n")
		}
	}
}

func (s *FTPService) HandlePacket(*PacketContext) error {
	return nil
}
