# HONEYTRAP

AI-powered deception framework for capturing attacker behavior through believable low-interaction services, structured session telemetry, and a clean backend/API split.

## Phase 1 Scope

Phase 1 delivers the core engine:

- Go honeypot runtime with TCP and UDP listener support
- Service emulators for SSH, HTTP, and FTP
- Session and event capture with timestamps, source addresses, and protocol details
- PostgreSQL schema for sessions, events, and tokens
- Fastify API scaffold with `GET /sessions` and `GET /events`
- CLI workflow with `deploy`, `status`, `sessions`, `events`, and `version`
- Docker and Compose assets for local demo deployment

## Layout

```text
cmd/honeytrap          Go CLI entrypoint
internal/app           App wiring
internal/cli           Command handlers
internal/config        Environment-driven configuration
internal/engine        Listener engine and session manager
internal/services      SSH, HTTP, FTP, and UDP decoy services
internal/storage       Storage abstraction and local JSONL repository
db/schema.sql          PostgreSQL schema
api/                   Fastify TypeScript API
docs/                  Architecture notes
```

## Running The Core Engine

```bash
go build ./cmd/honeytrap
./honeytrap status
./honeytrap deploy default
```

Default listeners:

- `2222/tcp` SSH emulator
- `8080/tcp` HTTP emulator
- `2121/tcp` FTP emulator
- `9161/udp` UDP decoy listener

Captured sessions and events are stored locally under `var/` in JSONL format during offline development builds.

## API

The API lives in [`api`](./api) and is designed to run against PostgreSQL using `db/schema.sql`.

```bash
cd api
npm install
npm run dev
```

Endpoints:

- `GET /sessions`
- `GET /events`
- `GET /health`

## Docker

```bash
docker compose up --build
```

This brings up:

- PostgreSQL with schema initialization
- The Go core engine
- The Fastify API service

## Notes

- The Go runtime is dependency-light and compiles with the standard library.
- In this sandboxed offline environment, PostgreSQL and Fastify packages are scaffolded but not installed locally.
- The schema and API are ready for runtime wiring once dependencies are installed in a connected environment.
