BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = timezone('utc', now());
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  service TEXT NOT NULL,
  protocol TEXT NOT NULL CHECK (protocol IN ('tcp', 'udp')),
  remote_ip INET NOT NULL,
  remote_addr TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now()),
  ended_at TIMESTAMPTZ,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  service TEXT NOT NULL,
  event_type TEXT NOT NULL,
  remote_addr TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now())
);

CREATE TABLE IF NOT EXISTS tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  kind TEXT NOT NULL,
  value TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  first_accessed_at TIMESTAMPTZ,
  last_accessed_at TIMESTAMPTZ,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now()),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now())
);

DROP TRIGGER IF EXISTS tokens_set_updated_at ON tokens;
CREATE TRIGGER tokens_set_updated_at
BEFORE UPDATE ON tokens
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_sessions_started_at_desc
  ON sessions (started_at DESC);

CREATE INDEX IF NOT EXISTS idx_sessions_remote_ip_started_at
  ON sessions (remote_ip, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_sessions_service_started_at
  ON sessions (service, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_events_session_id_occurred_at
  ON events (session_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_events_service_occurred_at
  ON events (service, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_events_event_type_occurred_at
  ON events (event_type, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_tokens_active_kind
  ON tokens (kind, created_at DESC)
  WHERE is_active = TRUE;

COMMIT;
