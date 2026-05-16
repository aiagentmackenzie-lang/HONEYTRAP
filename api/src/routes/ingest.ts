import fs from "fs";
import readline from "readline";
import type { FastifyInstance, FastifyRequest, FastifyReply } from "fastify";
import { z } from "zod";

const ingestRequestSchema = z.object({
  path: z.string().min(1),
  type: z.enum(["sessions", "events"]),
});

const ingestRecordSchema = z.object({
  type: z.enum(["sessions", "events"]),
  record: z.record(z.unknown()),
});

interface IngestResult {
  type: string;
  ingested: number;
  errors: number;
  skipped: number;
}

export default async function ingestRoutes(app: FastifyInstance) {
  // POST /ingest — Bulk ingest JSONL file into PostgreSQL
  // Accepts: { path: string, type: "sessions" | "events" }
  app.post("/ingest", async (request: FastifyRequest, reply: FastifyReply) => {
    const parsed = ingestRequestSchema.safeParse(request.body);
    if (!parsed.success) {
      reply.code(400);
      return { error: "Invalid request body", details: parsed.error.errors };
    }
    const { path, type } = parsed.data;

    if (!fs.existsSync(path)) {
      reply.code(404);
      return { error: `File not found: ${path}` };
    }

    const result: IngestResult = { type, ingested: 0, errors: 0, skipped: 0 };

    const fileStream = fs.createReadStream(path);
    const rl = readline.createInterface({ input: fileStream, crlfDelay: Infinity });

    for await (const line of rl) {
      if (!line.trim()) continue;

      try {
        const record = JSON.parse(line);
        const outcome = await ingestRecord(app, type, record);
        if (outcome === "ingested") result.ingested++;
        else if (outcome === "skipped") result.skipped++;
      } catch {
        result.errors++;
      }
    }

    return result;
  });

  // POST /ingest-record — Ingest a single record from the Go engine bridge
  // Accepts: { type: "sessions" | "events", record: {...} }
  app.post("/ingest-record", async (request: FastifyRequest, reply: FastifyReply) => {
    const parsed = ingestRecordSchema.safeParse(request.body);
    if (!parsed.success) {
      reply.code(400);
      return { error: "Invalid request body", details: parsed.error.errors };
    }
    const { type, record } = parsed.data;

    try {
      const outcome = await ingestRecord(app, type, record);
      reply.code(outcome === "ingested" ? 201 : 200);
      return { status: outcome, type };
    } catch (err: any) {
      reply.code(500);
      return { error: err.message };
    }
  });
}

async function ingestRecord(app: FastifyInstance, type: string, record: any): Promise<string> {
  if (type === "sessions") {
    const { id, service, protocol, remote_addr, remote_ip, started_at, ended_at, metadata } = record;

    // Check if session already exists
    const existing = await app.db.query("SELECT id FROM sessions WHERE id = $1", [id]);
    if (existing.rows.length > 0) {
      // Update ended_at if provided and not already set
      if (ended_at) {
        await app.db.query(
          "UPDATE sessions SET ended_at = COALESCE(ended_at, $1) WHERE id = $2 AND ended_at IS NULL",
          [ended_at, id]
        );
      }
      return "skipped";
    }

    await app.db.query(
      `INSERT INTO sessions (id, service, protocol, remote_addr, remote_ip, started_at, ended_at, metadata)
       VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
       ON CONFLICT (id) DO UPDATE SET ended_at = COALESCE(sessions.ended_at, EXCLUDED.ended_at)`,
      [id, service, protocol, remote_addr, remote_ip, started_at, ended_at || null, JSON.stringify(metadata || {})]
    );
    return "ingested";
  } else if (type === "events") {
    const { id, session_id, service, type: eventType, remote_addr, payload, occurred_at } = record;

    await app.db.query(
      `INSERT INTO events (id, session_id, service, type, remote_addr, payload, occurred_at)
       VALUES ($1, $2, $3, $4, $5, $6, $7)
       ON CONFLICT (id) DO NOTHING`,
      [id, session_id, service, eventType, remote_addr, JSON.stringify(payload || {}), occurred_at]
    );
    return "ingested";
  }

  return "skipped";
}