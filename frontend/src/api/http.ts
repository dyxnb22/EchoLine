export const API_BASE = import.meta.env.VITE_API_BASE ?? "";

export function authHeaders(token: string): HeadersInit {
  return {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };
}

export type ApiErrorBody = {
  error?: { code?: string; message?: string };
  message?: string;
};

/** Parses JSON API responses and surfaces server error messages. */
export async function parseResponse<T>(res: Response, fallback: string): Promise<T> {
  if (res.ok) {
    if (res.status === 204) return undefined as T;
    return res.json() as Promise<T>;
  }
  let message = fallback;
  try {
    const body = (await res.json()) as ApiErrorBody;
    message = body.error?.message ?? body.message ?? fallback;
  } catch {
    // non-JSON error body
  }
  throw new Error(message);
}

export type AuthFetch = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

let globalAuthFetch: AuthFetch | null = null;

/** Called from AuthProvider so API modules can use refresh-aware fetch. */
export function bindAuthFetch(fn: AuthFetch | null) {
  globalAuthFetch = fn;
}

/** Performs an authenticated request, preferring AuthContext refresh wrapper when bound. */
export async function authedRequest(
  token: string,
  path: string,
  init: RequestInit = {},
): Promise<Response> {
  const url = path.startsWith("http") ? path : `${API_BASE}${path}`;
  const headers = new Headers(init.headers);
  if (!headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const requestInit = { ...init, headers };
  if (globalAuthFetch) {
    return globalAuthFetch(url, requestInit);
  }
  return fetch(url, requestInit);
}
