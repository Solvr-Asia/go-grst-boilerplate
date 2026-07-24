# Architecture Layers (apps/api — the Go backend)

```
contract/       → Proto contracts (source of truth for gRPC + REST routes; at repo root)
cmd/server/     → API server entry point (HTTP + gRPC)
cmd/worker/     → RabbitMQ consumer entry point
cmd/migrate/    → migration + seeding CLI
cmd/protoc-gen-fiber/ → Codegen plugin: proto veemon.route → Fiber routes
config/         → Configuration (+ Validate), bootstrap/DI, and infrastructure init
handler/        → Presentation layer (gRPC handler impl + generated Fiber routes)
app/usecase/    → Business logic layer (depends on the repository interface)
repository/     → Data access layer (GORM)
entity/         → Domain entities
pkg/            → Shared infrastructure (token, authguard, middleware, rabbitmq,
                  redis, database, resilience, metrics, telemetry, logger, errors, …)
migrations/     → golang-migrate SQL files (the schema source of truth)
```

**REST routes are generated, not hand-written.** Declare a route with a
`veemon.route` option on the RPC in `contract/<svc>/<svc>.proto` (method, path,
`auth { required, roles }`, `body`, `response`, `rate_limit`), then run
`make proto`. `protoc-gen-fiber` emits `handler/grpc/<svc>/<svc>_fiber.pb.go`
(`Register<Svc>Routes` + the `<Svc>AuthConfig` gRPC auth map). Never add a Fiber
route or auth map by hand — change the proto and regenerate. See the
[proto-routes skill](../skills/proto-routes/SKILL.md) and the README
"Declaring Routes in Proto".

## Data Flow (both protocols share the same handler + usecase)

```
HTTP  → generated Fiber routes (handler/grpc/user) ─┐
                                                     ├→ handler.userHandler → UseCase → Repository → DB
gRPC  → UserApi server ──────────────────────────────┘
```

The generated Fiber routes and the gRPC server are both served by the same
`handler.userHandler` (which implements `UserApiServer`), so REST and gRPC
share one implementation and one auth policy map.
