# Monorepo Restructure — Design

**Date:** 2026-07-24
**Status:** Approved (design)
**Topic:** Convert `go-grst-boilerplate` into a polyglot starter-kit monorepo

---

## 1. Goal & Scope

Turn the current single-module Go service into a **reusable starter-kit monorepo**
hosting three apps that share one proto contract:

- **`apps/api`** — the existing Go backend (Fiber REST + gRPC), relocated unchanged.
- **`apps/web`** — React + TanStack Router + Vite, wrapped in a **Tauri** desktop shell.
- **`apps/ai`** — a **Mastra.ai** (TypeScript) service whose agent calls the Go API.

**Chosen approach: A — REST-over-shared-types.** The `contract/` proto stays the single
source of truth; buf generates Go (into `apps/api`) *and* TypeScript (into
`packages/api-client`). The TS apps call the **existing Fiber REST routes** through a
thin typed client — **no changes to the Go server's behavior**. (ConnectRPC end-to-end
was considered as "Approach C" and deferred as a future feature, not part of
monorepo-ification.)

**This pass = skeleton + full scaffolds:** relocate the Go app, stand up the workspace
+ shared contract pipeline, and fully scaffold `web` and `ai` as working starters (real
stacks wired to the API — not full products). Each app is then built out in its own
design→plan→build cycle.

### Non-goals (this pass)
- No behavioral changes to the Go server (relocate only).
- No ConnectRPC / gRPC-web layer (REST is reused).
- No auto-generated TS route client — deferred (see §7).
- No full feature build-out of `web` or `ai` beyond the demo flows below.

---

## 2. Target Repository Layout

```
go-grst-boilerplate/                 ← repo root becomes the monorepo (Bun workspace)
├── apps/
│   ├── api/                         ← ALL current Go code, module name unchanged
│   │   ├── app/ cmd/ config/ database/ entity/ handler/
│   │   ├── pkg/ repository/ clients/ examples/ migrations/ docs/
│   │   ├── go.mod  go.sum  Dockerfile  Makefile  .env.example
│   ├── web/                         ← @grst/web — React + TanStack Router + Vite + Tauri
│   │   ├── src/{routes,components,lib}/  main.tsx  index.html
│   │   ├── src-tauri/               ← Rust desktop shell
│   │   ├── vite.config.ts  package.json  tsconfig.json
│   └── ai/                          ← @grst/ai — Mastra.ai service
│       ├── src/mastra/{agents,workflows,tools}/  index.ts
│       └── package.json  tsconfig.json  .env.example
├── packages/
│   ├── api-client/                  ← @grst/api-client — buf-generated TS types + typed REST client
│   │   ├── src/gen/                 ← protobuf-es output (committed)
│   │   ├── src/client.ts  src/index.ts  package.json  tsconfig.json
│   └── tsconfig/                    ← @grst/tsconfig — shared base tsconfig
├── contract/                        ← proto source of truth (stays at root)
│   └── user/user.proto  grst/annotations.proto
├── docs/superpowers/specs/          ← repo-level design docs (this file) — STAYS at root
├── buf.yaml  buf.gen.yaml           ← output paths updated (Go→apps/api, +TS→packages/api-client)
├── package.json  bun.lock           ← Bun workspaces + root orchestration scripts
├── docker-compose.yml               ← infra unchanged; api/migrate build context → ./apps/api
├── .gitignore                       ← + node_modules, dist, src-tauri/target, apps/api/bin
└── README.md  CHANGELOG.md  LICENSE  claude.md
```

**Key decisions (approved):**
- `contract/` stays at repo root (language-neutral; buf already points there).
- `packages/api-client/src/gen/` is **committed** (fresh clones run without a proto toolchain).
- Workspace package scope is **`@grst/*`**.

---

## 3. `apps/api` Relocation (keep it green)

**Rule: the Go app moves, it does not change.** Because `go.mod` (module
`go-grst-boilerplate`) travels with the code, every `go-grst-boilerplate/...` import
stays byte-for-byte identical — the compiler is blind to the directory move.

**Moves into `apps/api/`:** `app/ cmd/ config/ database/ entity/ handler/ pkg/
repository/ clients/ examples/ migrations/` and the Go docs package (`docs/scalar.go` →
`apps/api/docs/scalar.go`, still imported as `go-grst-boilerplate/docs`), plus
`go.mod go.sum Dockerfile Makefile .env.example` and generated `*.pb.go`.

**Stays at repo root:** `contract/`, `buf.yaml`, `docker-compose.yml`, `docs/superpowers/`
(repo-level — must NOT be swept into `apps/api`), `README/LICENSE/CHANGELOG/claude.md`,
`.gitignore`.

**Wiring edits forced by the move:**

