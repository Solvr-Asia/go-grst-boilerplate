import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  component: HomePage,
});

function HomePage() {
  return (
    <div>
      <h1>@grst/web</h1>
      <p>
        React + TanStack Router + Tauri starter, wired to the Go API via
        <code> @grst/api-client</code>. Head to <strong>Login</strong> to
        exercise the auth flow end-to-end.
      </p>
    </div>
  );
}
