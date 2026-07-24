import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { TanStackRouterVite } from "@tanstack/router-plugin/vite";

// https://vite.dev + Tauri: fixed dev port, no auto-clear so Tauri logs stay visible.
export default defineConfig({
  plugins: [TanStackRouterVite({ target: "react" }), react()],
  clearScreen: false,
  server: {
    port: 5173,
    strictPort: true,
  },
});
