import { createApiClient } from "@veemon/api-client";
import { getToken } from "./auth";

// Single configured client instance for the whole app. Base URL comes from the
// Vite env (VITE_API_BASE_URL), defaulting to the local Go API port.
export const api = createApiClient({
  baseUrl: import.meta.env.VITE_API_BASE_URL ?? "http://localhost:3000",
  getToken,
});
