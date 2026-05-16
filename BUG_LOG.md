# HONEYTRAP — Code Quality & Security Bug Log

**Reviewer:** Lead Code Quality Reviewer & Lead Security Developer  
**Date:** 2026-05-16  
**Scope:** Full codebase review — Go core, Python AI emulator, TypeScript API, deployment configs, README accuracy  
**Method:** Systematic file-by-file review + `go vet` + `go test` + build verification  

---

## Summary

| Category | Critical | High | Medium | Low | Total |
|----------|----------|------|--------|-----|-------|
| Security | 2 | 3 | 4 | 2 | 11 |
| Correctness | 3 | 4 | 3 | 0 | 10 |
| Design | 0 | 2 | 4 | 2 | 8 |
| README/Docs | 0 | 5 | 4 | 2 | 11 |
| **Total** | **5** | **14** | **15** | **6** | **40** |

---

## CRITICAL Issues

### C-01: Deploy profiles have no effect on running services
**File:** `internal/cli/root.go`, `internal/app/app.go`, `internal/engine/engine.go`  
**Severity:** Critical (Correctness)  

The `deploy` command loads a YAML profile and prints its contents, but **never applies it to the engine**. The `App.New()` creates `Config` via `config.Load()` *before* `deploy` is called, and the engine is initialized with those defaults. The profile's port overrides, service toggles, and AI settings are completely ignored.

```go
// app.go — Engine created BEFORE deploy is called
core := engine.New(cfg, repo)  // uses config.Load() defaults
return &App{runner: cli.NewRunner(core)}, nil

// root.go — Profile loaded AFTER engine is created, never applied
profile, err := config.LoadProfile(profileName)  // only printed, not used
```

**Impact:** All 5 deploy profiles (`default`, `minimal`, `full-spectrum`, `raspberry-pi`, `corporate-internal`) are decorative. Running `./honeytrap deploy minimal` still starts all 7 services.

**Fix:** The `deploy` command must construct a modified `Config` from the profile and pass it to the engine before `Run()`.

---

### C-02: Profile service keys don't match engine service names
**File:** `profiles/*.yml`, `internal/engine/engine.go`  
**Severity:** Critical (Correctness)  

YAML profiles use underscored keys (`ssh_enhanced`, `http_enhanced`, `udp`) while the engine registers hyphenated names (`ssh-enhanced`, `http-enhanced`, `udp-decoy`). Even if profile config were applied, service name lookups would fail silently.

| YAML Key | Engine Key |
|----------|-----------|
| `ssh_enhanced` | `ssh-enhanced` |
| `http_enhanced` | `http-enhanced` |
| `udp` | `udp-decoy` |

**Fix:** Standardize on hyphenated names across all YAML files and Go code, or add a normalization function.

---

### C-03: Go engine and Fastify API share no data
**File:** `internal/app/app.go`, `api/src/routes/sessions.ts`  
**Severity:** Critical (Architecture)  

The Go honeypot engine writes sessions/events to in-memory JSONL files. The Fastify API server reads from PostgreSQL. **These are completely separate data paths with no bridge.** The architecture diagram shows a unified flow, but in reality:

- Go engine → `var/sessions.jsonl` + `var/events.jsonl` (flat files)
- Fastify API → PostgreSQL tables (separate database)

A session captured by the honeypot engine will **never appear** in the dashboard.

**Fix:** Either: (a) add a PostgreSQL writer to the Go engine, (b) add a JSONL ingestion endpoint to the Fastify API, or (c) document that the Go binary and API must be used independently.

---

### C-04: Captured credentials stored in plaintext accessible events
**File:** `internal/services/http_enhanced.go`, `internal/storage/memory.go`  
**Severity:** Critical (Security)  

The HTTP enhanced service captures attacker credentials and stores them in plaintext in event payloads:

```go
event["captured_username"] = username
event["captured_password"] = parsed.Get("password")
```

These are written to `var/events.jsonl` with `0o644` permissions (world-readable). Any local user on the system can read captured passwords.

**Fix:** (a) Encrypt credential fields at rest, (b) Restrict JSONL file permissions to `0o600`, (c) Redact passwords in API responses.

---

### C-05: Dashboard API and WebSocket have zero authentication
**Files:** `api/src/routes/*.ts`, `api/src/routes/ws.ts`  
**Severity:** Critical (Security)  

All API endpoints (`/sessions`, `/events`, `/tokens`, `/analytics`, `/ws`) require no authentication. Anyone who can reach port 3000 can:
- Read all captured session data including credentials
- Create/delete honeytokens
- Subscribe to real-time WebSocket event stream
- Access analytics with attacker IPs and patterns

**Fix:** Add bearer token authentication middleware. For production, bind to `127.0.0.1` at minimum.

---

## HIGH Issues

