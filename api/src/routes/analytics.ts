import { FastifyPluginCallback } from "fastify";

const analyticsPlugin: FastifyPluginCallback = async (app, _opts) => {
  // GET /analytics — aggregated attack data
  app.get("/analytics", async (req, reply) => {
    const db = app.db;

    try {
      // Top attacker IPs
      const topIps = await db.query(
        `SELECT source_ip, COUNT(*) as session_count, 
                COUNT(DISTINCT service) as services_targeted,
                MAX(started_at) as last_seen
         FROM sessions 
         GROUP BY source_ip 
         ORDER BY session_count DESC 
         LIMIT 20`
      );

      // Service breakdown
      const serviceBreakdown = await db.query(
        `SELECT service, COUNT(*) as attack_count
         FROM sessions 
         GROUP BY service 
         ORDER BY attack_count DESC`
      );

      // Attack timeline (hourly, last 48h)
      const timeline = await db.query(
        `SELECT DATE_TRUNC('hour', started_at) as hour, COUNT(*) as count
         FROM sessions 
         WHERE started_at > NOW() - INTERVAL '48 hours'
         GROUP BY hour 
         ORDER BY hour ASC`
      );

      // Token stats
      const tokenStats = await db.query(
        `SELECT 
           COUNT(*) FILTER (WHERE status = 'active') as active_tokens,
           COUNT(*) FILTER (WHERE status = 'triggered') as triggered_tokens,
           COUNT(*) FILTER (WHERE status = 'deactivated') as deactivated_tokens,
           SUM(access_count) as total_accesses
         FROM tokens`
      );

      // Recent alerts
      const recentAlerts = await db.query(
        `SELECT tal.*, t.kind as token_kind
         FROM token_access_log tal
         JOIN tokens t ON t.id = tal.token_id
         ORDER BY tal.accessed_at DESC
         LIMIT 50`
      );

      return {
        topIps: topIps.rows,
        serviceBreakdown: serviceBreakdown.rows,
        timeline: timeline.rows.map((r: any) => ({
          hour: r.hour,
          count: Number(r.count),
        })),
        tokenStats: tokenStats.rows[0] || { active_tokens: 0, triggered_tokens: 0, deactivated_tokens: 0, total_accesses: 0 },
        recentAlerts: recentAlerts.rows,
      };
    } catch (err: any) {
      req.log.error({ err }, "Analytics query failed");
      return reply.code(500).send({ error: "Analytics query failed", details: err.message });
    }
  });
};

export default analyticsPlugin;