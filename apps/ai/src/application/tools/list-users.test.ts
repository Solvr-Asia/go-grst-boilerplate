import { describe, it, expect } from "bun:test";
import type { UserDirectory } from "../../domain/users";
import { makeListUsersTool } from "./list-users";

// The port makes this trivially testable: no module mocking, no live API — just
// a fake UserDirectory.
const fakeUsers: UserDirectory = {
  async listUsers({ page, size }) {
    return {
      users: [{ id: "1", email: "a@b.c", name: "A", status: "active" }],
      total: page * size, // echo inputs so we can assert they were passed through
    };
  },
};

describe("list_users tool", () => {
  it("maps directory results into the tool's JSON output", async () => {
    const tool = makeListUsersTool(fakeUsers);
    const out = (await tool.invoke({ page: 2, size: 5 })) as string;
    const parsed = JSON.parse(out) as { users: unknown[]; total: number };

    expect(parsed.users).toHaveLength(1);
    expect(parsed.total).toBe(10); // 2 * 5 — page/size reached the directory
  });
});
