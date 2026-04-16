# HONEYTRAP — Active Deception & Honeypot Framework

**Purpose:** Deploy intelligent honeypots and deception assets to detect, track, and analyze attackers in real-time  
**Stack:** Go (core) + Python (AI emulation) + Docker (sandbox) + React (dashboard)  
**Portfolio Gap:** Deception tech (complements GHOSTWIRE/HATCHERY/DEADDROP)

---

## Competitive Analysis (April 2026)

| Project | Stars | Language | Key Features | Gap |
|---------|-------|----------|--------------|-----|
| **honeytrap/honeytrap** | 1,298 | Go | Advanced framework, plugins | No AI emulation, dated UI |
| **T-Pot (telekom-security)** | 9k | Multi | Multi-honeypot platform | Heavy, VM-based, not container-native |
| **Beelzebub** | 1.9k | Go | Low-code, AI system virtualization | Closest to our vision |
| **Splunk DECEIVE** | 281 | Python | LLM-powered honeypot | Splunk-locked, enterprise-focused |
| **VelLMes** | 75 | Python | Interactive LLM honeypots | Single-purpose, no framework |

**Our Edge:**
- AI-powered dynamic emulation (LLM-driven responses, not static)
- Container-native (Docker-first, not VM-heavy)
- Modern React dashboard with real-time attacker tracking
- Tight integration with existing portfolio (DEADDROP forensics, GHOSTWIRE network analysis, HATCHERY malware sandbox)
- Lightweight, deployable on Raspberry Pi or cloud

---

## Core Features

### 1. Honeypot Engine
- **Low-interaction:** Emulate services (SSH, HTTP, FTP, SMB, Redis, PostgreSQL)
- **High-interaction:** Containerized real services with monitoring
- **AI-emulated:** LLM-driven responses that adapt to attacker behavior
- **Custom protocols:** YAML-based service definitions

### 2. Deception Assets
- **Honeytokens:** Fake API keys, credentials, database entries
- **Decoy documents:** Plausible-looking files with tracking beacons
- **Fake services:** Internal "admin panels", "databases", "APIs"
- **Network deception:** Fake open ports, banner spoofing

### 3. Detection & Tracking
- **Real-time alerts:** WebSocket push to dashboard
- **Attacker profiling:** GeoIP, ASN, behavioral patterns
- **Session recording:** Full command logging (SSH, shell)
- **Screenshot capture:** Web interaction snapshots
- **Network forensics:** PCAP capture, JA4+ fingerprinting (reuse GHOSTWIRE)

### 4. AI Integration
- **Dynamic responses:** LLM generates plausible service responses
- **Behavioral analysis:** Detect scripted vs. human attackers
- **Intent classification:** Reconnaissance, exploitation, lateral movement
- **Auto-reporting:** Generate incident summaries

### 5. Dashboard
- **Live map:** Attacker geolocation
- **Session viewer:** Real-time command monitoring
- **Analytics:** Top IPs, services targeted, attack patterns
- **Export:** STIX/TAXII, JSON, CSV (reuse DEADDROP export)

### 6. CLI
- `honeytrap deploy <profile>` — Deploy honeypot profile
- `honeytrap status` — Show active honeypots
- `honeytrap sessions` — List captured sessions
- `honeytrap replay <id>` — Replay session recording
- `honeytrap tokens` — Manage honeytokens
- `honeytrap export` — Export data

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   HONEYTRAP CLI                     │
│              (deploy, status, tokens)               │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│                 Fastify API Server                  │
│          (10 endpoints + WebSocket for live)        │
└─────────────────────────────────────────────────────┘
           │                    │                    │
           ▼                    ▼                    ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  Honeypot Engine │  │  Token Manager   │  │  AI Emulator     │
│  (Go + Docker)   │  │  (PostgreSQL)    │  │  (Ollama + Python)│
└──────────────────┘  └──────────────────┘  └──────────────────┘
           │
           ▼
┌──────────────────┐
│  Docker Sandbox  │
│  (seccomp,       │
│   network ns)    │
└──────────────────┘
           │
           ▼