### H-01: Race condition on `engine.listeners` and `engine.packetConns` slices
**File:** `internal/engine/engine.go` lines ~110-150  
**Severity:** High (Concurrency)  

`serveTCP()` and `serveUDP()` are launched as goroutines that `append()` to `e.listeners` and `e.packetConns`. `Shutdown()` reads these slices. No mutex protects these operations, creating a data race.

```go
// goroutine:
e.listeners = append(e.listeners, listener)

// Shutdown():
for _, l := range e.listeners {  // RACE
    _ = l.Close()
}
```

**Fix:** Use a mutex or channel to pass listeners back to the main goroutine.

---

### H-02: HTTP Enhanced service `Connection: keep-alive` loop hangs
**File:** `internal/services/http_enhanced.go`  
**Severity:** High (Correctness)  

The enhanced HTTP service always responds with `Connection: keep-alive` and loops reading requests. If a client sends `Connection: close`, the service ignores it and continues waiting. Combined with the `SetDeadline` of 30s from `SessionContext` (but the deadline is set in the engine to 30s, while the HTTP enhanced handler has no explicit timeout for the loop), an attacker can hold connections open for 60 seconds.

Additionally, the `SetDeadline` in `engine.handleTCP` sets a 30s deadline, but the HTTP enhanced service reads multiple requests in a loop — the deadline applies to the *connection*, so the first successful read resets the timer... but actually `SetDeadline` is an absolute deadline, not idle timeout. This means the connection will be force-killed after 30s regardless, even if the client is legitimately browsing.

**Fix:** Use `SetReadDeadline` per-request instead of `SetDeadline` per-connection, and respect `Connection: close` headers.

---

### H-03: Redis AUTH always succeeds — no credential tracking
**File:** `internal/services/redis.go`  
**Severity:** High (Security Feature)  

`AUTH` always returns `+OK`, which is correct for a honeypot. However, the AUTH credentials are not captured as events. An attacker authenticating to Redis with stolen credentials should have those credentials logged.

**Fix:** Log AUTH commands as `redis.auth` events with the provided password (redacted or hashed for storage).

---

### H-04: Seccomp profile allows `execve` and `fork`
**File:** `docker/seccomp-honeytrap.json`  
**Severity:** High (Security)  

The seccomp whitelist includes `execve`, `fork`, `clone`, and `clone3`. For a honeypot that should be maximally restrictive:

- `execve` allows any binary execution inside the container
- `fork`/`clone` allow process creation, enabling potential container escape chains

The `clone` rule has a masked equality check, but `2080505856` (0x7C020000) includes `CLONE_NEWNS|CLONE_NEWUTS|CLONE_NEWIPC|CLONE_NEWNET|CLONE_NEWPID`, which allows namespace creation — the opposite of what you want.

