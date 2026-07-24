import { tool } from "@langchain/core/tools";
import { z } from "zod";
import type { UserDirectory } from "../../domain/users";

// A LangGraph/LangChain tool that lets the agent read registered users from the
// Go API, via the UserDirectory port (not the raw client) so it stays testable.
//
// SECURITY — prompt injection / privilege escalation:
// This tool runs with the AI service's identity (API_SERVICE_TOKEN), NOT the
// end user's. If the agent is exposed to untrusted input, a prompt-injection
// attack can make it call this tool and surface user records (including PII like
// email) into the model context, where they may be exfiltrated. The LLM is NOT a
// trustworthy authorization boundary. Before exposing this beyond a demo:
//   - Give API_SERVICE_TOKEN the LEAST privilege it needs — do not reuse an admin
//     token. Scope it to exactly the endpoints the agent may call.
//   - Enforce authorization on the API side per request (the Go route already
//     requires admin role); consider a dedicated, narrowly-scoped service role.
//   - Prefer aggregates over raw PII when the use case allows (e.g. return a
//     count rather than full records).
//   - Add input/output guardrails and human approval for sensitive tools
//     (LangGraph interrupts make this straightforward).
export function makeListUsersTool(users: UserDirectory) {
  return tool(
    async ({ page, size }) => {
      const res = await users.listUsers({ page, size });
      // Tools must return a string (or content parts); JSON keeps it structured.
      return JSON.stringify({ users: res.users, total: res.total });
    },
    {
      name: "list_users",
      description:
        "List registered users from the backend API. Use when asked how many users exist or who is registered. Requires admin scope on the API service token.",
      schema: z.object({
        page: z.number().int().min(1).default(1).describe("1-based page number"),
        size: z.number().int().min(1).max(100).default(10).describe("page size, max 100"),
      }),
    },
  );
}
