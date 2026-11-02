import axios, { type AxiosInstance } from 'axios';

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
    };
  }

  if (error instanceof Error) {
    return { code: 'UNEXPECTED_ERROR', message: error.message };
  }

  return { code: 'UNEXPECTED_ERROR', message: DEFAULT_ERROR_MESSAGE };
}
