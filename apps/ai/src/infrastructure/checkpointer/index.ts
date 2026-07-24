import { MemorySaver, type BaseCheckpointSaver } from "@langchain/langgraph";
import { PostgresSaver } from "@langchain/langgraph-checkpoint-postgres";
import type { Config } from "../../config";

// The checkpointer persists per-thread graph state. It is what turns the graphs
// into durable, resumable conversations: pass a `thread_id` and every run
// appends to that thread's history, enabling multi-turn memory, resume after a
// crash, and human-in-the-loop interrupts.
//
//   memory   -> MemorySaver: in-process, lost on restart. Good for dev/tests.
//   postgres -> PostgresSaver over apps/api's Postgres. Durable across restarts.
export async function createCheckpointer(config: Config): Promise<BaseCheckpointSaver> {
  if (config.checkpointer.kind === "postgres") {
    const saver = PostgresSaver.fromConnString(config.checkpointer.connectionString);
    // Creates the checkpoint tables if they don't exist yet (idempotent).
    await saver.setup();
    return saver;
  }
  return new MemorySaver();
}
