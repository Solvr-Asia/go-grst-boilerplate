import { createTool } from "@mastra/core/tools";
import { z } from "zod";
import { api } from "../api";

// A Mastra tool that lets the agent read registered users from the Go API.
// In @mastra/core v1.x, execute's FIRST argument is the validated inputSchema
// value (destructured directly); the second is the execution context.
export const listUsersTool = createTool({
  id: "list-users",
  description:
    "List registered users from the backend API. Use when asked about how many users exist or who is registered. Requires admin scope on the API service token.",
  inputSchema: z.object({
    page: z.number().int().min(1).default(1),
    size: z.number().int().min(1).max(100).default(10),
  }),
  outputSchema: z.object({
    users: z.array(
      z.object({
        id: z.string(),
        email: z.string(),
        name: z.string(),
        status: z.string(),
      }),
    ),
    total: z.number().optional(),
  }),
  execute: async ({ page, size }) => {
    const res = await api.listUsers({ page, size });
    return {
      users: res.users.map((u) => ({
        id: u.id,
        email: u.email,
        name: u.name,
        status: u.status,
      })),
      total: res.pagination?.total,
    };
  },
});
