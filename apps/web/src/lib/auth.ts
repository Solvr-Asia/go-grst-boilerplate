// PASETO token storage.
//
// SECURITY TRADEOFF: this scaffold uses localStorage so the browser build works
// with zero server changes (Approach A). localStorage is readable by any script
// on the page, so a successful XSS lets an attacker exfiltrate the token. Harden
// for production by choosing one of:
//   - httpOnly, Secure, SameSite cookies (needs the API to set/read cookies), or
//   - keeping the token in memory only (lost on reload; pair with a refresh flow), or
//   - the Tauri secure store / OS keychain for the desktop build.
// The getToken/setToken/clearToken interface is the single seam to swap — callers
// don't change.
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
