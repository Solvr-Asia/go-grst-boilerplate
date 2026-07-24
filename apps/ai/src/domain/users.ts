// Domain port: the user-directory capability agents/workflows depend on.
//
// This is intentionally decoupled from @grst/api-client — the application layer
// depends on this interface, and an adapter in infrastructure/ implements it.
// That keeps the graphs testable with a fake and swappable if the backend moves.

export interface UserSummary {
  id: string;
  email: string;
  name: string;
  status: string;
}

export interface ListUsersQuery {
  page: number;
  size: number;
}

export interface ListUsersResult {
  users: UserSummary[];
  total?: number;
}

export interface UserDirectory {
  listUsers(query: ListUsersQuery): Promise<ListUsersResult>;
}
