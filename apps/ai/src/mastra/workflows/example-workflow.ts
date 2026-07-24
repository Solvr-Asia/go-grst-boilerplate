import { createWorkflow, createStep } from "@mastra/core/workflows";
import { z } from "zod";
import { api } from "../api";

// A minimal one-step workflow demonstrating the primitive: fetch users from the
// Go API and report a count. Workflow steps receive { inputData } (unlike tools,
// which destructure the input value directly).
const countUsersStep = createStep({
  id: "count-users",
  description: "Fetch a page of users from the API and count them",
  inputSchema: z.object({
    page: z.number().int().min(1).default(1),
    size: z.number().int().min(1).max(100).default(10),
  }),
  outputSchema: z.object({
    page: z.number(),
    count: z.number(),
    total: z.number().optional(),
  }),
  execute: async ({ inputData }) => {
    const res = await api.listUsers({
      page: inputData.page,
      size: inputData.size,
    });
    return {
      page: inputData.page,
      count: res.users.length,
      total: res.pagination?.total,
    };
  },
});

export const userReportWorkflow = createWorkflow({
  id: "user-report",
  inputSchema: z.object({
    page: z.number().int().min(1).default(1),
    size: z.number().int().min(1).max(100).default(10),
  }),
  outputSchema: z.object({
    page: z.number(),
    count: z.number(),
    total: z.number().optional(),
  }),
})
  .then(countUsersStep)
  .commit();
