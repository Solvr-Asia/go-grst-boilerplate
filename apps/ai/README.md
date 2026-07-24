# @grst/ai

A [Mastra](https://mastra.ai) (v1) service. Its agent reaches the Go API
(`apps/api`) through the shared `@grst/api-client`, so the AI inherits the Go
server's auth policy and validation — one API surface, three consumers.

## Structure

```
src/mastra/
  api.ts                     configured @grst/api-client instance
  tools/user-tools.ts        list-users tool (calls api.listUsers)
  agents/assistant-agent.ts  agent with the model + tools
  workflows/example-workflow.ts  one-step "user-report" workflow
  index.ts                   the Mastra instance (agents + workflows)
```

## Run

```bash
bun install                          # from repo root
bun --filter @grst/ai run dev        # Mastra Studio on http://localhost:4111
```

Running the agent requires a **model provider API key** (`.env`, see
`.env.example`) — building and testing do not. Configure the model with
`AI_MODEL` (`"provider/model-name"`, default `anthropic/claude-sonnet-4-5`) and
the target API with `API_BASE_URL` / `API_SERVICE_TOKEN`.

## Test

```bash
bun --filter @grst/ai run test       # tool logic, api-client mocked
```
