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
    const text = await res.text();
    if (!text) return undefined as T;
    return JSON.parse(text) as T;
  }
  let message = fallback;
  try {
    const body = JSON.parse(await res.text()) as ApiErrorBody;
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

export async function publicJSON<T>(path: string, init: RequestInit, fallback: string): Promise<T> {
  const url = path.startsWith("http") ? path : `${API_BASE}${path}`;
  const res = await fetch(url, init);
  return parseResponse<T>(res, fallback);
}

export async function authedJSON<T>(
  token: string,
  path: string,
  init: RequestInit,
  fallback: string,
): Promise<T> {
  const res = await authedRequest(token, path, init);
  return parseResponse<T>(res, fallback);
}

export async function authedVoid(
  token: string,
  path: string,
  init: RequestInit,
  fallback: string,
): Promise<void> {
  const res = await authedRequest(token, path, init);
  await parseResponse(res, fallback);
}

/** Returns defaultValue when the server responds with a non-OK status. */
export async function authedJSONOr<T>(
  token: string,
  path: string,
  init: RequestInit,
  defaultValue: T,
): Promise<T> {
  const res = await authedRequest(token, path, init);
  if (!res.ok) return defaultValue;
  try {
    return (await res.json()) as T;
  } catch {
    return defaultValue;
  }
}

export async function authedBlob(
  token: string,
  path: string,
  init: RequestInit,
  fallback: string,
): Promise<Blob> {
  const res = await authedRequest(token, path, init);
  if (!res.ok) {
    await parseResponse(res, fallback);
  }
  return res.blob();
}
