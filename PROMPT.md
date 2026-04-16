# HONEYTRAP Phase 3 — Build Prompt for Next Agent

## Context

You are continuing the HONEYTRAP build — an AI-Powered Deception Framework honeypot system. Phases 1 and 2 are COMPLETE and pushed to GitHub. You are building Phase 3: Dashboard + Advanced Detection.

**Working Directory:** `/Users/main/Security Apps/HONEYTRAP`
**GitHub:** https://github.com/aiagentmackenzie-lang/HONEYTRAP
**Read first:** `PROGRESS.md` and `SPEC.md` in the project root

---

## What Already Exists (DO NOT recreate these)

### Go Engine (Phase 1+2)
- `cmd/honeytrap/main.go` — CLI entry point (deploy, status, sessions, events, version)
- `internal/engine/engine.go` — Core engine with 7 registered services
- `internal/engine/listener.go` — TCP/UDP listener
- `internal/engine/session_manager.go` — Session lifecycle + event recording
- `internal/config/config.go` — Env-based config for 7 services
- `internal/services/` — ssh.go, ssh_enhanced.go, http.go, http_enhanced.go, ftp.go, redis.go, udp.go
- `internal/ai/client.go` — Go client for AI emulator
- `internal/tokens/tokens.go` — Honeytoken generator + store
- `internal/models/models.go` — Session, Event, Token, ServiceStatus structs
- `internal/storage/` — repository.go, memory.go, postgres.go

### Python AI Emulator (Phase 2)
- `ai_emulator/emulator.py` — Ollama-powered dynamic response generator
- `ai_emulator/server.py` — FastAPI server (POST /ai-response, GET /ai/health, GET /ai/cache)
- `ai_emulator/test_emulator.py` — Tests
- `ai_emulator/requirements.txt` + `Dockerfile`

### Fastify API (Phase 1+2)
- `api/src/index.ts` — Server entry (registers sessions, events, tokens routes)
- `api/src/routes/sessions.ts` — GET /sessions
- `api/src/routes/events.ts` — GET /events
- `api/src/routes/tokens.ts` — CRUD + access detection alerts
- `api/src/plugins/db.ts` — PostgreSQL plugin

### Other
- `db/schema.sql` — sessions, events, tokens, token_access_log tables with indexes
- `docker-compose.yml` — postgres, api, ai-emulator, honeytrap services
- `decoys/` — fake-aws-credentials.json, fake-database-config.yml, fake-api-key.env
- `go.mod` — Go 1.26, module `github.com/aiagentmackenzie-lang/HONEYTRAP`

### Dashboard (barely started)
- `dashboard/package.json` — React 18, D3, Tailwind, Vite, react-router-dom, lucide-react
- `dashboard/src/` — empty component/hooks/pages directories
- `dashboard/public/` — empty

---

## Phase 3: Build This Now

### Step 1: Dashboard Scaffolding
Create the React + Vite + Tailwind project structure:
- `dashboard/vite.config.ts` — Vite config with React plugin, proxy /api to localhost:3000
- `dashboard/tailwind.config.js` — Dark theme, custom colors (honeytrap red, cyber green)
- `dashboard/postcss.config.js` — Tailwind + autoprefixer
- `dashboard/tsconfig.json` — TypeScript strict mode
- `dashboard/index.html` — Root HTML with dark background
- `dashboard/src/main.tsx` — React entry, Router, Tailwind import
- `dashboard/src/index.css` — Tailwind directives + custom dark theme

### Step 2: Layout & Navigation
- `dashboard/src/App.tsx` — Main layout with sidebar + router
- `dashboard/src/components/Sidebar.tsx` — Nav links: Dashboard, Sessions, Tokens, Analytics, Settings (use lucide-react icons, dark theme, active state highlighting)

### Step 3: Core Pages (5 pages)
1. `dashboard/src/pages/DashboardPage.tsx` — Overview page with stats cards + charts
2. `dashboard/src/pages/SessionsPage.tsx` — Session list + detail modal
3. `dashboard/src/pages/TokensPage.tsx` — Token management (list, create, deactivate)
4. `dashboard/src/pages/AnalyticsPage.tsx` — Charts + attacker map
5. `dashboard/src/pages/SettingsPage.tsx` — Service configuration display

