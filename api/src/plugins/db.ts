import fp from "fastify-plugin";
import { Pool } from "pg";

const dbPlugin = fp(async (app) => {
  const connectionString = process.env.DATABASE_URL;
  if (!connectionString) {
    throw new Error("DATABASE_URL is required");
  }

  const pool = new Pool({
    connectionString,
    max: 10,
    idleTimeoutMillis: 30_000,
    connectionTimeoutMillis: 5_000
  });

  await pool.query("select 1");
  app.decorate("db", pool);

  app.addHook("onClose", async () => {
    await pool.end();
  });
});

export default dbPlugin;
