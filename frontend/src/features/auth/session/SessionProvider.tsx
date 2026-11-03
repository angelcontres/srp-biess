import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { clearAllDrafts } from '@/features/postulacion/hooks/useDraftCache';
import { api, refreshAccessToken } from '@/services/api/client';
import { clearSession, onSessionExpired } from '@/store/authStore';

export type SessionStatus = 'restoring' | 'authenticated' | 'anonymous';

interface SessionContextValue {
  status: SessionStatus;
  /** Email of the signed-in applicant. */
  user?: string;
  /** True after the session expired, until the applicant signs in again. */
  expired: boolean;
  signIn: (email: string) => void;
  signOut: () => Promise<void>;
}

const SessionContext = createContext<SessionContextValue | null>(null);

/**
 * Global session state (SRP-015). Restores the session on load with the HttpOnly
 * refresh cookie, and ends it when a refresh fails while signed in.
 */
export function SessionProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<SessionStatus>('restoring');
  const [user, setUser] = useState<string>();
  const [expired, setExpired] = useState(false);

  useEffect(() => {
    let activo = true;
    void (async () => {
      let email: string | undefined;
      if (await refreshAccessToken()) {
        try {
          const { data } = await api.get<{ data: { email: string } }>('/auth/me');
          email = data.data.email;
        } catch {
          clearSession(); // the restored session is not valid
        }
      }
      if (!activo) return;
      setUser(email);
      setStatus(email ? 'authenticated' : 'anonymous');
    })();
    return () => {
      activo = false;
    };
  }, []);

  useEffect(
    () =>
      onSessionExpired(() => {
        clearAllDrafts(); // the local draft never outlives the session (DESIGN states.md)
        setUser(undefined);
        setExpired(true);
        setStatus('anonymous');
      }),
    [],
  );

  const signIn = useCallback((email: string) => {
    setUser(email);
    setExpired(false);
    setStatus('authenticated');
  }, []);

  const signOut = useCallback(async () => {
    try {
      await api.post('/auth/logout'); // the API revokes the session and deletes the cookie
    } catch {
      // the local session is cleared even if the request fails
    } finally {
      clearSession();
      clearAllDrafts();
    }
    setUser(undefined);
    setExpired(false);
    setStatus('anonymous');
  }, []);

  const value = useMemo(
    () => ({ status, user, expired, signIn, signOut }),
    [status, user, expired, signIn, signOut],
  );
  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession(): SessionContextValue {
  const context = useContext(SessionContext);
  if (!context) throw new Error('useSession must be used inside <SessionProvider>.');
  return context;
}
