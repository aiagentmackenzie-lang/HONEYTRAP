import Fastify from "fastify";
import dbPlugin from "./plugins/db.js";
import sessionsRoute from "./routes/sessions.js";
import eventsRoute from "./routes/events.js";

async function buildServer() {
  const app = Fastify({
    logger: {
      level: process.env.LOG_LEVEL ?? "info"
    }
  });

  await app.register(dbPlugin);
  await app.register(sessionsRoute);
  await app.register(eventsRoute);

  app.get("/health", async () => ({ status: "ok" }));

  return app;
}

async function start() {
  const app = await buildServer();
  const port = Number(process.env.PORT ?? 3000);
  const host = process.env.HOST ?? "0.0.0.0";

  try {
    await app.listen({ host, port });
  } catch (error) {
    app.log.error(error);
    process.exit(1);
  }
}

start();
