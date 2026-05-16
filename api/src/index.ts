import Fastify from "fastify";
import cors from "@fastify/cors";
import dbPlugin from "./plugins/db.js";
import authPlugin from "./plugins/auth.js";
import sessionsRoute from "./routes/sessions.js";
import eventsRoute from "./routes/events.js";
import tokensRoute from "./routes/tokens.js";
import wsPlugin from "./routes/ws.js";
import analyticsPlugin from "./routes/analytics.js";
import ingestRoutes from "./routes/ingest.js";

async function buildServer() {
  const app = Fastify({
    logger: {
      level: process.env.LOG_LEVEL ?? "info"
    }
  });

  // Register CORS (allow dashboard origin)
  await app.register(cors, { origin: true });

  // Register auth middleware (bearer token if API_TOKEN is set)
  await app.register(authPlugin);

  // Register database plugin
  await app.register(dbPlugin);

  // Register routes
  await app.register(sessionsRoute);
  await app.register(eventsRoute);
  await app.register(tokensRoute);
  await app.register(wsPlugin);
  await app.register(analyticsPlugin);
  await app.register(ingestRoutes);

  app.get("/health", async () => ({ status: "ok", version: "0.2.0" }));

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