import type { FastifyPluginAsync } from "fastify";
import { z } from "zod";

const querySchema = z.object({
  limit: z.coerce.number().int().min(1).max(500).default(100)
});

const sessionsRoute: FastifyPluginAsync = async (app) => {
  app.get("/sessions", async (request) => {
    const { limit } = querySchema.parse(request.query);
    const result = await app.db.query(
      `select
        id,
        service,
        protocol,
        host(remote_ip) as remote_ip,
        remote_addr,
        started_at,
        ended_at,
        metadata
      from sessions
      order by started_at desc
      limit $1`,
      [limit]
    );

    return { data: result.rows, count: result.rowCount ?? result.rows.length };
  });
};

export default sessionsRoute;