**Fix:** Remove `execve` (the Go binary is statically compiled and doesn't need it), and restrict `clone`/`clone3` to only allow thread creation flags (`SIGCHLD`).

---

### H-05: `Shutdown()` method is defined but never called by CLI
**File:** `internal/engine/engine.go`, `internal/cli/root.go`  
**Severity:** High (Correctness)  

The `Engine.Shutdown()` method exists for graceful connection draining, but the CLI's `deploy` command only cancels the context (via `signal.NotifyContext`). When context is cancelled, `listener.Accept()` returns `net.ErrClosed`, which causes `serveTCP` to return `nil` (not calling `Shutdown()`). This means active connections are NOT gracefully drained despite the documented 10-second drain timeout in `Run()`.

**Fix:** Call `engine.Shutdown(ctx)` in the CLI before `Run()` completes.

---

### H-06: Duplicate `Token` types with no bridge
**Files:** `internal/models/models.go`, `internal/tokens/tokens.go`  
**Severity:** High (Design)  

There are two completely separate Token systems:
- `models.Token` — used by storage, STIX export, and API
- `tokens.Token` — used by the generator

They have different field sets, different `Kind` values (`"aws_credentials"` vs `"aws-creds"` in tests), and no conversion function. Tokens generated by `tokens.Generator` cannot be persisted or exported via STIX without manual field-by-field mapping that doesn't exist.

**Fix:** Unify into a single `Token` type, or add explicit mappers between `tokens.Token` → `models.Token`.

---

## MEDIUM Issues

### M-01: UDP service creates a session per datagram (no deduplication)
**File:** `internal/services/udp.go`, `internal/engine/engine.go`  
**Severity:** Medium (DoS)  

Each UDP datagram creates a new session and two events (opened + closed). A 1KB/s UDP flood would create thousands of sessions, exhausting memory. There's no IP-based deduplication or rate limiting.

**Fix:** Cache active UDP sessions by source IP and reuse them within a TTL window.

---

### M-02: JSONL file permissions are world-readable
**File:** `internal/storage/memory.go`  
**Severity:** Medium (Security)  

`appendJSONL` creates files with mode `0o644` (owner read/write, group/other read). Since captured credentials are stored in these files, any local user can read them.

**Fix:** Change to `0o600`.

---

### M-03: AI emulator has no authentication
**File:** `ai_emulator/server.py`  
**Severity:** Medium (Security)  

The `/ai-response` endpoint accepts unauthenticated requests. Anyone on the network can:
- Generate honeypot responses (information disclosure about service templates)
- Consume Ollama compute resources
- Poison the response cache

**Fix:** Add API key validation middleware.

---

### M-04: `expandEnv` allows arbitrary environment variable expansion in profiles
**File:** `internal/config/profile.go`  
**Severity:** Medium (Security)  

The `expandEnv` function calls `os.ExpandEnv` on profile values, which expands any `${VAR}` in the YAML. If a profile file contains `${AWS_SECRET_ACCESS_KEY}` or other env vars, they'll be expanded and potentially leaked.

**Fix:** Only expand variables matching a whitelist pattern (e.g., `HONEYTRAP_*`).

---

### M-05: `app.New()` prints warning about `HONEYTRAP_DATABASE_URL` but the field is never used
**File:** `internal/app/app.go`  
**Severity:** Medium (UX)  

```go
if cfg.DatabaseURL != "" {
    fmt.Fprintln(os.Stderr, "honeytrap: warning: HONEYTRAP_DATABASE_URL is set but the Go binary uses in-memory storage...")
}
```

This warning fires every time the app starts, but `PostgresRepository` is a stub that always returns `ErrPostgresDriverUnavailable`. The PostgreSQL integration exists only in the Fastify API.

**Fix:** Remove the `DatabaseURL` field from `Config`, or add a proper `--database` flag.

---

### M-06: No CORS configuration on Fastify API
**File:** `api/src/index.ts`  
**Severity:** Medium (Functionality)  

The Fastify API has no CORS plugin. Browser-based dashboard at `localhost:5173` making requests to `localhost:3000` will be blocked by same-origin policy.

**Fix:** Install `@fastify/cors` and register it in the API.

---

### M-07: `ai_emulator/server.py` uses deprecated FastAPI `on_event` handler
**File:** `ai_emulator/server.py` line ~26  
**Severity:** Medium (Deprecation)  

```python
@app.on_event("shutdown")
async def shutdown():
    await emulator.close()
```

`on_event` is deprecated in FastAPI. Should use the lifespan pattern.

**Fix:** Refactor to use `FastAPI(lifespan=...)`.

---

### M-08: Redis HELLO response is invalid RESP3
**File:** `internal/services/redis.go`  
**Severity:** Medium (Correctness)  

The HELLO command returns a manually constructed RESP3 map `%7\r\n...` that doesn't conform to the RESP3 specification. redis-cli and Redis client libraries will fail to parse this response.

**Fix:** Either respond with a RESP2-compatible error (`-ERR NOPROTO sorry this protocol version is not supported`) or construct a valid RESP3 map.

---

### M-09: Profile `deploy` command doesn't actually configure services
**File:** `internal/cli/root.go`  
**Severity:** Medium (UX) — see C-01 for details. The CLI prints profile details but doesn't reconfigure the engine.

---

### M-10: Fastify API tokens route uses raw SQL with no input sanitization on table references
**File:** `api/src/routes/tokens.ts`  
**Severity:** Medium (Security)  

The tokens route constructs SQL queries with user-provided `kind` filter directly in WHERE clauses. While Zod validates `kind` as a string in other routes, the tokens GET endpoint uses `request.query as { kind?, active? }` without Zod validation. The `kind` value is interpolated as `$N` parameters, which is safe from injection, but there's no validation that `kind` is a valid token type.

---

### M-11: `Token.Kind` value inconsistencies
**Files:** `internal/tokens/tokens.go`, `internal/e2e/e2e_test.go`  
**Severity:** Medium (Correctness)  

The `tokens` package defines `KindAWSCreds = "aws_credentials"`, but e2e tests create `models.Token` with `Kind: "aws-creds"`. These are different strings and will never match if filtering by kind.

**Fix:** Use the `tokens.Kind*` constants everywhere, or define a shared `Kind` enum.

---

## LOW Issues

### L-01: Docker Compose missing service ports
**File:** `docker-compose.yml`  
**Severity:** Low (Configuration)  

The `honeytrap` service only exposes ports 2222, 8080, 2121, 9161/udp. Missing: 2223 (SSH Enhanced), 8443 (HTTP Enhanced), 6379 (Redis). The README documents all 7 services but Docker Compose only routes 4.

---

### L-02: Dockerfile EXPOSE missing ports
**File:** `Dockerfile`  
**Severity:** Low (Configuration)  

`EXPOSE 2222 8080 2121 9161/udp` is missing 2223, 8443, and 6379.

---

### L-03: Decoy files in git repo
**Files:** `decoys/*`  
**Severity:** Low (Security)  

Realistic-looking fake credentials are committed to the repository. While labeled "FAKE", they could be mistakenly used or trigger security scanners. Consider `.gitignore`-ing them and generating on deploy.

---

### L-04: `e2e_test.go:SystemdServices` test is a no-op
**File:** `internal/e2e/e2e_test.go`  
**Severity:** Low (Test Quality)  

The test does `http.Get("deploy/honeytrap.service")` which always fails (it's a path, not a URL), and the test discards the error and passes regardless.

---

### L-05: `e2e_test.go:SeccompProfile` test doesn't validate the real seccomp file
**File:** `internal/e2e/e2e_test.go`  
**Severity:** Low (Test Quality)  

The test creates a dummy seccomp JSON object instead of reading and validating `docker/seccomp-honeytrap.json`.

---

### L-06: `tokens.Store` has no concurrency protection
**File:** `internal/tokens/tokens.go`  
**Severity:** Low (Race)  

The `Store` uses a plain `map[string]Token` with no mutex. If used from multiple goroutines (e.g., HTTP handlers), it will race. Currently it's only used in tests, but the public API suggests concurrent use.

---

## README / Documentation Inaccuracies

### D-01: Non-existent `tokens` CLI command
The README lists `./honeytrap tokens` as a command. It does not exist. The CLI only has: `deploy`, `profiles`, `status`, `sessions`, `events`, `version`, `help`.

### D-02: Non-existent `export` CLI command  
The SPEC and README reference `honeytrap export` for STIX export. The STIX export code exists in Go but has no CLI command or API endpoint to invoke it.

### D-03: Non-existent `replay` CLI command
The SPEC mentions `honeytrap replay <id>` for session replay. Not implemented.

### D-04: Port conflict between AI Emulator and HTTP Enhanced service
The README documents the AI Emulator on port 8443 (`python server.py 8443`), which conflicts with the HTTP Enhanced honeypot also configured on port 8443.

### D-05: LOC claims are inflated
README claims `~8,500+` and `~10,500 LOC`. Actual source code (Go + Python + TS + SQL + YAML + Shell) totals ~6,800 lines. Only reaches "10,500" if counting dashboard framework code and generated artifacts.

### D-06: Component count slightly off
README says "12 Components" but there are actually 13 (including `Sidebar.tsx`).

### D-07: Architecture diagram is misleading
The architecture shows a unified data flow through "Fastify API Server", but the Go engine and API are completely disconnected (see C-03).

### D-08: Version number inconsistency
CLI says `v0.4.0` in `root.go`, AI emulator says `v0.2.0`, API health endpoint says `v0.2.0`. README doesn't mention version.

### D-09: `docker-compose up -d` command doesn't work as-is
The compose file references `./db/schema.sql` for PostgreSQL init, but doesn't create the `var/` directory the Go binary needs.

### D-10: AI emulator `python server.py 8443` default port conflicts with HTTP Enhanced
See D-04.

### D-11: Python `pytest` required but not documented
The AI emulator tests use `pytest` and `pytest-asyncio` but `requirements.txt` only lists `fastapi`, `uvicorn`, `httpx`, `pydantic`, `jinja2`.

---

## Test Results

```
✅ go build ./cmd/honeytrap       — PASSES
✅ go vet ./...                    — PASSES (no issues)
✅ go test ./... — 17 tests total — ALL PASS
   ├── internal/analysis           — 9 tests PASS
   ├── internal/e2e               — 7 tests PASS (2 are no-ops)
   └── internal/tokens            — 9 tests PASS
```

**Coverage gaps:** No unit tests for:
- `internal/services/*` (7 service emulators — zero test coverage)
- `internal/engine/engine.go` (core engine — zero test coverage)
- `internal/storage/memory.go` (storage layer — zero test coverage)
- `internal/ai/client.go`, `internal/ai/responder.go` (AI client — zero test coverage)
- `internal/alerts/alerts.go` (alert routing — only tested via e2e)
- `internal/export/stix.go` (only tested via e2e)
- `internal/cli/root.go` (CLI — zero test coverage)
- `internal/config/*` (config loading — zero test coverage)

---

## Recommended Priority fixes

1. **C-01 + C-02 + H-05**: Wire deploy profiles to engine config (architectural gap)
2. **C-03**: Bridge Go engine data → Fastify API (or document separation clearly)
3. **C-04 + C-05**: Add auth middleware to API + restrict file permissions
4. **H-01**: Add mutex for listener/packetConn slices
5. **H-02**: Fix HTTP enhanced Keep-Alive handling and per-request deadlines
6. **H-04**: Tighten seccomp profile (remove execve, restrict clone)
7. **D-01, D-02, D-03**: Add missing CLI commands or remove from docs