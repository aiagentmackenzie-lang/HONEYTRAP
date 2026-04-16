package services

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPService struct{}

func NewHTTPService() *HTTPService {
	return &HTTPService{}
}

func (s *HTTPService) Name() string {
	return "http"
}

func (s *HTTPService) HandleConn(ctx *SessionContext) error {
	if ctx.Conn == nil {
		return errors.New("http service requires a TCP connection")
	}

	_ = ctx.Conn.SetDeadline(time.Now().Add(ctx.Deadline))
	req, err := http.ReadRequest(bufio.NewReader(ctx.Conn))
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	defer req.Body.Close()

	body, err := io.ReadAll(io.LimitReader(req.Body, 8192))
	if err != nil {
		return err
	}

	event := map[string]any{
		"method":     req.Method,
		"path":       req.URL.Path,
		"user_agent": req.UserAgent(),
	}
	if len(body) > 0 {
		event["body"] = string(body)
	}
	_ = ctx.Recorder.Event(ctx.Context, ctx.Session, "http.request", event)

	html := fakeAdminPanel(req.URL.Path)
	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=utf-8\r\nContent-Length: %d\r\nConnection: close\r\nServer: nginx/1.24.0\r\n\r\n%s",
		len(html),
		html,
	)
	_, err = io.Copy(ctx.Conn, bytes.NewBufferString(response))
	return err
}

func (s *HTTPService) HandlePacket(*PacketContext) error {
	return nil
}

func fakeAdminPanel(path string) string {
	title := "Operations Control Panel"
	if strings.Contains(path, "login") {
		title = "Secure Login Gateway"
	}
	return "<!doctype html><html><head><title>" + title + "</title></head><body><h1>" + title + "</h1><p>Node health nominal.</p><form method='post' action='/login'><input name='username' placeholder='Username'/><input name='password' type='password' placeholder='Password'/><button type='submit'>Sign in</button></form></body></html>"
}
