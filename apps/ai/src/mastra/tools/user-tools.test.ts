import { describe, it, expect, mock } from "bun:test";

// Mock the shared api-client so the tool's logic is tested without a live API.
mock.module("@grst/api-client", () => ({
  createApiClient: () => ({
    listUsers: async () => ({
      users: [
        {
          id: "1",
          email: "a@b.c",
          name: "A",
          phone: "",
          status: "active",
          createdAt: "",
        },
      ],
      pagination: { page: 1, size: 10, total: 1, totalPages: 1 },
    }),
  }),
}));

const { listUsersTool } = await import("./user-tools");

type Execute = (
  input: { page: number; size: number },
  ctx?: unknown,
) => Promise<{ users: unknown[]; total?: number }>;

describe("listUsersTool", () => {
  it("maps API users into the tool output shape", async () => {
    const execute = listUsersTool.execute as unknown as Execute;
    const out = await execute({ page: 1, size: 10 }, {});
    expect(out.users).toHaveLength(1);
    expect(out.total).toBe(1);
  });
});
