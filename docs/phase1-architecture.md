# HONEYTRAP Phase 1 Architecture

## Core Engine

The Go core exposes a low-interaction deception engine with:

- TCP listeners for SSH, HTTP, and FTP emulation
- UDP listener support through a generic decoy responder
- Session lifecycle tracking with event capture
- CLI-based deployment and inspection workflow

The engine uses a storage interface so the listener layer is isolated from persistence concerns. In this offline build, the Go runtime uses append-only JSONL logs under `var/` for deterministic local development. The accompanying PostgreSQL schema and Fastify API are prepared for the persistent backend planned by the Phase 1 architecture.

## Data Flow

1. A connection or datagram hits a configured listener.
2. The engine opens a session record with service, protocol, timestamps, and source address metadata.
3. The service emulator records protocol-specific events such as FTP commands or HTTP request metadata.
4. The session closes and the CLI or API can later inspect the captured data.

## Container Support

- Root `Dockerfile` builds the Go engine into a minimal Alpine runtime image.
- `api/Dockerfile` builds the Fastify service.
- `docker-compose.yml` provisions PostgreSQL, the core engine, and the API together for portfolio demos.
