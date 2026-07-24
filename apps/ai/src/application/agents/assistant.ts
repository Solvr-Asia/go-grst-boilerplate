import { createReactAgent } from "@langchain/langgraph/prebuilt";
import type { BaseCheckpointSaver } from "@langchain/langgraph";
import type { LanguageModelLike } from "@langchain/core/language_models/base";
import type { StructuredToolInterface } from "@langchain/core/tools";

const ASSISTANT_PROMPT = `You are a helpful assistant for the GRST platform.
You can look up registered users with the list_users tool. When asked about
users, call the tool rather than guessing. Keep answers concise.`;

// A tool-calling ReAct agent built from LangGraph's prebuilt. The checkpointer
// gives it durable per-thread memory (pass a thread_id at run time). Swap in
// more tools or a custom StateGraph here as capabilities grow.
export function buildAssistantAgent(deps: {
  model: LanguageModelLike;
  tools: StructuredToolInterface[];
  checkpointer: BaseCheckpointSaver;
}) {
  return createReactAgent({
    llm: deps.model,
    tools: deps.tools,
    checkpointSaver: deps.checkpointer,
    prompt: ASSISTANT_PROMPT,
  });
}
