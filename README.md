# HONEYTRAP

**AI-Powered Deception Framework** — Make Attackers Think They Won

---

## Status

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | Core Engine — TCP/UDP listeners, SSH/HTTP/FTP emulators, CLI, storage | ✅ Complete |
| Phase 2 | AI Emulation + Tokens — Ollama AI, enhanced services, honeytokens, decoy docs | ✅ Complete |
| Phase 3 | Dashboard + Detection — React dashboard, D3 charts, WebSocket, behavioral analysis | ✅ Complete |
| Phase 4 | Hardening + Export — Deploy profiles, STIX export, alert integrations, seccomp, systemd | ✅ Complete |

### Known Limitations

- **No `tokens` or `export` CLI commands** — token generation and STIX export exist as Go packages but are not yet wired to CLI subcommands. Use the API (`POST /tokens`, `GET /analytics`) or the Go packages directly.

---

## Purpose

Deploy intelligent honeypots and deception assets to detect, track, and analyze attackers in real-time. Complements GHOSTWIRE (network forensics), HATCHERY (malware sandbox), and DEADDROP (digital forensics).

---

## Quick Specs

| Attribute | Value |
|-----------|-------|
| **Stack** | Go + Python + TypeScript + Docker + React + PostgreSQL |
| **LOC** | ~6,800 (source), ~10,500 (including dashboard framework) |
| **Source Files** | ~75 |
| **Go Services** | 7 (SSH, SSH+, HTTP, HTTP+, FTP, Redis, UDP) |
| **API Routes** | 9 (sessions, events, tokens CRUD, analytics, health, WebSocket) |
| **Dashboard** | 13 components, 5 pages, 2 hooks |
| **Alert Integrations** | 3 (Slack, Telegram, Email stub) |
| **Deploy Profiles** | 5 (default, minimal, full-spectrum, raspberry-pi, corporate-internal) |

---

## Services

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| SSH | 2222 | TCP | Basic SSH banner capture |
| SSH Enhanced | 2223 | TCP | Full banner exchange, kex capture |
| HTTP | 8080 | TCP | Basic fake admin panel |
| HTTP Enhanced | 8443 | TCP | Full login pages, dashboard, API endpoints |
| FTP | 2121 | TCP | Fake file listings (payroll.csv, backups) |
| Redis | 6379 | TCP | Plausible keys with tempting names |
| UDP Decoy | 9161 | UDP | Generic UDP capture |

> **Port conflict:** The AI emulator defaults to port 8443 (same as HTTP Enhanced). When running both, change the AI emulator port, e.g., `python server.py 8444`, and set `HONEYTRAP_AI_URL=http://localhost:8444`.

---

## AI Emulator (Python)

The Python AI emulator uses Ollama for dynamic response generation:

- **Endpoint:** `POST /ai-response` — Generate dynamic service responses
- **Health:** `GET /ai/health` — Check Ollama connectivity
- **Cache:** `GET /ai/cache` — Response cache statistics
- **Intent Classification:** Automatically classifies attacker intent (recon, exploitation, lateral movement)
- **Fallback:** Static responses when Ollama is unavailable

```bash
cd ai_emulator
pip install -r requirements.txt    # note: also needs pytest, pytest-asyncio for tests
python server.py 8444              # default is 8443, use 8444 to avoid conflict with HTTP+
```

---

## Honeytokens

Generate and track fake credentials to detect unauthorized access:

- **API Keys:** `sk-proj-htk-...` (OpenAI-style)
- **AWS Credentials:** `AKIA...` (AWS-style)
- **Database URLs:** `postgres://admin:password@db.internal...`
- **Document URLs:** `https://internal.honeytrap.local/docs/...`

### Decoy Documents

- `decoys/fake-aws-credentials.json` — Planted AWS keys
- `decoys/fake-database-config.yml` — Fake DB config with passwords
- `decoys/fake-api-key.env` — Planted environment variables

> ⚠️ Decoy files contain realistic-looking fake credentials. Handle with care — do not commit to public repos or expose to systems that might auto-scan them.

---

## Build & Run

```bash
# Build the Go binary
go build ./cmd/honeytrap

# Check status (shows configured services)
./honeytrap status

# Deploy honeypot engine (reads HONEYTRAP_* env vars)
./honeytrap deploy default

# View sessions (JSON)
./honeytrap sessions

# View events (JSON)
./honeytrap events

# List profiles
./honeytrap profiles

# Version
./honeytrap version
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HONEYTRAP_DATA_DIR` | `var` | Directory for JSONL log files |
| `HONEYTRAP_NODE_NAME` | `local-node` | Node identifier |
| `HONEYTRAP_ENV` | `development` | Environment label |
| `HONEYTRAP_PROFILE` | `default` | Deploy profile (informational) |
| `HONEYTRAP_SSH_PORT` | 2222 | SSH listener port |
| `HONEYTRAP_SSH_ENHANCED_PORT` | 2223 | SSH Enhanced port |
| `HONEYTRAP_HTTP_PORT` | 8080 | HTTP listener port |
| `HONEYTRAP_HTTP_ENHANCED_PORT` | 8443 | HTTP Enhanced port |
| `HONEYTRAP_FTP_PORT` | 2121 | FTP listener port |
| `HONEYTRAP_REDIS_PORT` | 6379 | Redis listener port |
| `HONEYTRAP_UDP_PORT` | 9161 | UDP decoy port |
| `HONEYTRAP_AI_URL` | `http://localhost:8443` | AI emulator URL |
| `HONEYTRAP_PROFILES_DIR` | `profiles` | Deploy profile directory |
| `HONEYTRAP_API_URL` | `http://localhost:3000` | API URL for JSONL data bridge |
| `API_TOKEN` | _(empty)_ | Bearer token for API auth (empty = dev mode, no auth) |