| File | Change | Why |
|------|--------|-----|
| `buf.gen.yaml` | Go `out: handler/grpc` → `out: apps/api/handler/grpc`; plugin `./bin/protoc-gen-fiber` → `apps/api/bin/protoc-gen-fiber` | buf runs from root; Go lands in the moved module. `go_package` is already `go-grst-boilerplate/handler/grpc/...`, so imports are unchanged. |
| `docker-compose.yml` | `app` + `migrate` build `context: .` → `context: ./apps/api` | Dockerfile now lives in `apps/api`. Infra services (postgres/redis/rabbitmq/jaeger) untouched. |
| `apps/api/Makefile` | `proto` target builds `protoc-gen-fiber` locally, then `cd ../.. && buf generate` | proto source is one level up |
| root `package.json` | `api:*` scripts shell to `make -C apps/api …` | Bun orchestrates Go without owning it |
| `.gitignore` | add `apps/web/src-tauri/target/`, `**/node_modules/`, `**/dist/`, `apps/api/bin/` | polyglot artifacts |

**Verification gate:** after the move,
`cd apps/api && make proto && go build ./... && go test -race ./...` passes exactly as
today. Files moved with `git mv` to preserve history.

---

## 4. Shared Contract Pipeline (proto → Go + TS)

```
                    contract/*.proto  (single source of truth)
                            │ buf generate
             ┌───────────────┴────────────────┐
     protoc-gen-go/-grpc               protoc-gen-es (protobuf-es, target=ts)
     + protoc-gen-fiber                        │
             ▼                                  ▼
  apps/api/handler/grpc/*.pb.go     packages/api-client/src/gen/  (TS message types)
  (Go server, unchanged)                        │
                                     packages/api-client/src/client.ts
                                     (thin typed REST client, hand-written)
                                                │
                                    ┌───────────┴───────────┐
                                    ▼                       ▼
                               apps/web                  apps/ai
```

**`buf.gen.yaml` gains a TS output** (`protoc-gen-es`, `target=ts`) → `packages/api-client/src/gen/`.
No Connect plugin (REST reused).

**`packages/api-client` (`@grst/api-client`):**
- `src/gen/` — generated protobuf-es message/enum types (committed).
- `src/client.ts` — `createApiClient({ baseUrl, getToken })` returning typed methods
  (`register`, `login`, `refreshToken`, `getMe`, `logout`, `listUsers`) that call the
  existing Fiber REST routes with generated request/response types, attach the PASETO
  bearer via `getToken()`, decode the standard camelCase `{ success, data, meta }`
  envelope (matches `pkg/response`), and throw a typed `ApiError(code, message)` on
  `{ success:false, error:{...} }`.
- `src/index.ts` — barrel export.
- deps: `@bufbuild/protobuf`.

**Why hand-write `client.ts`:** the `grst.route` annotations that drive Go route
generation aren't understood by TS codegen, so route knowledge is bridged by hand once,
here. A `protoc-gen-ts-client` mirroring `protoc-gen-fiber` is the symmetry play — see §7.

---

## 5. `apps/web` (React + TanStack Router + Vite + Tauri)

```
apps/web/                       package: @grst/web
├── src/
│   ├── routes/                 ← TanStack Router file-based routes
│   │   ├── __root.tsx  index.tsx  login.tsx  me.tsx
│   ├── routeTree.gen.ts        ← generated by @tanstack/router-plugin
│   ├── lib/{api.ts, auth.ts}   ← configured @grst/api-client + token storage
│   ├── components/  main.tsx (RouterProvider + QueryClientProvider)  styles.css
├── src-tauri/                  ← Rust shell: Cargo.toml, tauri.conf.json, src/main.rs
├── index.html  vite.config.ts (react + tanstackRouter plugins)
├── tsconfig.json (extends @grst/tsconfig)  package.json
```

- **Data layer:** TanStack Query over `@grst/api-client` (queries for `getMe`/`listUsers`,
  mutation for `login`).
- **Demo flow shipped:** unauthenticated → login page (POST Go `/login`, store PASETO
  token) → authenticated page showing `getMe` + a `listUsers` table. Exercises auth,
  token header, response envelope, and typed errors end-to-end.
- **Tauri:** `bun run dev` = plain Vite in a browser (fast day-to-day); `bun run tauri:dev`
  = desktop window. **Requires a Rust toolchain + OS webview deps** (browser dev does not).
- **CORS (dev):** add the Vite origin (`http://localhost:5173`) to the Go server's
  `CORS_ORIGINS` in `apps/api/.env.example` (never `*` — `Config.Validate()` rejects that
  in prod).

---

## 6. `apps/ai` (Mastra.ai service)

> Designed against the Mastra skill's current reference (packages not yet installed).
> Verify exact APIs/model ids from embedded docs at implementation time.

```
apps/ai/                        package: @grst/ai
├── src/mastra/
│   ├── tools/user-tools.ts     ← createTool(...) wrapping @grst/api-client (listUsers, getUserProfile)
│   ├── agents/assistant-agent.ts ← new Agent({ id, name, instructions, model, tools })
│   ├── workflows/example-workflow.ts ← one sample multi-step workflow
│   └── index.ts                ← export const mastra = new Mastra({ agents, workflows })
├── .env.example                ← model provider key + API_BASE_URL + optional service token
├── tsconfig.json               ← ES2022 / moduleResolution "bundler" (Mastra requirement)
└── package.json                ← @grst/ai
```

