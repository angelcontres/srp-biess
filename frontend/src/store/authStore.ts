/**
 * In-memory session store (SDD 6.3).
 *
 * The access token lives ONLY here, in a module-scoped variable: never in
 * localStorage, sessionStorage or a cookie readable by JavaScript. Reloading the
 * page drops it on purpose; SRP-015 restores the session with POST /auth/refresh.
 */
let accessToken: string | null = null;

export function getAccessToken(): string | null {
  return accessToken;
}

export function setAccessToken(token: string | null): void {
  accessToken = token;
}

export function clearSession(): void {
  accessToken = null;
}
