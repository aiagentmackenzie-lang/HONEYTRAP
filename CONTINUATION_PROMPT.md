# HONEYTRAP — Continuation Prompt

You are picking up a partially-complete code quality and security review on the HONEYTRAP project at `/Users/main/Security Apps/HONEYTRAP`. The previous session completed an initial review, fixed 12 issues, and pushed commit `c4df80d` to `origin/main`. 

**Your role:** Lead Code Quality Reviewer & Lead Security Developer. Continue fixing the remaining bugs documented in `BUG_LOG.md`. The remaining issues are architectural and require deeper changes.

---

## What Was Already Fixed (commit c4df80d)

These are DONE. Do NOT redo them:

1. ✅ **Race condition on `engine.listeners`/`engine.packetConns`** — Added `sync.Mutex` to protect concurrent append/read in `Shutdown()`
2. ✅ **JSONL file permissions** — Changed from `0o644` to `0o600` in `memory.go`
3. ✅ **Profile key name normalization** — Added `normalizeServiceName()` in `profile.go` mapping `ssh_enhanced` → `ssh-enhanced`, `http_enhanced` → `http-enhanced`, `udp` → `udp-decoy`
4. ✅ **FTP credential event logging** — Added `ftp.login` and `ftp.password` events
5. ✅ **Redis AUTH event logging** — Added `redis.auth` events in HandleConn for both RESP and inline paths
6. ✅ **Dockerfile missing ports** — Added `EXPOSE 2223 8443 6379`
7. ✅ **docker-compose missing ports** — Added all 7 service port mappings; moved AI emulator to 8444
8. ✅ **E2E test quality** — Replaced no-op seccomp/systemd tests with real file validation
9. ✅ **Removed spurious DatabaseURL warning** — Cleaned up `app.go`
10. ✅ **Removed unused DatabaseURL field** — Removed from `Config` struct
11. ✅ **CLI help text** — Documented that deploy profiles are informational only
12. ✅ **README rewrite** — Corrected all inaccuracies, added Known Limitations section

---

## Remaining Issues — Fix These Now

Work through these in priority order. Each fix should be committed separately with a clear message.

### 🔴 CRITICAL-01: Wire deploy profiles to engine configuration

**Files:** `internal/app/app.go`, `internal/cli/root.go`, `internal/config/config.go`, `internal/engine/engine.go`

**Problem:** The `deploy` CLI command loads a YAML profile, prints it, then runs the engine with `config.Load()` defaults. The profile's port overrides, service toggles, max_sessions, AI settings, etc. are completely ignored.

**How it works now:**
```
main.go → app.New() → config.Load() [env vars only] → engine.New(cfg, repo) → cli.Runner
                                                                          ↓
cli.Runner.Run("deploy", "minimal") → LoadProfile("minimal") → just prints → engine.Run()
```

**How it should work:**
```
main.go → cli.Run("deploy", "minimal") → config.Load() → applyProfile(config, profile) → engine.New(cfg, repo) → engine.Run()
```

**Required changes:**

1. **`internal/config/config.go`** — Add a function:
```go
func ApplyProfile(cfg *Config, profile *DeployProfile) *Config {
    // For each service in profile.Services:
    //   - Map the profile key to engine service name (use normalizeServiceName)
    //   - Override port from profile if set
    //   - Override enabled from profile
    //   - Override max_sessions from profile
    // For AI settings:
    //   - Update HONEYTRAP_AI_URL if profile.AI.OllamaURL is set
    // For logging:
    //   - If profile.Logging.PCAPCapture, note it (PCAP not yet implemented)
    return cfg
}
```

2. **`internal/app/app.go`** — Change `New()` to accept an optional profile name. If provided, load the profile and apply it to the config before creating the engine:
```go
func New(profileName string) (*App, error) {
    cfg, err := config.Load()
    if err != nil {
        return nil, err
    }
    
    if profileName != "" {
        profile, err := config.LoadProfile(profileName)
        if err != nil {
            fmt.Fprintf(os.Stderr, "honeytrap: warning: %v\n", err)
        } else {
            cfg = *config.ApplyProfile(&cfg, profile)
        }
    }
    // ... rest of init
}
```

