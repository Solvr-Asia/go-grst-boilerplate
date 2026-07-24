# @grst/web

React 19 + TanStack Router + TanStack Query, bundled by Vite and wrapped in a
Tauri v2 desktop shell. Talks to the Go API (`apps/api`) through the shared
`@grst/api-client` (typed REST over the proto contract).

## Develop

```bash
bun install                 # from repo root
bun --filter @grst/web run dev      # browser dev server on http://localhost:5173
```

Set the API base URL via `.env` (see `.env.example`); it defaults to
`http://localhost:3000`. For the login flow to work, add the Vite origin to the
Go server's `CORS_ORIGINS` (see `apps/api/.env.example`).

## Desktop (Tauri)

The desktop shell in `src-tauri/` requires a **Rust toolchain** (`rustup`,
`cargo`) plus your OS's webview dependencies — the browser dev flow above does
not. Once Rust is installed:

```bash
bun --filter @grst/web run tauri:dev     # desktop window (spawns Vite)
bun --filter @grst/web run tauri:build   # production desktop bundle
```

Generate the bundle icons once (writes `src-tauri/icons/`):

```bash
bun --filter @grst/web run tauri icon path/to/logo.png
```

## Routes

File-based (`src/routes/`), compiled to a typed route tree
(`src/routeTree.gen.ts`) by `@tanstack/router-plugin`:

- `/` — home
- `/login` — email/password → `api.login()` → stores token → `/me`
- `/me` — `api.getMe()` + `api.listUsers()` (admin-only list)
