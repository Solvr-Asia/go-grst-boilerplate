import { Mastra } from "@mastra/core";
import { assistantAgent } from "./agents/assistant-agent";
import { userReportWorkflow } from "./workflows/example-workflow";

// The Mastra instance registered agents/workflows are served from. `mastra dev`
// (Studio on :4111) and `mastra build` discover this export.
export const mastra = new Mastra({
  agents: { assistantAgent },
  workflows: { userReportWorkflow },
});
