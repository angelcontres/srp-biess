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

type SessionExpiredListener = () => void;
const expiredListeners = new Set<SessionExpiredListener>();

/** Subscribes to session expiry (a failed refresh while signed in). Returns the unsubscribe function. */
export function onSessionExpired(listener: SessionExpiredListener): () => void {
  expiredListeners.add(listener);
  return () => {
    expiredListeners.delete(listener);
  };
}

/** Clears the session and notifies the app that it expired (SRP-015). */
export function expireSession(): void {
  accessToken = null;
  expiredListeners.forEach((listener) => listener());
}
