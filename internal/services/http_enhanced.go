package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// EnhancedHTTPService serves realistic-looking fake web applications
// that capture attacker credentials and track their behavior.
type EnhancedHTTPService struct {
	BaseService
}

func NewEnhancedHTTPService() *EnhancedHTTPService {
	return &EnhancedHTTPService{}
}

func (s *EnhancedHTTPService) Name() string { return "http-enhanced" }

func (s *EnhancedHTTPService) HandleConn(ctx *SessionContext) error {
	if ctx.Conn == nil {
		return fmt.Errorf("http-enhanced service requires a TCP connection")
	}

	_ = ctx.Conn.SetDeadline(time.Now().Add(ctx.Deadline))

	for {
		req, err := http.ReadRequest(bufio.NewReader(ctx.Conn))
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		defer req.Body.Close()

		body, _ := io.ReadAll(io.LimitReader(req.Body, 16384))

		// Capture POST data (credentials)
		event := map[string]any{
			"method":     req.Method,
			"path":       req.URL.Path,
			"user_agent": req.UserAgent(),
			"host":       req.Host,
			"referer":    req.Referer(),
		}
		if len(body) > 0 {
			event["body"] = string(body)
			event["content_type"] = req.Header.Get("Content-Type")

			// Extract form credentials
			if strings.Contains(req.Header.Get("Content-Type"), "form") {
				if err := req.ParseForm(); err == nil {
					if username := req.FormValue("username"); username != "" {
						event["captured_username"] = username
						event["captured_password"] = req.FormValue("password")
					}
				}
			}
		}

		_ = ctx.Recorder.Event(ctx.Context, ctx.Session, "http.request", event)

		// Route to appropriate fake page
		html := routeFakePage(req.URL.Path, req.Method, body)

		response := fmt.Sprintf(
			"HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=utf-8\r\nContent-Length: %d\r\nConnection: keep-alive\r\nServer: nginx/1.24.0\r\nX-Request-ID: %s\r\n\r\n%s",
			len(html),
			ctx.Session.ID,
			html,
		)

		_, err = io.Copy(ctx.Conn, bytes.NewBufferString(response))
		if err != nil {
			return err
		}
	}
}

func (s *EnhancedHTTPService) HandlePacket(*PacketContext) error {
	return nil
}

// routeFakePage returns realistic HTML based on the URL path
func routeFakePage(path, method string, body []byte) string {
	switch {
	case strings.Contains(path, "login") || path == "/":
		return fakeLoginPage()
	case strings.Contains(path, "dashboard") || strings.Contains(path, "admin"):
		return fakeDashboard()
	case strings.Contains(path, "api"):
		return fakeAPIResponse(path)
	case strings.Contains(path, "config") || strings.Contains(path, "settings"):
		return fakeConfigPage()
	case strings.Contains(path, "health"):
		return fakeHealthCheck()
	default:
		return fake404()
	}
}

