import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { ZodTypeAny } from 'zod';
import { borradorValido } from '../schemas/resolver';

const PREFIX = 'draft:';
const DEBOUNCE_MS = 2000;

interface DraftEnvelope<T> {
  values: T;
  savedAt: string;
}

/** SHA-256 of the cédula: the draft key never contains it in clear (DESIGN states.md). */
export async function hashCedula(cedula: string): Promise<string> {
  const bytes = new TextEncoder().encode(cedula);
  const digest = await crypto.subtle.digest('SHA-256', bytes);
  return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, '0')).join('');
}

/** Removes every local draft. Call it on logout and on session expiry. */
export function clearAllDrafts(): void {
  try {
    Object.keys(sessionStorage)
      .filter((key) => key.startsWith(PREFIX))
      .forEach((key) => sessionStorage.removeItem(key));
  } catch {
    // Storage unavailable: nothing to clean.
  }
}

function read<T>(key: string): DraftEnvelope<T> | null {
  try {
    const raw = sessionStorage.getItem(key);
    return raw ? (JSON.parse(raw) as DraftEnvelope<T>) : null;
  } catch {
    return null;
  }
}

/**
 * Local draft in sessionStorage (DESIGN states.md): 2 s debounce, key
 * `draft:{hashDeCédula}:{idSección}`, validated with the section schema before
 * restoring. It never replaces "Guardar progreso" and never holds files
 * (only the server references of already uploaded PDFs).
 */
export function useDraftCache<T>(namespace: string, sectionId: string, schema: ZodTypeAny) {
  const key = `${PREFIX}${namespace}:${sectionId}`;
  const timer = useRef<ReturnType<typeof setTimeout>>();
  const [restored] = useState<DraftEnvelope<T> | null>(() => {
    const draft = read<T>(key);
    if (draft && !borradorValido(schema, draft.values)) {
      try {
        sessionStorage.removeItem(key);
      } catch {
        // ignore
      }
      return null;
    }
    return draft;
  });

  const schedule = useCallback(
    (values: T) => {
      clearTimeout(timer.current);
      timer.current = setTimeout(() => {
        try {
          const envelope: DraftEnvelope<T> = { values, savedAt: new Date().toISOString() };
          sessionStorage.setItem(key, JSON.stringify(envelope));
        } catch {
          // Storage full or blocked: the form keeps working without the draft.
        }
      }, DEBOUNCE_MS);
    },
    [key],
  );

  const clear = useCallback(() => {
    clearTimeout(timer.current);
    try {
      sessionStorage.removeItem(key);
    } catch {
      // ignore
    }
  }, [key]);

  useEffect(() => () => clearTimeout(timer.current), []);

  return useMemo(() => ({ restored, schedule, clear }), [restored, schedule, clear]);
}