- **Deps (current Mastra):** `@mastra/core@latest` + `zod@^4`; dev-dep `mastra@latest`.
  Scripts: `dev` → `mastra dev` (Studio :4111), `build` → `mastra build`. Node 20+.
- **Contract wiring:** `user-tools.ts` `execute` calls `@grst/api-client` against the Go
  REST API, so the agent answers data questions by actually hitting Go routes with typed
  req/res. Auth via configurable `API_BASE_URL` + optional service bearer token from `.env`.
- **Prerequisite:** a model-provider API key is required to *run* the agent (not to
  build/test). Default provider **Anthropic**, configurable via `.env`. Exact model id
  (`"provider/model-name"`) verified against Mastra's provider registry at implementation.
- **Ports conflict-free:** api :3000 (HTTP; gRPC :50051) · web :5173 · ai :4111.

**Design payoff:** the AI app has no independent data path — it borrows the API's, so it
inherits the Go server's auth policy, validation, and business rules. One API surface,
three consumers.

---

## 7. Orchestration, Tooling, Testing & CI

**Root `package.json` (Bun workspaces = `["apps/web", "apps/ai", "packages/*"]`;
`apps/api` is NOT a member — orchestrated via `make`):**

| Script | Does |
|--------|------|
| `bun run proto` | build `protoc-gen-fiber` → `buf generate` → Go + TS clients |
| `bun run dev` | `api` + `web` + `ai` concurrently (:3000 / :5173 / :4111) |
| `bun run build` | api (make) · api-client (tsc) · web (vite build) · ai (mastra build) |
| `bun run test` | api (`go test -race`) · api-client/web/ai (vitest) |
| `bun run lint` / `fmt` | golangci-lint (api) · eslint/prettier (TS) |

Per-app scripts stay local; `apps/api` keeps its exact current `make` targets.

**Testing (each unit tested in isolation):**
- **api** — existing `go test -race ./...`, unchanged, from `apps/api`.
- **api-client** — vitest over envelope decode + `ApiError` mapping with mocked `fetch`
  (highest-value TS test: guards the contract boundary).
- **web** — vitest + Testing Library for login→me with `@grst/api-client` mocked; routes
  compile-time-checked by TanStack Router.
- **ai** — vitest over `user-tools` with `@grst/api-client` mocked (no live model in CI).

**Error handling (one boundary):** `client.ts` is the sole place the `{success,error}`
envelope is decoded → downstream gets typed data or a thrown `ApiError`. `web` renders it
via Query error states; `ai` tools catch and return a structured error to the agent.

**CI:** `.github/workflows/ci.yml` split into path-filtered jobs — the Go job gains
`working-directory: apps/api`; new `web`/`ai`/`api-client` jobs run `bun install` +
typecheck/test, triggered only on their paths. `release.yml` Go build paths gain the
`apps/api` prefix.

---

## 8. Execution Order (each step independently verifiable)

1. **Relocate Go → `apps/api`**; fix buf/compose/Makefile paths.
   **Gate:** `cd apps/api && make proto && go build ./... && go test -race ./...` green.
2. **Workspace + contract**: root Bun workspace, `@grst/tsconfig`, `packages/api-client`
   (add TS output to buf, write `client.ts`).
   **Gate:** `bun run proto` emits TS; `@grst/api-client` builds & its vitest passes.
3. **Scaffold `apps/web`** (Vite/React/TanStack/Tauri, login→me flow).
   **Gate:** login→me works against a locally running `apps/api`.
4. **Scaffold `apps/ai`** (Mastra agent + `user-tools` calling the client).
   **Gate:** agent tool successfully calls the api via the client (mocked in CI).
5. **CI/release + docs** (`ci.yml`, `release.yml`, README, `claude.md`).
   **Gate:** CI green across all jobs.

---

## 9. Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Relocation breaks Go imports/build | Module name unchanged; move whole module with `go.mod`; hard green gate in step 1. |
| `docs/superpowers/` swept into `apps/api` during move | Move only the Go docs package (`docs/scalar.go`); keep `docs/superpowers/` at root. |
| Tauri needs Rust toolchain | Browser dev (`bun run dev`) works without it; Tauri is opt-in. Documented prerequisite. |
| Mastra needs a model API key | Required only to *run* the agent, not to build/test; tools mocked in CI. |
| Hand-written `client.ts` drifts from `grst.route` | Documented; deferred `protoc-gen-ts-client` restores full generation symmetry. |
| CI paths assume root Go | Step 5 updates `ci.yml`/`release.yml` with `apps/api` working-directory + path filters. |

---

## 10. Future Follow-ups (out of scope this pass)

- `protoc-gen-ts-client` — generate the typed REST client from `grst.route` (mirrors `protoc-gen-fiber`).
- ConnectRPC end-to-end (Approach C) — typed streaming RPC for web/ai.
- Shared Go module via `go.work` if a second Go service appears.
- Shared UI package (`packages/ui`) once `web` grows.
- Build `apps/web` and `apps/ai` out to full features (separate design cycles).
