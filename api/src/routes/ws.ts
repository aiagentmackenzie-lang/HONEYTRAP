import { FastifyPluginCallback } from "fastify";

// WebSocket plugin using @fastify/websocket
// Requires `@fastify/websocket` and `ws` as peer dependencies
const wsPlugin: FastifyPluginCallback = async (app, _opts) => {
  // Track connected clients for broadcasting
  const clients = new Set<any>();

  app.get("/ws", { websocket: true }, (socket: any, req: any) => {
    clients.add(socket);
    app.log.info({ clients: clients.size }, "WebSocket client connected");

    socket.on("close", () => {
      clients.delete(socket);
      app.log.info({ clients: clients.size }, "WebSocket client disconnected");
    });

    socket.on("error", (err: any) => {
      app.log.error({ err }, "WebSocket error");
      clients.delete(socket);
    });
  });

  // Broadcast helper - attach to app.decorate
  const broadcast = (type: string, data: any) => {
    const message = JSON.stringify({ type, data, timestamp: new Date().toISOString() });
    for (const client of clients) {
      if (client.readyState === 1) { // WebSocket.OPEN = 1
        client.send(message);
      }
    }
  };

  // Decorate app so routes can broadcast
  if (!app.hasDecorator("wsBroadcast")) {
    app.decorate("wsBroadcast", broadcast);
  }
};

export default wsPlugin;

// Augment FastifyInstance
declare module "fastify" {
  interface FastifyInstance {
    wsBroadcast?: (type: string, data: any) => void;
  }
}