func fakeLoginPage() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Secure Login Gateway</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f0f2f5; display: flex; align-items: center; justify-content: center; min-height: 100vh; }
.login-container { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); width: 400px; }
.logo { text-align: center; margin-bottom: 30px; }
.logo h1 { color: #1a73e8; font-size: 24px; }
.logo p { color: #5f6368; font-size: 14px; margin-top: 4px; }
.form-group { margin-bottom: 20px; }
.form-group label { display: block; font-size: 14px; color: #333; margin-bottom: 6px; font-weight: 500; }
.form-group input { width: 100%; padding: 12px; border: 1px solid #dadce0; border-radius: 4px; font-size: 14px; }
.form-group input:focus { outline: none; border-color: #1a73e8; }
.btn { width: 100%; padding: 12px; background: #1a73e8; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; }
.btn:hover { background: #1557b0; }
.footer { text-align: center; margin-top: 20px; font-size: 12px; color: #5f6368; }
.security-badge { text-align: center; margin-top: 15px; font-size: 11px; color: #1a73e8; }
</style>
</head>
<body>
<div class="login-container">
<div class="logo">
<h1>🔒 Secure Login</h1>
<p>Operations Control Panel</p>
</div>
<form method="POST" action="/login" id="loginForm">
<div class="form-group">
<label for="username">Username</label>
<input type="text" id="username" name="username" placeholder="Enter your username" autocomplete="username">
</div>
<div class="form-group">
<label for="password">Password</label>
<input type="password" id="password" name="password" placeholder="Enter your password" autocomplete="current-password">
</div>
<button type="submit" class="btn">Sign In</button>
</form>
<div class="security-badge">🛡️ Protected by Enterprise Security</div>
<div class="footer">© 2026 Operations Team · <a href="/forgot-password">Forgot password?</a></div>
</div>
</body>
</html>`
}

func fakeDashboard() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Operations Dashboard</title>
<style>
body { font-family: -apple-system, sans-serif; background: #1a1a2e; color: #eee; margin: 0; }
.sidebar { width: 220px; background: #16213e; height: 100vh; position: fixed; padding: 20px; }
.sidebar h2 { color: #0f3460; margin-bottom: 30px; }
.sidebar a { display: block; color: #a8a8b3; text-decoration: none; padding: 10px; margin: 2px 0; border-radius: 4px; }
.sidebar a:hover { background: #0f3460; color: white; }
.main { margin-left: 260px; padding: 20px; }
.card { background: #16213e; border-radius: 8px; padding: 20px; margin: 10px 0; }
.stat { display: inline-block; width: 23%; margin: 1%; padding: 20px; background: #16213e; border-radius: 8px; text-align: center; }
.stat h3 { font-size: 32px; color: #4ecca3; }
.stat p { color: #a8a8b3; font-size: 12px; }
table { width: 100%; border-collapse: collapse; }
th, td { text-align: left; padding: 12px; border-bottom: 1px solid #0f3460; }
.badge { padding: 4px 8px; border-radius: 12px; font-size: 11px; }
.badge-green { background: #4ecca3; color: #1a1a2e; }
.badge-yellow { background: #e8b84a; color: #1a1a2e; }
</style>
</head>
<body>
<div class="sidebar">
<h2>⬡ Operations</h2>
<a href="/dashboard">📊 Dashboard</a>
<a href="/admin">👥 Admin Panel</a>
<a href="/config">⚙️ Configuration</a>
<a href="/api/health">🩺 System Health</a>
<a href="/settings">🔒 Security</a>
</div>
<div class="main">
<h1>System Overview</h1>
<div class="stat"><h3>98.7%</h3><p>Uptime</p></div>
<div class="stat"><h3>1,247</h3><p>Active Sessions</p></div>
<div class="stat"><h3>42</h3><p>Alerts Today</p></div>
<div class="stat"><h3>3</h3><p>Pending Reviews</p></div>
<div class="card">
<h3>Recent Activity</h3>
<table>
<tr><th>Time</th><th>User</th><th>Action</th><th>Status</th></tr>
<tr><td>16:42</td><td>admin</td><td>Config update</td><td><span class="badge badge-green">Success</span></td></tr>
<tr><td>16:38</td><td>svc-account</td><td>Deployment</td><td><span class="badge badge-green">Success</span></td></tr>
<tr><td>16:15</td><td>db-admin</td><td>Schema migration</td><td><span class="badge badge-yellow">Pending</span></td></tr>
</table>
</div>
</div>
</body>
</html>`
}

func fakeAPIResponse(path string) string {
	data := map[string]any{
		"status":  "healthy",
		"version": "2.4.1",
		"uptime":  86400,
		"services": map[string]string{
			"auth":       "running",
			"database":   "running",
			"cache":      "running",
			"queue":      "running",
			"monitoring": "running",
		},
	}
	if strings.Contains(path, "keys") {
		data = map[string]any{
			"keys": []map[string]string{
				{"id": "key-001", "name": "production-api-key", "status": "active"},
				{"id": "key-002", "name": "staging-api-key", "status": "active"},
			},
		}
	}
	b, _ := json.Marshal(data)
	return "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nConnection: close\r\nServer: nginx/1.24.0\r\n\r\n" + string(b)
}

func fakeConfigPage() string {
	return `<!DOCTYPE html>
<html><head><title>Configuration</title></head>
<body style="font-family:monospace;background:#1e1e1e;color:#d4d4d4;padding:20px;">
<h1>⚙️ System Configuration</h1>
<pre>
# /etc/honeytrap/config.yml
server:
  host: 0.0.0.0
  port: 443
  ssl: true
  certificate: /etc/ssl/honeytrap.crt

database:
  host: db.internal.honeytrap.local
  port: 5432
  name: operations
  user: admin
  # WARNING: password stored in vault

auth:
  method: ldap
  server: ldap.internal.honeytrap.local
  base_dn: dc=honeytrap,dc=local

monitoring:
  enabled: true
  endpoint: https://metrics.internal.honeytrap.local
  api_key: sk-proj-htk-REDACTED
</pre>
</body></html>`
}

func fakeHealthCheck() string {
	data := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"services": map[string]string{
			"api":      "healthy",
			"database": "healthy",
			"cache":    "healthy",
		},
	}
	b, _ := json.Marshal(data)
	return "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nConnection: close\r\n\r\n" + string(b)
}

func fake404() string {
	return `<!DOCTYPE html>
<html><head><title>404 - Not Found</title></head>
<body style="font-family:sans-serif;text-align:center;padding-top:100px;">
<h1>404</h1>
<p>The page you requested was not found.</p>
<p><a href="/">Return to Dashboard</a></p>
</body></html>`
}
