import { describe, it, expect, beforeEach } from "bun:test";

// Minimal localStorage stub so the browser-oriented module runs under bun.
const store = new Map<string, string>();
(globalThis as unknown as { localStorage: Storage }).localStorage = {
  getItem: (k: string) => store.get(k) ?? null,
  setItem: (k: string, v: string) => void store.set(k, v),
  removeItem: (k: string) => void store.delete(k),
  clear: () => store.clear(),
  key: () => null,
  length: 0,
} as Storage;

const { getToken, setToken, clearToken } = await import("./auth");

describe("auth token storage", () => {
  beforeEach(() => store.clear());

  it("returns null when no token is set", () => {
    expect(getToken()).toBeNull();
  });

  it("stores and reads a token", () => {
    setToken("abc");
    expect(getToken()).toBe("abc");
  });

  it("clears a token", () => {
    setToken("abc");
    clearToken();
    expect(getToken()).toBeNull();
  });
});
