# HONEYTRAP

**AI-Powered Deception Framework** — Make Attackers Think They Won

---

## Status: Phase 3 COMPLETE 🕸️

- **Spec:** ✅ Complete (SPEC.md)
- **Phase 1:** ✅ Core Engine — TCP/UDP listeners, SSH/HTTP/FTP emulators, CLI, PostgreSQL schema
- **Phase 2:** ✅ AI Emulation + Tokens — Ollama AI, enhanced services, honeytokens, decoy docs
- **Phase 3:** ✅ Dashboard + Advanced Detection — React dashboard, D3 charts, WebSocket, behavioral analysis
- **Phase 4:** 🔲 Hardening + Export + Docker

---

## Purpose

Deploy intelligent honeypots and deception assets to detect, track, and analyze attackers in real-time. Complements GHOSTWIRE (network forensics), HATCHERY (malware sandbox), and DEADDROP (digital forensics).

---

## Quick Specs

| Attribute | Value |
|-----------|-------|
| **Stack** | Go + Python + Docker + React + PostgreSQL |
| **LOC** | ~8,500+ |
| **Phases** | 4 (Core → AI → Dashboard → Hardening) |
| **GitHub** | `github.com/aiagentmackenzie-lang/HONEYTRAP` |
| **Portfolio Gap** | Deception tech |

---

## Services (Phase 2)

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| SSH | 2222 | TCP | Basic SSH banner capture |
| SSH Enhanced | 2223 | TCP | Full banner exchange, kex capture |
| HTTP | 8080 | TCP | Basic fake admin panel |
| HTTP Enhanced | 8443 | TCP | Full login pages, dashboard, API endpoints |
| FTP | 2121 | TCP | Fake file listings (payroll.csv, backups) |
| Redis | 6379 | TCP | Plausible keys with tempting names |
| UDP Decoy | 9161 | UDP | Generic UDP capture |

---

## AI Emulator (Phase 2)

The Python AI emulator uses Ollama for dynamic response generation:

- **Endpoint:** `POST /ai-response` — Generate dynamic service responses
- **Health:** `GET /ai/health` — Check Ollama connectivity
- **Cache:** `GET /ai/cache` — Response cache statistics
- **Intent Classification:** Automatically classifies attacker intent (recon, exploitation, lateral movement)
- **Fallback:** Static responses when Ollama is unavailable

---

## Honeytokens (Phase 2)

Generate and track fake credentials to detect unauthorized access:

- **API Keys:** `sk-proj-htk-...` (OpenAI-style)
- **AWS Credentials:** `AKIA...` (AWS-style)
- **Database URLs:** `postgres://admin:password@db.internal...`
- **Document URLs:** `https://internal.honeytrap.local/docs/...`

### Decoy Documents

- `decoys/fake-aws-credentials.json` — Planted AWS keys
- `decoys/fake-database-config.yml` — Fake DB config with passwords
- `decoys/fake-api-key.env` — Planted environment variables

---

## Build & Run

```bash
# Build the Go binary
go build ./cmd/honeytrap

# Check status
./honeytrap status

# Deploy honeypot
./honeytrap deploy default

# View sessions
./honeytrap sessions

# View events
./honeytrap events

# Manage tokens
./honeytrap tokens
```

### Docker

```bash
docker-compose up -d
```

### AI Emulator (Python)

```bash
cd ai_emulator
pip install -r requirements.txt
python server.py 8443
```

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   HONEYTRAP CLI                     │
│        (deploy, status, sessions, tokens)           │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│                 Fastify API Server                  │
│    (sessions, events, tokens, AI-response)          │
└─────────────────────────────────────────────────────┘
           │                    │                    │
           ▼                    ▼                    ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  Honeypot Engine │  │  Token Manager   │  │  AI Emulator     │
│  (Go + Docker)   │  │  (PostgreSQL)    │  │  (Ollama + Python)│
└──────────────────┘  └──────────────────┘  └──────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────┐
│  7 Services: SSH, SSH+, HTTP, HTTP+, FTP, Redis, UDP │
└──────────────────────────────────────────────────────┘
```

## Dashboard (Phase 3)

React + Vite + Tailwind + D3 cyberpunk dashboard:

- **5 Pages:** Dashboard, Sessions, Tokens, Analytics, Settings
- **12 Components:** StatsCards, SessionViewer, SessionDetail, AttackerMap, ServiceChart, TimelineChart, TokenList, TokenAlerts, EventLog, ServiceStatus, CredentialCapture, AIStatus
- **Real-time:** WebSocket hook with auto-reconnect
- **D3 Charts:** Bar chart (service attacks), area chart (24h timeline), world map (attacker geolocation)
- **Dark Theme:** #0a0a1a background, #4ecca3 green accent, #e84545 alerts

```bash
cd dashboard && npm install && npm run dev
# Dashboard runs at http://localhost:5173
# Proxies /api to localhost:3000
```

## Behavioral Analysis (Phase 3)

Go module for attacker profiling:

- **IsScripted()** — Detects automated tools (uniform command intervals, CV < 0.3)
- **IsHuman()** — Detects human attackers (variable timing, thinking pauses)
- **ClassifyTool()** — Identifies nmap, hydra, metasploit, nikto, sqlmap, nuclei
- **RiskScore()** — 0-1 risk score (6 factors: events, tool, scripted, duration, dangerous commands, login attempts)

---

**Created:** April 16, 2026  
**Part of:** Raphael's Security Portfolio (21+ projects)