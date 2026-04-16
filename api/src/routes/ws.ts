import { FastifyPluginCallback } from "fastify";
import { WebSocket } from "ws";

const wsPlugin: FastifyPluginCallback = async (app, _opts) => {
  const clients = new Set<WebSocket>();

  app.get("/ws", { websocket: true }, (socket: WebSocket) => {
    clients.add(socket);
    app.log.info({ clients: clients.size }, "WebSocket client connected");

    socket.on("close", () => {
      clients.delete(socket);
      app.log.info({ clients: clients.size }, "WebSocket client disconnected");
    });

    socket.on("error", (err) => {
      app.log.error({ err }, "WebSocket error");
      clients.delete(socket);
    });
  });

  // Broadcast helper - attach to app.decorate
  const broadcast = (type: string, data: any) => {
    const message = JSON.stringify({ type, data, timestamp: new Date().toISOString() });
    for (const client of clients) {
      if (client.readyState === WebSocket.OPEN) {
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