3. **`internal/cli/root.go`** — Pass the profile name to `app.New()`:
```go
func (r *Runner) Run(ctx context.Context, args []string) error {
    profileName := ""
    if len(args) > 0 && args[0] == "deploy" {
        profileName = "default"
        if len(args) > 1 {
            profileName = args[1]
        }
    }
    application, err := app.New(profileName)
    // ...
}
```

4. **Remove the warning about "profiles are informational only"** from `root.go` help text and from `README.md` Known Limitations (since they'll now actually work).

### 🔴 CRITICAL-03: Bridge Go engine data to Fastify API

**Files:** `api/src/routes/sessions.ts`, `api/src/routes/events.ts`, plus new files

**Problem:** The Go honeypot engine writes sessions/events to JSONL files at `var/sessions.jsonl` and `var/events.jsonl`. The Fastify API reads from PostgreSQL. These are completely separate. No data flows between them, so the dashboard never sees honeypot data.

**Recommended approach — Option B (JSONL ingestion endpoint):**

Add a new route `POST /ingest` to the Fastify API that reads a JSONL file and inserts its records into PostgreSQL. This is simpler than adding a PostgreSQL driver to the Go binary.

1. **`api/src/routes/ingest.ts`** — New file:
```typescript
import fs from "fs";
import readline from "readline";

// POST /ingest — Bulk ingest JSONL sessions/events into PostgreSQL
// Accepts: { path: string, type: "sessions" | "events" }
// Reads the JSONL file line-by-line and upserts into the appropriate table
```

2. **`api/src/index.ts`** — Register the ingest route.

3. **`internal/app/app.go`** — Add a goroutine that periodically POSTs JSONL data to the API:
```go
// After engine starts, spawn a background goroutine that:
// 1. Reads new entries from var/sessions.jsonl and var/events.jsonl
// 2. POSTs them to the API /ingest endpoint
// 3. Marks ingested entries (simple offset tracking)
```

**Alternative approach — Option A (PostgreSQL writer in Go):**

Add `lib/pq` as a Go dependency and implement a `PostgresRepository` that actually works. This is cleaner but adds a CGO dependency. The stub at `internal/storage/postgres.go` exists but always returns `ErrPostgresDriverUnavailable`.

Pick whichever approach is simpler. Document the chosen approach in `README.md`.

### 🔴 CRITICAL-05: Add API authentication middleware

**Files:** New file `api/src/plugins/auth.ts`, `api/src/index.ts`

**Problem:** All API endpoints (`/sessions`, `/events`, `/tokens`, `/analytics`, `/ws`) require zero authentication. Anyone who reaches port 3000 can read all captured data.

**Implementation:**

1. **`api/src/plugins/auth.ts`** — New file implementing bearer token auth:
```typescript
import fp from "fastify-plugin";

const API_TOKEN = process.env.API_TOKEN || "";

const authPlugin = fp(async (app) => {
  // Skip auth if API_TOKEN is not set (dev mode)
  if (!API_TOKEN) return;

  app.addHook("onRequest", async (request, reply) => {
    // Skip /health endpoint
    if (request.url === "/health") return;
    
    const auth = request.headers.authorization;
    if (!auth || auth !== `Bearer ${API_TOKEN}`) {
      reply.code(401).send({ error: "Unauthorized" });
    }
  });
});

export default authPlugin;
```

2. **`api/src/index.ts`** — Register `authPlugin` before routes.
3. **`deploy/honeytrap-api.service`** — Add `Environment=API_TOKEN=%n` line with a note to set it.
4. **`docker-compose.yml`** — Add `API_TOKEN` to the api service env, with a strong default.
5. **`README.md`** — Document the `API_TOKEN` env var.
6. **`.env.example`** in `api/` — Add `API_TOKEN=changeme`.

### 🟡 HIGH-02: Fix HTTP Enhanced keep-alive handling

**File:** `internal/services/http_enhanced.go`

**Problem:** The `HandleConn` method for enhanced HTTP loops forever with `Connection: keep-alive`, ignoring `Connection: close` headers. The `SetDeadline` in the engine is absolute (30s from connection open), so even legitimate multi-page browsing gets killed after 30s total.

**Fix:**

1. Parse the request's `Connection` header. If `close`, respond with `Connection: close` and break the loop after sending the response.
2. In `engine.handleTCP()`, change from a single `SetDeadline` to using `SetReadDeadline` before each read. This allows per-request timeouts rather than per-connection.
3. Add a `KeepAlive` field to `SessionContext` (default `true`). The HTTP enhanced service can set it to `false` when `Connection: close` is received, and the engine loop can check it.

### 🟡 HIGH-04: Tighten seccomp profile

**File:** `docker/seccomp-honeytrap.json`

**Problem:** The seccomp profile allows `execve` (binary execution) and broad `clone` flags including namespace creation (`CLONE_NEWNS|CLONE_NEWUTS|CLONE_NEWIPC|CLONE_NEWNET|CLONE_NEWPID`). This enables container escape chains.

**Fix:**

1. Remove `execve` from the allowed syscall list. Go binaries on Alpine don't need it (they're statically compiled and run as PID 1).
2. Change the `clone`/`clone3` masked value from `2080505856` (0x7C020000 = CLONE_NEWNS|CLONE_NEWUTS|CLONE_NEWIPC|CLONE_NEWNET|CLONE_NEWPID) to `17` (0x11 = CLONE_VM|SIGCHLD — thread creation only).
3. Remove `fork` from the allowed list (Go doesn't fork at runtime).

### 🟡 HIGH-05: Wire Shutdown() into CLI signal handling

**Files:** `internal/cli/root.go`, `internal/app/app.go`

**Problem:** `Engine.Shutdown()` exists for graceful draining but is never called. When SIGINT is received, `signal.NotifyContext` cancels the context, causing `listener.Accept()` to return `net.ErrClosed`, which kills connections immediately rather than draining them.

**Fix:**

In `cli/root.go`, restructure the deploy command to call `engine.Shutdown()` on signal:

```go
func (r *Runner) deploy(ctx context.Context, profileName string) error {
    // ... profile loading ...
    
    runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
    defer stop()
    
    errCh := make(chan error, 1)
    go func() {
        errCh <- r.engine.Run(runCtx)
    }()
    
    select {
    case err := <-errCh:
        return err
    case <-runCtx.Done():
        fmt.Fprintln(os.Stderr, "honeytrap: shutting down gracefully...")
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()
        if err := r.engine.Shutdown(shutdownCtx); err != nil {
            return err
        }
        return <-errCh
    }
}
```

This requires `Engine` to be accessible from `Runner`. Currently `Runner` only holds `*Engine`. Add a `Shutdown()` method or expose the engine.

### 🟡 HIGH-06: Unify Token types

**Files:** `internal/models/models.go`, `internal/tokens/tokens.go`, `internal/e2e/e2e_test.go`

**Problem:** Two separate Token types exist. `models.Token` (with `FirstAccessedAt`, `LastAccessedAt`, `Metadata`) is used by storage and STIX export. `tokens.Token` (with `ID`, `Active`, `CreatedAt`, `Metadata`) is used by the generator. They have different `Kind` string values: `tokens.KindAWSCreds = "aws_credentials"` vs e2e test uses `"aws-creds"`.

**Fix:**

Option A (preferred): Replace `tokens.Token` with `models.Token`. Move `Generator` and `Store` into the `models` package or keep them in `tokens` but have them produce `models.Token`. Delete the duplicate struct.

Option B: Add an explicit converter:
```go
func ToModel(t Token) models.Token {
    return models.Token{
        ID:          t.ID,
        Name:        t.Name,
        Kind:        string(t.Kind),
        Value:       t.Value,
        Description: t.Description,
        // ... field mapping ...
    }
}
```

Also add unit tests that verify `tokens.Kind*` constants match the strings used throughout.

### 🟠 MEDIUM-01: UDP session deduplication

**File:** `internal/engine/engine.go`, `internal/services/udp.go`

**Problem:** Every UDP datagram creates a new session + 2 events (\~300 bytes of JSONL each). A 10KB/s UDP flood creates ~30 sessions/second, exhausting memory over hours.

**Fix:** In `engine.serveUDP()`, add a `sync.Map` keyed by `(service, IP)` that holds active session IDs. Before creating a new session, check if one already exists for that IP+service pair within a TTL (e.g., 5 minutes). If so, reuse the session and just emit a `udp.datagram` event instead of opening/closing.

### 🟠 MEDIUM-03: AI emulator authentication

**File:** `ai_emulator/server.py`

**Problem:** The `/ai-response` endpoint is unauthenticated. Anyone on the network can consume Ollama resources or poison the response cache.

**Fix:** Add a simple API key check:
```python
API_KEY = os.environ.get("HONEYTRAP_API_KEY", "")

@app.middleware("http")
async def auth_middleware(request: Request, call_next):
    if not API_KEY:
        return await call_next(request)  # No auth in dev mode
    if request.url.path in ["/ai/health", "/docs", "/openapi.json"]:
        return await call_next(request)
    auth = request.headers.get("Authorization", "")
    if auth != f"Bearer {API_KEY}":
        return JSONResponse(status_code=401, content={"error": "Unauthorized"})
    return await call_next(request)
```

### 🟠 MEDIUM-06: Add CORS to Fastify API

**Files:** `api/package.json`, `api/src/index.ts`

**Fix:**
```bash
cd api && npm install @fastify/cors
```

Then in `api/src/index.ts`:
```typescript
import cors from "@fastify/cors";
await app.register(cors, { origin: true }); // or restrict to dashboard origin
```

### 🟠 MEDIUM-07: Fix deprecated FastAPI `on_event`

**File:** `ai_emulator/server.py`

**Fix:** Replace:
```python
@app.on_event("shutdown")
async def shutdown():
    await emulator.close()
```

With lifespan:
```python
from contextlib import asynccontextmanager

@asynccontextmanager
async def lifespan(app: FastAPI):
    yield
    await emulator.close()

app = FastAPI(title="...", lifespan=lifespan)
```
Remove the `@app.on_event("shutdown")` decorator.

### 🟠 MEDIUM-08: Fix Redis HELLO response

**File:** `internal/services/redis.go`

**Problem:** The `HELLO` command returns an invalid RESP3 response. Redis clients will fail to parse it.

**Fix:** Change the HELLO response to decline RESP3 negotiation (which is the correct honeypot behavior — we only support RESP2):
```go
case "HELLO":
    return "-ERR NOPROTO sorry this protocol version is not supported\r\n"
```
This forces the client to fall back to RESP2, which is what our inline parser handles.

### 🟠 MEDIUM-10: Add Zod validation to tokens API route

**File:** `api/src/routes/tokens.ts`

**Fix:** Add Zod schemas for the token list and create endpoints:
```typescript
const tokenListQuerySchema = z.object({
  kind: z.string().optional(),
  active: z.enum(["true", "false"]).optional(),
});

const createTokenSchema = z.object({
  name: z.string().min(1),
  kind: z.string().min(1),
  value: z.string().min(1),
  description: z.string().optional(),
});
```

### 🟠 MEDIUM-11: Fix `Token.Kind` string inconsistencies

**Files:** `internal/e2e/e2e_test.go`, `internal/export/stix.go`

**Fix:** In `e2e_test.go`, change `Kind: "aws-creds"` to `Kind: string(tokens.KindAWSCreds)` or `"aws_credentials"`. In `stix.go`, import `tokens` package and use `tokens.KindAWSCreds` constants in descriptions.

Also add a shared test that verifies all `tokens.Kind*` constants produce valid values.

### 🟢 LOW-03: Move decoy files out of git

**Fix:** Add `decoys/` to `.gitignore` and add a `decoys/README.md` explaining how to generate them with `./honeytrap tokens` (once that CLI command exists).

### 🟢 LOW-06: Add mutex to `tokens.Store`

**File:** `internal/tokens/tokens.go`

**Fix:** Add `sync.RWMutex` to `Store`:
```go
type Store struct {
    mu     sync.RWMutex
    tokens map[string]Token
}

func (s *Store) Add(token Token) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.tokens[token.ID] = token
}
// ... etc for Get, GetByValue, List, RecordAccess, Deactivate
```

---

## Missing CLI Commands to Implement

### `tokens` command

**File:** `internal/cli/root.go`

Add a `tokens` subcommand that:
- `./honeytrap tokens list` — Lists all generated tokens
- `./honeytrap tokens generate <kind> [name]` — Generates a new token and prints it
- `./honeytrap tokens check <value>` — Looks up a token by value (for detecting access)

This requires making `tokens.Generator` and `tokens.Store` accessible from the CLI.

### `export` command  

**File:** `internal/cli/root.go`

Add an `export` subcommand that:
- `./honeytrap export stix` — Exports all sessions and tokens as STIX 2.1 bundles

This requires wiring `export.STIXExporter` to the CLI with an output directory flag.

---

## Test Coverage Gaps to Fill

Add unit tests for the following packages that currently have zero coverage:

1. **`internal/services/`** — Each of the 7 service emulators should have at least a basic `TestHandleConn` that creates a `net.Pipe()`, passes it to `HandleConn`, and verifies the response.
2. **`internal/engine/engine.go`** — Test `New()`, `Status()`, `canAccept()`.
3. **`internal/storage/memory.go`** — Test `CreateSession`, `CloseSession`, `RecordEvent`, `ListSessions`, `ListEvents` (including eviction caps).
4. **`internal/config/config.go`** — Test `Load()` with env vars, `ApplyProfile()`.
5. **`internal/alerts/alerts.go`** — Test alert creation and formatting (without actually sending to Slack/Telegram).

---

## File Map for Quick Reference

```
/Users/main/Security Apps/HONEYTRAP/
├── cmd/honeytrap/main.go              # CLI entry point
├── internal/
│   ├── app/app.go                     # App bootstrap (currently: env-only config)
│   ├── cli/root.go                    # CLI commands (deploy, status, sessions, events)
│   ├── config/config.go               # Config from env vars
│   ├── config/profile.go              # Deploy profile YAML loading + normalizeServiceName()
│   ├── engine/engine.go               # Core: TCP/UDP listeners, service dispatch, conn mgmt
│   ├── engine/session_manager.go      # Session/event tracking
│   ├── services/service.go            # Service interfaces (SessionContext, PacketContext)
│   ├── services/ssh.go                # Basic SSH honeypot
│   ├── services/ssh_enhanced.go       # Enhanced SSH (banner + kex capture)
│   ├── services/http.go               # Basic HTTP honeypot
│   ├── services/http_enhanced.go       # Enhanced HTTP (login pages, dashboard, API)
│   ├── services/ftp.go                # FTP honeypot (now logs USER/PASS events)
│   ├── services/redis.go              # Redis honeypot with RESP parser (now logs AUTH)
│   ├── services/udp.go                # UDP decoy
│   ├── storage/repository.go          # Repository interface
│   ├── storage/memory.go              # In-memory + JSONL (now 0o600 perms)
│   ├── storage/postgres.go            # Stub (always returns error)
│   ├── ai/client.go                  # HTTP client to AI emulator
│   ├── ai/responder.go               # AI response with fallback
│   ├── tokens/tokens.go              # Token generator + in-memory store
│   ├── tokens/tokens_test.go          # Token tests
│   ├── analysis/behavioral.go         # IsScripted, IsHuman, ClassifyTool, RiskScore
│   ├── analysis/behavioral_test.go    # Behavioral tests
│   ├── alerts/alerts.go               # Slack, Telegram, Email alert routing
│   ├── export/stix.go                 # STIX 2.1 export
│   ├── docker/spec.go                 # Container spec (unused stub)
│   ├── e2e/e2e_test.go               # End-to-end tests (now validates real files)
│   └── models/models.go              # Data models (Session, Event, Token, ServiceStatus)
├── api/src/                            # Fastify TypeScript API
│   ├── index.ts                       # App bootstrap (add auth + CORS here)
│   ├── plugins/db.ts                  # PostgreSQL connection pool
│   └── routes/
│       ├── sessions.ts                # GET /sessions
│       ├── events.ts                  # GET /events
│       ├── tokens.ts                  # CRUD for tokens + access recording
│       ├── analytics.ts              # GET /analytics (aggregated stats)
│       └── ws.ts                      # WebSocket broadcast
├── ai_emulator/                        # Python FastAPI AI emulator
│   ├── server.py                      # FastAPI app (add auth middleware, fix lifespan)
│   ├── emulator.py                    # Ollama client + cache + prompts
│   ├── test_emulator.py               # Pytest tests
│   └── requirements.txt               # Missing: pytest, pytest-asyncio
├── dashboard/src/                      # React + Vite + Tailwind + D3
│   ├── components/ (13 files)
│   ├── pages/ (5 files)
│   ├── hooks/ (useApi.ts, useWebSocket.ts)
│   └── App.tsx
├── profiles/                           # 5 deploy profiles (YAML)
├── decoys/                             # Fake credentials for planting
├── deploy/                             # Systemd services + install.sh
├── docker/seccomp-honeytrap.json       # Seccomp profile (needs tightening)
├── docker-compose.yml                  # All 7 service ports + AI on 8444
├── Dockerfile                          # Go binary build (all ports EXPOSEd)
├── BUG_LOG.md                          # Full bug inventory (40 issues)
└── README.md                           # Updated with known limitations
```

---

## Testing Protocol

After each fix, run:
```bash
cd "/Users/main/Security Apps/HONEYTRAP"
go build ./cmd/honeytrap && go vet ./... && go test ./... -count=1
```

All tests must pass before committing. Commit each major fix separately with a descriptive message referencing the BUG_LOG ID (e.g., "fix(C-01): wire deploy profiles to engine config").

When all fixes are done, push to `origin/main`.

---

## Priority Order

1. **C-01 + C-02**: Wire deploy profiles → engine config (these are the same fix)
2. **CRITICAL-05**: Add API authentication middleware
3. **HIGH-02**: Fix HTTP Enhanced keep-alive and per-request deadlines
4. **HIGH-04**: Tighten seccomp profile
5. **HIGH-05**: Wire Shutdown() into CLI signal handling
6. **HIGH-06**: Unify Token types
7. **CRITICAL-03**: Bridge Go engine → Fastify API (choose approach first)
8. **MEDIUM-01**: UDP session deduplication
9. **MEDIUM-03**: AI emulator auth middleware
10. **MEDIUM-06**: CORS on Fastify API
11. **MEDIUM-07**: Fix deprecated FastAPI on_event
12. **MEDIUM-08**: Fix Redis HELLO response
13. **MEDIUM-10+11**: Zod validation + Token.Kind consistency
14. **MEDIUM-04**: Restrict expandEnv to HONEYTRAP_* prefix
15. Add `tokens` and `export` CLI commands
16. Add unit tests for services, engine, storage, config, alerts
17. **LOW-03**: Move decoy files, **LOW-06**: Add mutex to Token.Store