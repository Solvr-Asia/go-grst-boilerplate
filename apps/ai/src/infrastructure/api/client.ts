import { createApiClient, type ApiClient } from "@grst/api-client";
import type { Config } from "../../config";

// The shared typed REST client the AI service uses to reach the Go API. The AI
// service authenticates with an optional service token; it has no direct
// database access, so it inherits the Go server's auth policy and validation.
export function createApi(config: Config): ApiClient {
  return createApiClient({
    baseUrl: config.api.baseUrl,
    getToken: () => config.api.serviceToken,
  });
}
