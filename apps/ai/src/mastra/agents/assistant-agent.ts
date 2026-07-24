import { Agent } from "@mastra/core/agent";
import { listUsersTool } from "../tools/user-tools";

// Model is a "provider/model-name" string (Mastra model router). It is resolved
// at run time and requires the matching provider API key in the environment.
export const assistantAgent = new Agent({
  id: "assistant-agent",
  name: "Assistant Agent",
  instructions: `You are a helpful assistant for the GRST platform.
You can look up registered users with the list-users tool. When asked about
users, call the tool rather than guessing. Keep answers concise.`,
  model: process.env.AI_MODEL ?? "anthropic/claude-sonnet-4-5",
  tools: { listUsersTool },
});
