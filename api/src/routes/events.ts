import type { FastifyPluginAsync } from "fastify";
import { z } from "zod";

const querySchema = z.object({
  limit: z.coerce.number().int().min(1).max(500).default(250)
});

const eventsRoute: FastifyPluginAsync = async (app) => {
  app.get("/events", async (request) => {
    const { limit } = querySchema.parse(request.query);
    const result = await app.db.query(
      `select
        id,
        session_id,
        service,
        event_type,
        remote_addr,
        payload,
        occurred_at
      from events
      order by occurred_at desc
      limit $1`,
      [limit]
    );

    return { data: result.rows, count: result.rowCount ?? result.rows.length };
  });
};

export default eventsRoute;