### Step 4: Components (~12)
- `dashboard/src/components/StatsCards.tsx` — 4 cards: Total Sessions, Active Now, Alerts Today, Tokens Triggered
- `dashboard/src/components/SessionViewer.tsx` — Live session table (IP, service, started, status)
- `dashboard/src/components/SessionDetail.tsx` — Expand/collapse session with event timeline
- `dashboard/src/components/AttackerMap.tsx` — D3 world map with attack points (use mock geo data for now)
- `dashboard/src/components/ServiceChart.tsx` — D3 bar chart: attacks per service (SSH, HTTP, FTP, Redis)
- `dashboard/src/components/TimelineChart.tsx` — D3 area chart: attacks over last 24h
- `dashboard/src/components/TokenList.tsx` — Token table with kind badges, status, access count
- `dashboard/src/components/TokenAlerts.tsx` — Alert feed when tokens are accessed
- `dashboard/src/components/EventLog.tsx` — Real-time scrolling event stream
- `dashboard/src/components/ServiceStatus.tsx` — Grid showing all 7 services with status indicators
- `dashboard/src/components/CredentialCapture.tsx` — Show captured credentials from HTTP login
- `dashboard/src/components/AIStatus.tsx` — Ollama health, model, cache stats

### Step 5: Hooks
- `dashboard/src/hooks/useApi.ts` — Fetch wrapper for REST API (baseURL from env, error handling, typed responses)
- `dashboard/src/hooks/useWebSocket.ts` — WebSocket hook connecting to Fastify WS endpoint, auto-reconnect, message parsing

### Step 6: Backend Additions
- `api/src/routes/ws.ts` — WebSocket endpoint for live session/event push
- `api/src/routes/analytics.ts` — GET /analytics (top IPs, service breakdown, attack timeline, token stats)
- Update `api/src/index.ts` to register ws and analytics routes

### Step 7: Behavioral Analysis (Go)
- `internal/analysis/behavioral.go` — Analyze session patterns:
  - `IsScripted(session []Event) bool` — Detect automated tools (rapid commands, identical intervals)
  - `IsHuman(session []Event) bool` — Detect human behavior (variable timing, typos, exploration)
  - `ClassifyTool(events []Event) string` — Identify tool (nmap, hydra, metasploit, custom)
  - `RiskScore(session Session, events []Event) float64` — 0-1 risk score

### Step 8: Dashboard Docker
- `dashboard/Dockerfile` — Multi-stage build (npm install → npm run build → nginx serve)
- Update `docker-compose.yml` to include dashboard service

---

## Design Guidelines

- **Dark theme throughout** — #0a0a1a background, #1a1a2e cards, #4ecca3 accent green, #e84545 alert red
- **Cyberpunk aesthetic** — Monospace fonts for data, subtle glow effects, grid lines
- **Real-time feel** — WebSocket updates, animated counters, pulse effects on new events
- **D3 for all charts** — No chart libraries, raw D3 for full control
- **Responsive** — Works on desktop, sidebar collapses on mobile
- **Professional** — This is a portfolio piece, make it look premium

---

## Build Verification

After completing Phase 3:
1. Run `cd /Users/main/Security Apps/HONEYTRAP && go build ./cmd/honeytrap` — must pass
2. Run `cd dashboard && npm install && npm run build` — must pass (if npm available)
3. Run `go test ./...` — must pass
4. Update PROGRESS.md to reflect Phase 3 completion
5. Update README.md with dashboard section
6. Git commit and push: `git add -A && git commit -m "feat: HONEYTRAP Phase 3 — Dashboard and advanced detection"`
7. Watch out for GitHub push protection on decoy files — use DECOY/NOT_REAL markers

---

## Important Gotchas

1. **PacketContext has no Session field** — Removed in Phase 2. UDP services reference `models.Session` directly.
2. **BaseService is an empty struct** — Used as embedded field in enhanced services.
3. **GitHub push protection** — Fake credentials in decoys must have DECOY/NOT_REAL markers.
4. **Dashboard package.json already exists** — Don't recreate it, just add the missing config files.
5. **Working directory has spaces** — Path is `/Users/main/Security Apps/HONEYTRAP` (quote it in shell commands).