### Docker

```bash
docker-compose up -d
```

> **Note:** The `docker-compose.yml` exposes all 7 service ports (2222, 2223, 8080, 8443, 2121, 6379, 9161/udp). The AI emulator runs on port 8444 internally (mapped to avoid conflict with HTTP Enhanced on 8443).

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   HONEYTRAP CLI                     │
│        (deploy, status, sessions, events)           │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│              Honeypot Engine (Go)                    │
│    TCP/UDP listeners, service emulators, sessions   │
│    Writes: var/sessions.jsonl, var/events.jsonl     │
└─────────────────────────────────────────────────────┘
          │                    │
          ▼                    ▼
┌──────────────────┐  ┌──────────────────────────────┐
│  AI Emulator     │  │  Fastify API Server (TS)      │
│  (Ollama+Python) │  │  Reads: PostgreSQL            │
│  Port 8444       │  │  Sessions, Events, Tokens,    │
│                  │  │  Analytics, WebSocket           │
└──────────────────┘  └──────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │  React Dashboard │
                    │  Port 8082       │
                    └──────────────────┘

⚠️ Data flow: The Go engine automatically ingests sessions and events
   into the Fastify API via periodic POST /ingest-record calls.
   Set HONEYTRAP_API_URL to configure the API endpoint.
```

---

## Dashboard (Phase 3)

React + Vite + Tailwind + D3 cyberpunk dashboard:

- **5 Pages:** Dashboard, Sessions, Tokens, Analytics, Settings
- **13 Components:** StatsCards, SessionViewer, SessionDetail, AttackerMap, ServiceChart, TimelineChart, TokenList, TokenAlerts, EventLog, ServiceStatus, CredentialCapture, AIStatus, Sidebar
- **2 Hooks:** useApi, useWebSocket
- **Real-time:** WebSocket hook with auto-reconnect
- **D3 Charts:** Bar chart (service attacks), area chart (24h timeline), world map (attacker geolocation)
- **Dark Theme:** #0a0a1a background, #4ecca3 green accent, #e84545 alerts

```bash
cd dashboard && npm install && npm run dev
# Dashboard runs at http://localhost:5173
# Proxies /api to localhost:3000
```

---

## Behavioral Analysis (Phase 3)

Go module for attacker profiling:

- **IsScripted()** — Detects automated tools (uniform command intervals, CV < 0.3)
- **IsHuman()** — Detects human attackers (variable timing, thinking pauses)
- **ClassifyTool()** — Identifies nmap, hydra, metasploit, nikto, sqlmap, nuclei, curl, wget
- **RiskScore()** — 0-1 risk score (6 factors: events, tool, scripted, duration, dangerous commands, login attempts)

---

## Deploy Profiles (Phase 4)

YAML-based deployment configurations:

| Profile | Services | AI | Use Case |
|---------|----------|----|----------|
| **default** | All 7 | ✅ | Full deployment |
| **minimal** | SSH + HTTP | ❌ | Lightweight |
| **full-spectrum** | All 7 + PCAP | ✅ | Maximum deception |
| **raspberry-pi** | SSH + Redis | ❌ | Low-resource devices |
| **corporate-internal** | SSH + HTTP + FTP | ✅ | Windows/AD environment |

```bash
./honeytrap deploy default    # applies profile ports/settings to engine config
./honeytrap profiles          # lists available profiles
```

---

## STIX Export (Phase 4)

Export honeypot data as STIX 2.1 bundles for threat intel sharing:

- Session data → observed-data + IPv4 address objects
- Token access → indicator objects with confidence scores
- Full STIX bundle with identity and relationship objects

> **Note:** STIX export is available as a Go package (`internal/export`) but not yet wired to a CLI command. Call programmatically or via a future `export` command.

---

## Alert Integrations (Phase 4)

Real-time alerts when attackers interact with honeypots:

- **Slack** — Webhook-based alerts with severity emojis
- **Telegram** — Bot API with Markdown formatting
- **Email** — SMTP structure ready (requires agentmail or net/smtp for full integration)
- Severity levels: low → medium → high → critical

---

## Hardening (Phase 4)

- **Seccomp** — Whitelist profile (default deny, 150+ allowed syscalls)
- **Systemd** — Hardened service files (NoNewPrivileges, ProtectSystem, PrivateTmp)
- **Docker** — Network isolation, read-only FS support, resource limits
- **Install script** — `sudo bash deploy/install.sh`

---

## Version

- Go CLI: `v0.4.0`
- AI Emulator: `v0.2.0`
- API: `v0.2.0`

---

**Created:** April 16, 2026  
**Part of:** Security Portfolio  
**Review:** Bug log available at `BUG_LOG.md`