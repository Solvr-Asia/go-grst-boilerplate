// Public surface of @veemon/api-client.
//
// - The typed REST client and its lean wire DTOs (the everyday API).
// - The full protobuf-es message types under the `proto` namespace, generated
//   from contract/*.proto (the single source of truth), for consumers that need
//   the complete message schemas.
export * from "./client.js";
export * as proto from "./gen/user/user_pb.js";
