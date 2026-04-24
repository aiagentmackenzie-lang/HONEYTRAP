import type { FastifyInstance, FastifyRequest, FastifyReply } from "fastify";

interface Token {
  id: string;
  name: string;
  kind: string;
  value: string;
  description: string;
  active: boolean;
  first_accessed_at: string | null;
  last_accessed_at: string | null;
  created_at: string;
  metadata: Record<string, unknown>;
}

interface TokenAccessLog {
  id: string;
  token_id: string;
  source_ip: string;
  user_agent: string | null;
  accessed_at: string;
  metadata: Record<string, unknown>;
}

export default async function tokensRoutes(app: FastifyInstance) {
  // GET /tokens — List all honeytokens
  app.get("/tokens", async (request: FastifyRequest, reply: FastifyReply) => {
    const { kind, active } = request.query as { kind?: string; active?: string };
    
    // Build query with combined filters instead of overwriting
    const conditions: string[] = [];
    const params: any[] = [];
    let paramIdx = 1;
    
    if (kind) {
      conditions.push(`kind = $${paramIdx++}`);
      params.push(kind);
    }
    if (active === "true") {
      conditions.push(`is_active = true`);
    }
    
    const where = conditions.length > 0 ? `WHERE ${conditions.join(" AND ")}` : "";
    const result = await app.db.query(
      `SELECT * FROM tokens ${where} ORDER BY created_at DESC`,
      params
    );
    return result.rows;
  });

  // POST /tokens — Create a new honeytoken
  app.post("/tokens", async (request: FastifyRequest, reply: FastifyReply) => {
    const body = request.body as Partial<Token>;
    if (!body.name || !body.kind || !body.value) {
      reply.code(400);
      return { error: "name, kind, and value are required" };
    }
    const result = await app.db.query(
      `INSERT INTO tokens (name, kind, value, description, is_active)
       VALUES ($1, $2, $3, $4, true)
       RETURNING *`,
      [body.name, body.kind, body.value, body.description || ""]
    );
    reply.code(201);
    return result.rows[0];
  });

  // GET /tokens/:id — Get a specific token
  app.get("/tokens/:id", async (request: FastifyRequest, reply: FastifyReply) => {
    const { id } = request.params as { id: string };
    const result = await app.db.query("SELECT * FROM tokens WHERE id = $1", [id]);
    if (result.rows.length === 0) {
      reply.code(404);
      return { error: "token not found" };
    }
    return result.rows[0];
  });

  // POST /tokens/:id/access — Record token access (triggers alert)
  app.post("/tokens/:id/access", async (request: FastifyRequest, reply: FastifyReply) => {
    const { id } = request.params as { id: string };
    const body = request.body as { source_ip?: string; user_agent?: string; metadata?: Record<string, unknown> };

    // Check token exists
    const tokenResult = await app.db.query("SELECT * FROM tokens WHERE id = $1", [id]);
    if (tokenResult.rows.length === 0) {
      reply.code(404);
      return { error: "token not found" };
    }

    // Record access
    const now = new Date().toISOString();
    await app.db.query(
      `UPDATE tokens SET last_accessed_at = $1, first_accessed_at = COALESCE(first_accessed_at, $1)
       WHERE id = $2`,
      [now, id]
    );

    await app.db.query(
      `INSERT INTO token_access_log (token_id, source_ip, user_agent, accessed_at, metadata)
       VALUES ($1, $2, $3, $4, $5)`,
      [id, body.source_ip || "unknown", body.user_agent || null, now, JSON.stringify(body.metadata || {})]
    );

    return {
      alert: true,
      token_id: id,
      message: "Token access detected — potential intruder!",
      source_ip: body.source_ip,
      accessed_at: now,
    };
  });

  // DELETE /tokens/:id — Deactivate a token
  app.delete("/tokens/:id", async (request: FastifyRequest, reply: FastifyReply) => {
    const { id } = request.params as { id: string };
    await app.db.query("UPDATE tokens SET is_active = false WHERE id = $1", [id]);
    return { status: "deactivated", id };
  });
}