┌──────────────────┐
│   React + D3     │
│   Dashboard      │
└──────────────────┘
```

---

## Build Plan (4 Phases)

### Phase 1: Core Engine (2-3 days)
**Goal:** Basic honeypot framework with low-interaction services

- [ ] Project scaffolding (Go module, directory structure)
- [ ] TCP/UDP listener engine
- [ ] Service emulators: SSH, HTTP, FTP
- [ ] Session logging (commands, timestamps, IPs)
- [ ] PostgreSQL schema (sessions, events, tokens)
- [ ] Fastify API: GET /sessions, GET /events
- [ ] Basic CLI: `deploy`, `status`, `sessions`

**Files:** ~15 Go files, 1 schema.sql, 5 CLI commands  
**LOC Estimate:** ~2,500

---

### Phase 2: AI Emulation + Deception Assets (2-3 days)
**Goal:** AI-powered responses and honeytoken tracking

- [ ] Python AI emulator service (Ollama integration)
- [ ] Dynamic SSH response generation (LLM-driven)
- [ ] HTTP honeypot with fake login pages
- [ ] Honeytoken generator (API keys, credentials)
- [ ] Token usage tracking (alert when accessed)
- [ ] Decoy document templates (with tracking pixels)
- [ ] Fastify API: POST /ai-response, GET /tokens, POST /tokens
- [ ] CLI: `tokens` command

**Files:** ~10 Python files, ~8 Go files, token templates  
**LOC Estimate:** ~3,000

---

### Phase 3: Dashboard + Advanced Detection (2-3 days)
**Goal:** Real-time monitoring and attacker tracking

- [ ] React + Vite + Tailwind setup
- [ ] Live session viewer (WebSocket)
- [ ] Attacker map (geoIP + D3)
- [ ] Analytics dashboard (charts: top IPs, services, timelines)
- [ ] Session replay (command-by-command playback)
- [ ] Behavioral analysis module (scripted vs. human)
- [ ] PCAP capture integration (reuse GHOSTWIRE lib)
- [ ] Fastify API: WebSocket endpoint, GET /analytics

**Files:** ~12 React components, ~5 Go handlers  
**LOC Estimate:** ~3,500

---

### Phase 4: Hardening + Export + Docker (2 days)
**Goal:** Production-ready, deployable, integrated

- [ ] Docker sandbox image (seccomp profiles, network namespaces)
- [ ] Deploy profiles (YAML: which services, ports, AI settings)
- [ ] STIX/TAXII export (reuse DEADDROP)
- [ ] Alert integrations (Slack, Telegram, Email via agentmail)
- [ ] Systemd service files
- [ ] Documentation (README, deployment guide, API docs)
- [ ] End-to-end testing (deploy, attack, capture, alert)
- [ ] GitHub push

**Files:** Dockerfile, systemd units, deploy profiles, tests  
**LOC Estimate:** ~1,500

---

## Total Estimates

| Metric | Value |
|--------|-------|
| **Total LOC** | ~10,500 |
| **Total Files** | ~80 |
| **Build Time** | 8-11 days (1-2 phases per session) |
| **Complexity** | High (AI + real-time + Docker isolation) |

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| AI responses too slow | Cache common responses, async generation |
| Container escape | Seccomp profiles, read-only FS, network namespaces |
| False positives | Behavioral analysis, confidence scoring |
| Resource exhaustion | Rate limiting per IP, session timeouts |
| Legal concerns | Clear documentation: "research/defense only" |

---

## Integration Points

| Portfolio Project | Integration |
|-------------------|-------------|
| **GHOSTWIRE** | Reuse JA4+ fingerprinting, network forensics |
| **DEADDROP** | Reuse STIX export, YARA scanning for captured malware |
| **HATCHERY** | Send captured malware samples for sandbox analysis |
| **AI Agent Security Monitor** | Share PostgreSQL schema, alert system |

---

## Success Metrics

- ✅ Deploy 5+ honeypot profiles (SSH, HTTP, Redis, SMB, custom AI)
- ✅ Capture 100+ real attacks in first 30 days
- ✅ AI emulation undetectable from real services (attacker stays engaged)
- ✅ <100ms latency on AI responses (cached + async)
- ✅ Zero container escapes in security testing
- ✅ Dashboard shows real-time sessions with <2s delay

---

## Next Steps

1. **Approve this spec** (or request changes)
2. **Phase 1 kickoff** — Core engine scaffolding
3. **Session cadence:** 1-2 phases per deep-work session

---

**GitHub Target:** `github.com/aiagentmackenzie-lang/HONEYTRAP`  
**Tagline:** "AI-Powered Deception Framework — Make Attackers Think They Won"