// PASETO token storage. Browser build uses localStorage; a Tauri build could
// swap this for the Tauri store/secure storage without touching callers.
const KEY = "grst.token";

export function getToken(): string | null {
  return localStorage.getItem(KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(KEY);
}
