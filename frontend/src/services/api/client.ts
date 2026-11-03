import axios, { type AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import { getAccessToken, clearSession, expireSession, setAccessToken } from '@/store/authStore';

let refreshPromise: Promise<string | null> | null = null;

/**
 * Base URL of the SRP-BIESS REST API.
 * Override it through VITE_API_URL in the environment file.
 */
export const API_BASE_URL: string =
  import.meta.env.VITE_API_URL ?? '/api/v1';

/**
 * Shared HTTP client for the SPA (SDD 6.6).
 *
 * `withCredentials` is mandatory: the refresh token travels in an HttpOnly cookie
 * issued by the API, so every request must carry cookies.
 *
 * The access token is NEVER persisted (no localStorage, no sessionStorage). It lives
 * in memory inside the auth store and is attached by the session interceptor (SRP-015).
 */
export const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 15000,
});

/**
 * Attaches the in-memory access token to every request. Without this interceptor
 * the API answers 401 "Cabecera Authorization no proporcionada".
 */
api.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`);
  }
  return config;
});

/** RFC 7807 problem details as returned by the API. */
export interface ApiProblem {
  type?: string;
  title?: string;
  status?: number;
  detail?: string;
  instance?: string;
  code?: string;
  message?: string;
  error?: string;
  fields?: Record<string, string>;
  traceId?: string;
}

/** Shape consumed by the UI layer after every failed request. */
export interface NormalizedApiError {
  code: string;
  message: string;
  status?: number;
  fields?: Record<string, string>;
  traceId?: string;
  /** Seconds from the `Retry-After` header (HTTP 423/429), when present. */
  retryAfterSeconds?: number;
}

/** Lockout length used when a 423 arrives without `Retry-After` (EF-11: 10 minutes). */
export const DEFAULT_LOCKOUT_SECONDS = 600;

function parseRetryAfter(value: unknown): number | undefined {
  const seconds = Number.parseInt(String(value ?? ''), 10);
  return Number.isFinite(seconds) && seconds > 0 ? seconds : undefined;
}

const DEFAULT_ERROR_MESSAGE = 'Ocurrió un error inesperado. Intente nuevamente.';

/**
 * Converts any thrown value into `{ code, message, fields, traceId }`.
 * The UI shows `message` and maps `fields` back to the failing inputs.
 */
export function normalizeError(error: unknown): NormalizedApiError {
  if (axios.isAxiosError<ApiProblem>(error)) {
    const data = error.response?.data;

    if (!error.response) {
      return {
        code: error.code ?? 'NETWORK_ERROR',
        message: 'No se pudo conectar con el servidor. Verifique su conexión.',
      };
    }

    return {
      code: data?.code ?? String(error.response.status),
      message: data?.detail ?? data?.message ?? data?.title ?? data?.error ?? DEFAULT_ERROR_MESSAGE,
      status: error.response.status,
      fields: data?.fields,
      traceId: data?.traceId,
      retryAfterSeconds: parseRetryAfter(error.response.headers?.['retry-after']),
    };
  }

  if (error instanceof Error) {
    return { code: 'UNEXPECTED_ERROR', message: error.message };
  }

  return { code: 'UNEXPECTED_ERROR', message: DEFAULT_ERROR_MESSAGE };
}

/** Pide un access token nuevo usando la cookie HttpOnly. Si llegan varias peticiones a la vez, solo hace una llamada. */
export function refreshAccessToken(): Promise<string | null> {
  refreshPromise ??= axios
    // Se usa axios "crudo" a propósito, sin la instancia `api`, para que no pase por los interceptores y no se forme un bucle
    .post<{ data: { access_token: string } }>(`${API_BASE_URL}/auth/refresh`, null, {
      withCredentials: true, // necesario para que se envíe la cookie
    })
    .then((res) => {
      const token = res.data?.data?.access_token ?? null;
      setAccessToken(token);
      return token;
    })
    .catch(() => {
      // La cookie expiró o se revocó. Si había sesión abierta, se avisa a la app para
      // llevar al usuario al inicio de sesión; al restaurar la sesión al cargar, no.
      if (getAccessToken()) expireSession();
      else clearSession();
      return null;
    })
    .finally(() => {
      refreshPromise = null;
    });
  return refreshPromise;
}

api.interceptors.response.use(undefined, async (error) => {
  const original = error.config as (InternalAxiosRequestConfig & { _retry?: boolean }) | undefined;
  const isAuthEndpoint = /\/auth\/(login|refresh)$/.test(original?.url ?? '');

  if (error.response?.status === 401 && original && !original._retry && !isAuthEndpoint) {
    original._retry = true; // reintenta una sola vez
    const token = await refreshAccessToken();
    if (token) {
      original.headers.set('Authorization', `Bearer ${token}`);
      return api(original);
    }
  }
  return Promise.reject(error);
});
