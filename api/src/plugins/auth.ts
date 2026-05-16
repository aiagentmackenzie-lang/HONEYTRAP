import fp from "fastify-plugin";

const API_TOKEN = process.env.API_TOKEN || "";

const authPlugin = fp(async (app) => {
  // Skip auth if API_TOKEN is not set (dev mode)
  if (!API_TOKEN) return;

  app.addHook("onRequest", async (request, reply) => {
    // Skip health endpoint
    if (request.url === "/health") return;

    const auth = request.headers.authorization;
    if (!auth || auth !== `Bearer ${API_TOKEN}`) {
      reply.code(401);
      throw { message: "Unauthorized" };
    }
  });
});

export default authPlugin;