import { StateGraph, START, END, Annotation, type BaseCheckpointSaver } from "@langchain/langgraph";
import type { UserDirectory } from "../../domain/users";

// A minimal explicit StateGraph demonstrating the workflow primitive: fetch a
// page of users from the Go API and report a count. Unlike the agent (which the
// model drives), a workflow is deterministic control flow you compose from nodes
// and edges — add nodes/branches here for real multi-step pipelines.
const UserReportState = Annotation.Root({
  page: Annotation<number>({ reducer: (_prev, next) => next, default: () => 1 }),
  size: Annotation<number>({ reducer: (_prev, next) => next, default: () => 10 }),
  count: Annotation<number>({ reducer: (_prev, next) => next, default: () => 0 }),
  total: Annotation<number | undefined>({ reducer: (_prev, next) => next, default: () => undefined }),
});

export function buildUserReportWorkflow(deps: {
  users: UserDirectory;
  checkpointer: BaseCheckpointSaver;
}) {
  return new StateGraph(UserReportState)
    .addNode("countUsers", async (state) => {
      const res = await deps.users.listUsers({ page: state.page, size: state.size });
      return { count: res.users.length, total: res.total };
    })
    .addEdge(START, "countUsers")
    .addEdge("countUsers", END)
    .compile({ checkpointer: deps.checkpointer });
}
