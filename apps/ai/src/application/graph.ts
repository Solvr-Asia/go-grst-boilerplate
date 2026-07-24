// The shape the HTTP layer needs from any runnable graph (agent or workflow).
//
// LangGraph's compiled agents and StateGraphs both satisfy this once adapted in
// the composition root; the interface keeps the transport layer free of
// LangGraph-specific option plumbing (thread_id, streamMode).

export interface Graph {
  invoke(input: unknown, opts: { threadId: string }): Promise<unknown>;
  stream(input: unknown, opts: { threadId: string }): Promise<AsyncIterable<unknown>>;
}
