# Monorepo Restructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert `veemon` into a polyglot Bun-workspace monorepo — `apps/api` (Go, relocated unchanged), `apps/web` (React + TanStack Router + Vite + Tauri), `apps/ai` (Mastra.ai) — sharing one proto contract via generated Go + TypeScript clients.

**Architecture:** `contract/*.proto` is the single source of truth. `buf` generates Go into `apps/api` and TypeScript (protobuf-es) into `packages/api-client`. The TS apps call the existing Fiber REST routes through a thin typed client (Approach A). Bun workspaces orchestrate the TS apps; the Go app keeps its `make` targets.

**Tech Stack:** Go 1.25.12, buf + protoc-gen-go/-grpc/-fiber/-es, Bun 1.3.13, React 19, TanStack Router + Query, Vite, Tauri v2, Mastra (@mastra/core + zod v4), protobuf-es.

## Global Constraints

- Go module name stays **`veemon`** — never rename (imports depend on it).
- Workspace package scope is **`@veemon/*`**.
- Ports: api **:3000** HTTP / **:50051** gRPC · web **:5173** · ai **:4111**.
- `buf` installed via `go install github.com/bufbuild/buf/cmd/buf@latest`.
- TS codegen plugin `protoc-gen-es` resolved locally from `node_modules/.bin` (not global).
- `packages/api-client/src/gen/` is **committed**.
- Node 20+ / Bun for all TS; `apps/api` keeps `make`, is NOT a Bun workspace member.
- **Environment limits (verify what runs, flag what can't):** Rust/`cargo` absent → Tauri Rust build (`tauri build`) is NOT runnable here; verify the web app via `vite build` instead and scaffold `src-tauri/` files correctly. Running the Mastra agent needs a model API key (absent) → scaffold + typecheck + mocked tests only.
- Every moved file uses `git mv` (preserve history). Commit after each task.

---

## Task 1: Prerequisite tooling

**Files:** none (environment setup)

- [ ] **Step 1: Install buf**

Run: `go install github.com/bufbuild/buf/cmd/buf@latest`
Expected: `buf` on PATH (`buf --version` prints a version).

- [ ] **Step 2: Verify current proto regen still works at root (baseline)**

Run: `make proto && go build ./... && go test -race ./... 2>&1 | tail -5`
Expected: proto regenerates, build succeeds, tests pass. This is the pre-move baseline — if it's not green now, stop and fix before moving anything.

---

## Task 2: Relocate Go module → `apps/api`

**Files:**
- Move: `app/ cmd/ config/ database/ entity/ handler/ pkg/ repository/ clients/ examples/ migrations/ docs/scalar.go go.mod go.sum Dockerfile Makefile .env.example` → under `apps/api/`
- Keep at root: `contract/ buf.yaml buf.gen.yaml docker-compose.yml docs/superpowers/ README.md CHANGELOG.md LICENSE claude.md .gitignore .github/`
- Modify: `buf.gen.yaml`, `docker-compose.yml`, `apps/api/Makefile`, `.gitignore`

**Interfaces:**
- Produces: Go module rooted at `apps/api/` with unchanged import paths (`veemon/...`); `apps/api/bin/protoc-gen-fiber` built by `make -C apps/api proto`.

- [ ] **Step 1: Create apps/api and git mv the Go tree**

```bash
mkdir -p apps/api
git mv app cmd config database entity handler pkg repository clients examples migrations go.mod go.sum Dockerfile Makefile apps/api/
mkdir -p apps/api/docs && git mv docs/scalar.go apps/api/docs/scalar.go
# .env.example if tracked:
git mv .env.example apps/api/.env.example 2>/dev/null || true
```

- [ ] **Step 2: Update `buf.gen.yaml` output paths** (root file)

```yaml
version: v2
plugins:
  - local: protoc-gen-go
    out: apps/api/handler/grpc
    opt: paths=source_relative
  - local: protoc-gen-go-grpc
    out: apps/api/handler/grpc
    opt: paths=source_relative
  - local: apps/api/bin/protoc-gen-fiber
    out: apps/api/handler/grpc
    opt: paths=source_relative
```

- [ ] **Step 3: Update `apps/api/Makefile` `proto` target** to build the plugin locally and run buf from repo root.

```makefile
proto:
	@echo "Building protoc-gen-fiber plugin..."
	@mkdir -p bin
	$(GOBUILD) -o ./bin/protoc-gen-fiber ./cmd/protoc-gen-fiber
	@echo "Generating proto files (buf runs from repo root)..."
	cd ../.. && buf generate
	@echo "Proto generation completed"
```

- [ ] **Step 4: Update `docker-compose.yml`** — change `app` and `migrate` service build context.

For both services: `context: .` → `context: ./apps/api` (dockerfile stays `Dockerfile`).

- [ ] **Step 5: Update root `.gitignore`** — append:

```
# JS/TS
**/node_modules/
**/dist/
# Tauri
apps/web/src-tauri/target/
# Go build output
apps/api/bin/
```

- [ ] **Step 6: GATE — Go app is green in its new home**

```bash
make -C apps/api proto
cd apps/api && go build ./... && go test -race ./... 2>&1 | tail -5
```
Expected: proto regenerates into `apps/api/handler/grpc`, build + tests pass identically to Task 1 baseline.

- [ ] **Step 7: Commit**

```bash
git add -A && git commit -m "refactor: relocate Go module to apps/api"
```

---

## Task 3: Bun workspace root + shared tsconfig

**Files:**
- Create: `package.json` (root), `packages/tsconfig/package.json`, `packages/tsconfig/base.json`

**Interfaces:**
- Produces: root Bun workspace (`apps/*`, `packages/*`); `@veemon/tsconfig/base.json` consumed by all TS packages.

- [ ] **Step 1: Root `package.json`**

```json
{
  "name": "veemon",
  "private": true,
  "workspaces": ["apps/web", "apps/ai", "packages/*"],
  "scripts": {
    "proto": "make -C apps/api proto",
    "build": "bun run build:client && bun run build:web && bun run build:ai && bun run build:api",
    "build:api": "make -C apps/api build",
    "build:client": "bun --filter @veemon/api-client run build",
    "build:web": "bun --filter @veemon/web run build",
    "build:ai": "bun --filter @veemon/ai run build",
    "test": "bun run test:api && bun run test:ts",
    "test:api": "cd apps/api && go test -race ./...",
    "test:ts": "bun --filter './packages/*' --filter '@veemon/web' --filter '@veemon/ai' run test",
    "dev:api": "make -C apps/api dev",
    "dev:web": "bun --filter @veemon/web run dev",
    "dev:ai": "bun --filter @veemon/ai run dev",
    "dev": "concurrently -n api,web,ai -c blue,green,magenta 'bun run dev:api' 'bun run dev:web' 'bun run dev:ai'"
  },
  "devDependencies": {
    "concurrently": "^9.1.0",
    "typescript": "^5.7.0"
  }
}
```

- [ ] **Step 2: `packages/tsconfig/package.json`**

```json
{ "name": "@veemon/tsconfig", "version": "0.0.0", "private": true, "files": ["base.json"] }
```

- [ ] **Step 3: `packages/tsconfig/base.json`**

```json
{
  "$schema": "https://json.schemastore.org/tsconfig",
  "compilerOptions": {
    "target": "ES2022",
    "module": "ES2022",
    "moduleResolution": "bundler",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "verbatimModuleSyntax": true,
    "noUncheckedIndexedAccess": true
  }
}
```

- [ ] **Step 4: Install & commit**

```bash
bun install
git add -A && git commit -m "chore: add Bun workspace root and shared tsconfig"
```
Expected: `bun.lock` created, `node_modules` populated, `concurrently`/`typescript` resolved.

---

## Task 4: `packages/api-client` — generated types + typed REST client

**Files:**
- Create: `packages/api-client/package.json`, `tsconfig.json`, `tsconfig.build.json`, `src/index.ts`, `src/client.ts`, `src/client.test.ts`
- Create (generated): `packages/api-client/src/gen/**`
- Modify: `buf.gen.yaml` (add TS output)

**Interfaces:**
- Produces: `@veemon/api-client` exporting `createApiClient(opts: ApiClientOptions): ApiClient`, `ApiError`, and generated message types. `ApiClientOptions = { baseUrl: string; getToken?: () => string | null | Promise<string|null>; fetch?: typeof fetch }`. `ApiClient` methods: `login`, `register`, `refreshToken`, `getMe`, `logout`, `listUsers`.

- [ ] **Step 1: `packages/api-client/package.json`**

```json
{
  "name": "@veemon/api-client",
  "version": "0.0.0",
  "type": "module",
  "main": "./dist/index.js",
  "types": "./dist/index.d.ts",
  "exports": { ".": { "types": "./dist/index.d.ts", "default": "./dist/index.js" } },
  "scripts": {
    "build": "tsc -p tsconfig.build.json",
    "test": "bun test",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": { "@bufbuild/protobuf": "^2.2.3" },
  "devDependencies": { "@bufbuild/protoc-gen-es": "^2.2.3", "@veemon/tsconfig": "*", "typescript": "^5.7.0" }
}
```

- [ ] **Step 2: `tsconfig.json` + `tsconfig.build.json`**

`tsconfig.json`:
```json
{ "extends": "@veemon/tsconfig/base.json", "compilerOptions": { "noEmit": true }, "include": ["src"] }
```
`tsconfig.build.json`:
```json
{ "extends": "@veemon/tsconfig/base.json", "compilerOptions": { "outDir": "dist", "declaration": true, "noEmit": false }, "include": ["src"], "exclude": ["src/**/*.test.ts"] }
```

- [ ] **Step 3: Add TS output to root `buf.gen.yaml`** (append plugin)

```yaml
  - local: ./node_modules/.bin/protoc-gen-es
    out: packages/api-client/src/gen
    opt:
      - target=ts
      - import_extension=js
```

- [ ] **Step 4: Install & generate**

```bash
bun install
make -C apps/api proto   # now emits Go (apps/api) AND TS (packages/api-client/src/gen)
ls packages/api-client/src/gen
```
Expected: `user_pb.ts` (and `veemon/annotations_pb.ts`) generated.

- [ ] **Step 5: Write `src/client.ts`** — typed REST client over the Fiber routes.

Reads the `veemon.route` paths from the proto (documented in `contract/user/user.proto`). Decodes the `{ success, data, meta }` / `{ success, error }` envelope. (Exact request/response field names come from generated `src/gen/user_pb.ts`; map JSON via camelCase per `pkg/response`.)

```ts
import type { Message } from "@bufbuild/protobuf";

export class ApiError extends Error {
  constructor(public code: number, message: string) { super(message); this.name = "ApiError"; }
}
export interface ApiClientOptions {
  baseUrl: string;
  getToken?: () => string | null | Promise<string | null>;
  fetch?: typeof fetch;
}
interface Envelope<T> { success: boolean; data?: T; meta?: unknown; error?: { code: number; message: string }; }

export function createApiClient(opts: ApiClientOptions) {
  const doFetch = opts.fetch ?? fetch;
  async function request<T>(method: string, path: string, body?: unknown, auth = true): Promise<T> {
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    if (auth && opts.getToken) { const t = await opts.getToken(); if (t) headers["Authorization"] = `Bearer ${t}`; }
    const res = await doFetch(`${opts.baseUrl}${path}`, { method, headers, body: body ? JSON.stringify(body) : undefined });
    const json = (await res.json()) as Envelope<T>;
    if (!json.success) throw new ApiError(json.error?.code ?? res.status, json.error?.message ?? res.statusText);
    return json.data as T;
  }
  return {
    register: (b: { email: string; password: string; name?: string }) => request("POST", "/api/v1/users/register", b, false),
    login: (b: { email: string; password: string }) => request<{ accessToken: string; refreshToken?: string }>("POST", "/api/v1/users/login", b, false),
    refreshToken: (b: { refreshToken: string }) => request<{ accessToken: string }>("POST", "/api/v1/users/refresh", b, false),
    getMe: () => request<{ id: string; email: string; name?: string; roles?: string[] }>("GET", "/api/v1/users/me"),
    logout: () => request<{ success: boolean }>("POST", "/api/v1/users/logout"),
    listUsers: (q?: { page?: number; size?: number }) => request<{ users: unknown[]; meta?: unknown }>("GET", `/api/v1/users${q ? `?page=${q.page ?? 1}&size=${q.size ?? 10}` : ""}`),
  };
}
export type ApiClient = ReturnType<typeof createApiClient>;
```

> NOTE: verify the exact route paths against `apps/api/handler/grpc/user/user_fiber.pb.go` (the `veemon.route` `path` values) during execution and correct the strings above to match.

- [ ] **Step 6: `src/index.ts`**

```ts
export * from "./client.js";
export * from "./gen/user/user_pb.js";
```

- [ ] **Step 7: Write failing test `src/client.test.ts`**

```ts
import { describe, it, expect } from "bun:test";
import { createApiClient, ApiError } from "./client.js";

function mockFetch(status: number, payload: unknown): typeof fetch {
  return (async () => new Response(JSON.stringify(payload), { status, headers: { "Content-Type": "application/json" } })) as unknown as typeof fetch;
}

describe("api-client", () => {
  it("returns data on success envelope", async () => {
    const c = createApiClient({ baseUrl: "http://x", fetch: mockFetch(200, { success: true, data: { id: "1", email: "a@b.c" } }) });
    expect(await c.getMe()).toEqual({ id: "1", email: "a@b.c" });
  });
  it("throws ApiError on failure envelope", async () => {
    const c = createApiClient({ baseUrl: "http://x", fetch: mockFetch(400, { success: false, error: { code: 40001, message: "bad" } }) });
    await expect(c.getMe()).rejects.toBeInstanceOf(ApiError);
  });
  it("attaches bearer token when getToken provided", async () => {
    let seen = "";
    const f = (async (_u: string, init: RequestInit) => { seen = (init.headers as Record<string,string>)["Authorization"] ?? ""; return new Response(JSON.stringify({ success: true, data: {} }), { status: 200 }); }) as unknown as typeof fetch;
    const c = createApiClient({ baseUrl: "http://x", fetch: f, getToken: () => "tok" });
    await c.getMe();
    expect(seen).toBe("Bearer tok");
  });
});
```

- [ ] **Step 8: GATE — run tests + build + typecheck**

```bash
cd packages/api-client && bun test && bun run typecheck && bun run build && ls dist
```
Expected: 3 tests pass, `dist/index.js` + `.d.ts` emitted.

- [ ] **Step 9: Commit**

```bash
git add -A && git commit -m "feat: add @veemon/api-client (generated types + typed REST client)"
```

---

## Task 5: Scaffold `apps/web` (React + TanStack Router + Vite + Tauri)

**Files:** Create under `apps/web/`: `package.json`, `tsconfig.json`, `vite.config.ts`, `index.html`, `src/main.tsx`, `src/routes/{__root,index,login,me}.tsx`, `src/lib/{api.ts,auth.ts}`, `src/styles.css`, `.env.example`, and `src-tauri/{Cargo.toml,tauri.conf.json,build.rs,src/main.rs,src/lib.rs}`. Plus a route test `src/lib/api.test.ts`.

**Interfaces:**
- Consumes: `@veemon/api-client` (`createApiClient`, `ApiError`).
- Produces: browser-buildable Vite app; Tauri shell files (Rust compile deferred).

- [ ] **Step 1: `apps/web/package.json`**

```json
{
  "name": "@veemon/web",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc --noEmit && vite build",
    "preview": "vite preview",
    "test": "bun test",
    "tauri:dev": "tauri dev",
    "tauri:build": "tauri build"
  },
  "dependencies": {
    "@veemon/api-client": "*",
    "@tanstack/react-query": "^5.62.0",
    "@tanstack/react-router": "^1.87.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  },
  "devDependencies": {
    "@veemon/tsconfig": "*",
    "@tanstack/router-plugin": "^1.87.0",
    "@tauri-apps/cli": "^2.1.0",
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "@vitejs/plugin-react": "^4.3.4",
    "typescript": "^5.7.0",
    "vite": "^6.0.0"
  }
}
```

- [ ] **Step 2: `tsconfig.json`, `vite.config.ts`, `index.html`, `src/styles.css`**

`tsconfig.json`:
```json
{ "extends": "@veemon/tsconfig/base.json", "compilerOptions": { "jsx": "react-jsx", "noEmit": true, "types": ["vite/client"] }, "include": ["src"] }
```
`vite.config.ts`:
```ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { TanStackRouterVite } from "@tanstack/router-plugin/vite";
export default defineConfig({
  plugins: [TanStackRouterVite({ target: "react", autoCodeSplitting: true }), react()],
  clearScreen: false,
  server: { port: 5173, strictPort: true },
});
```
`index.html`:
```html
<!doctype html><html lang="en"><head><meta charset="UTF-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/><title>@veemon/web</title></head><body><div id="root"></div><script type="module" src="/src/main.tsx"></script></body></html>
```
`src/styles.css`: minimal reset (body font, container spacing).

- [ ] **Step 3: `src/lib/auth.ts` and `src/lib/api.ts`**

`auth.ts`:
```ts
const KEY = "veemon.token";
export const getToken = () => localStorage.getItem(KEY);
export const setToken = (t: string) => localStorage.setItem(KEY, t);
export const clearToken = () => localStorage.removeItem(KEY);
```
`api.ts`:
```ts
import { createApiClient } from "@veemon/api-client";
import { getToken } from "./auth";
export const api = createApiClient({
  baseUrl: import.meta.env.VITE_API_BASE_URL ?? "http://localhost:3000",
  getToken,
});
```

- [ ] **Step 4: Routes** — `__root.tsx` (layout + nav + QueryClientProvider outlet), `index.tsx` (home), `login.tsx` (form → `api.login` → `setToken` → navigate to `/me`), `me.tsx` (uses `useQuery` for `api.getMe` + `api.listUsers`, redirect to `/login` on `ApiError`). `src/main.tsx` wires `createRouter(routeTree)` + `QueryClientProvider` + `RouterProvider`.

(Full component code written at execution — each route is a small typed component consuming `api`.)

- [ ] **Step 5: `.env.example`**

```
VITE_API_BASE_URL=http://localhost:3000
```

- [ ] **Step 6: `src-tauri/` shell** — `Cargo.toml` (tauri v2 deps), `tauri.conf.json` (`build.devUrl=http://localhost:5173`, `build.frontendDist=../dist`, `beforeDevCommand=bun run dev`, `beforeBuildCommand=bun run build`), `build.rs`, `src/main.rs` + `src/lib.rs` (default `tauri::Builder` run). These are correct-by-construction; Rust compile is NOT run here (no cargo).

- [ ] **Step 7: GATE — install, typecheck, build (browser), test**

```bash
bun install
cd apps/web && bun run build 2>&1 | tail -15   # tsc --noEmit && vite build → dist/
bun test 2>&1 | tail -5
```
Expected: TanStack route tree generates, `tsc` clean, `vite build` emits `dist/`, tests pass. (Tauri build intentionally skipped — needs Rust.)

- [ ] **Step 8: Commit**

```bash
git add -A && git commit -m "feat: scaffold apps/web (React + TanStack Router + Vite + Tauri shell)"
```

---

## Task 6: Scaffold `apps/ai` (Mastra.ai)

**Files:** Create under `apps/ai/`: `package.json`, `tsconfig.json`, `.env.example`, `src/mastra/index.ts`, `src/mastra/tools/user-tools.ts`, `src/mastra/agents/assistant-agent.ts`, `src/mastra/workflows/example-workflow.ts`, `src/mastra/tools/user-tools.test.ts`.

**Interfaces:**
- Consumes: `@veemon/api-client`.
- Produces: `@veemon/ai` Mastra project; `userTools` (createTool) calling the API; `assistantAgent`; `mastra` instance.

> Verify current Mastra APIs from `node_modules/@mastra/core/dist/docs` after install; correct constructor/tool signatures if they differ from below.

- [ ] **Step 1: `package.json`**

```json
{
  "name": "@veemon/ai",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "scripts": { "dev": "mastra dev", "build": "mastra build", "test": "bun test", "typecheck": "tsc --noEmit" },
  "dependencies": { "@veemon/api-client": "*", "@mastra/core": "latest", "zod": "^4.0.0" },
  "devDependencies": { "@veemon/tsconfig": "*", "mastra": "latest", "typescript": "^5.7.0" }
}
```

- [ ] **Step 2: `tsconfig.json`**

```json
{ "extends": "@veemon/tsconfig/base.json", "compilerOptions": { "noEmit": true, "outDir": "dist" }, "include": ["src"] }
```

- [ ] **Step 3: `.env.example`**

```
ANTHROPIC_API_KEY=
API_BASE_URL=http://localhost:3000
API_SERVICE_TOKEN=
```

- [ ] **Step 4: `src/mastra/tools/user-tools.ts`** — a `createTool` wrapping `@veemon/api-client`.

```ts
import { createTool } from "@mastra/core/tools";
import { z } from "zod";
import { createApiClient } from "@veemon/api-client";

const api = createApiClient({
  baseUrl: process.env.API_BASE_URL ?? "http://localhost:3000",
  getToken: () => process.env.API_SERVICE_TOKEN ?? null,
});

export const listUsersTool = createTool({
  id: "list-users",
  description: "List registered users from the backend API",
  inputSchema: z.object({ page: z.number().default(1), size: z.number().default(10) }),
  outputSchema: z.object({ users: z.array(z.unknown()) }),
  execute: async ({ context }) => {
    const data = await api.listUsers({ page: context.page, size: context.size });
    return { users: data.users ?? [] };
  },
});
```

- [ ] **Step 5: `agents/assistant-agent.ts` + `workflows/example-workflow.ts` + `mastra/index.ts`**

Agent: `new Agent({ id, name, instructions, model: process.env model or "anthropic/claude-sonnet-4-5", tools: { listUsersTool } })`. Workflow: one sample step. `index.ts`: `export const mastra = new Mastra({ agents: { assistantAgent }, workflows: { exampleWorkflow } })`. (Model id verified via Mastra provider registry at execution.)

- [ ] **Step 6: Failing test `tools/user-tools.test.ts`** — mock `@veemon/api-client` and assert `listUsersTool.execute` returns mapped users.

```ts
import { describe, it, expect, mock } from "bun:test";
mock.module("@veemon/api-client", () => ({ createApiClient: () => ({ listUsers: async () => ({ users: [{ id: "1" }] }) }) }));
const { listUsersTool } = await import("./user-tools.js");
describe("listUsersTool", () => {
  it("returns users from the api", async () => {
    const out = await (listUsersTool as any).execute({ context: { page: 1, size: 10 } });
    expect(out.users).toHaveLength(1);
  });
});
```

- [ ] **Step 7: GATE — install, typecheck, test**

```bash
bun install
cd apps/ai && bun run typecheck 2>&1 | tail -10 && bun test 2>&1 | tail -5
```
Expected: typecheck clean, tool test passes. (`mastra build` and live agent need a model key — not run here.)

- [ ] **Step 8: Commit**

```bash
git add -A && git commit -m "feat: scaffold apps/ai (Mastra agent + api-client tool)"
```

---

## Task 7: CI, release, and docs

**Files:** Modify `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `README.md`, `claude.md`; create `apps/web/README.md`, `apps/ai/README.md` (short).

- [ ] **Step 1: `ci.yml`** — set the Go job `defaults.run.working-directory: apps/api` (or prefix `working-directory` per step) and add `paths` filters. Add a `web-ai` job: `oven-sh/setup-bun`, `bun install`, `bun --filter '@veemon/api-client' --filter '@veemon/web' --filter '@veemon/ai' run test`, and `bun run build:client`.

- [ ] **Step 2: `release.yml`** — prefix Go build paths with `apps/api` (`working-directory: apps/api`, outputs to repo `bin/`).

- [ ] **Step 3: Update `claude.md`** — Architecture Layers gains the monorepo `apps/` + `packages/` layout; note `contract/` at root drives Go + TS; add web/ai app notes and the Bun scripts.

- [ ] **Step 4: Update root `README.md`** — monorepo quickstart (`bun install`, `bun run proto`, `bun run dev`), per-app sections, prerequisites (Rust for Tauri, model key for ai).

- [ ] **Step 5: GATE — full-repo verification**

```bash
make -C apps/api proto && (cd apps/api && go build ./... && go test -race ./...) \
  && bun install && bun run build:client \
  && bun --filter '@veemon/api-client' --filter '@veemon/ai' run test \
  && (cd apps/web && bun run build)
```
Expected: Go green; api-client builds + tests; ai tests; web browser build emits `dist/`.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "ci+docs: update workflows and docs for monorepo layout"
```

---

## Final verification

- [ ] `git status` clean; branch `feat/monorepo-restructure` contains all commits.
- [ ] Re-run Task 7 Step 5 gate end-to-end — all green.
- [ ] Summarize what runs here vs. what needs Rust (Tauri build) / a model key (live ai agent).
