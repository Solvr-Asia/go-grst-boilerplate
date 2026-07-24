import { createApiClient } from "@grst/api-client";

// Shared client the agent's tools and workflows use to reach the Go API.
// The AI service authenticates with an optional service token; it has no direct
// database access, so it inherits the Go server's auth policy and validation.
export const api = createApiClient({
  baseUrl: process.env.API_BASE_URL ?? "http://localhost:3000",
  getToken: () => process.env.API_SERVICE_TOKEN ?? null,
});
