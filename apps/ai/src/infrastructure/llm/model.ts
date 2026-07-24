import { ChatAnthropic } from "@langchain/anthropic";
import type { Config } from "../../config";

// The chat model powering the agents. LangChain's ChatAnthropic wraps the
// official Anthropic SDK; the model id and token budget come from config.
//
// The constructor does not require the API key to be present — it is read
// lazily on the first request, so the server can boot (and serve /health) even
// when ANTHROPIC_API_KEY is unset. A missing key surfaces as a clear error only
// when an agent is actually invoked.
export function createChatModel(config: Config): ChatAnthropic {
  return new ChatAnthropic({
    model: config.model.name,
    maxTokens: config.model.maxTokens,
    apiKey: config.anthropicApiKey,
  });
}
