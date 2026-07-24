import { loadConfig } from "./config";
import { buildContainer } from "./composition";
import { buildServer } from "./interface/http/server";

// Entry point. Loads + validates config, wires the container (composition root),
// and serves the Hono app. Run with Bun, which serves the default export.
const config = loadConfig();
const container = await buildContainer(config);
const app = buildServer(container);

console.log(
  `[ai] LangGraph service listening on http://localhost:${config.port} ` +
    `(model=${config.model.name}, checkpointer=${config.checkpointer.kind})`,
);

export default { port: config.port, fetch: app.fetch };
