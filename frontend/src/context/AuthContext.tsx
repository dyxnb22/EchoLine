import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { refreshToken } from "../api";

type AuthContextValue = {
  token: string | null;
  setToken: (t: string | null) => void;
  logout: () => void;
  authFetch: (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setTokenState] = useState<string | null>(localStorage.getItem("echoline_token"));

  const setToken = useCallback((t: string | null) => {
    if (t) {
      localStorage.setItem("echoline_token", t);
    } else {
      localStorage.removeItem("echoline_token");
      localStorage.removeItem("echoline_refresh");
    }
    setTokenState(t);
  }, []);

  const logout = useCallback(() => setToken(null), [setToken]);

  const authFetch = useCallback(async (input: RequestInfo | URL, init?: RequestInit) => {
    let access = token ?? localStorage.getItem("echoline_token");
    const headers = new Headers(init?.headers);
    if (access) headers.set("Authorization", `Bearer ${access}`);
    let res = await fetch(input, { ...init, headers });
    if (res.status === 401) {
      const refresh = localStorage.getItem("echoline_refresh");
      if (refresh) {
        try {
          const pair = await refreshToken(refresh);
          localStorage.setItem("echoline_token", pair.access_token);
          localStorage.setItem("echoline_refresh", pair.refresh_token);
          setTokenState(pair.access_token);
          headers.set("Authorization", `Bearer ${pair.access_token}`);
          res = await fetch(input, { ...init, headers });
        } catch {
          logout();
        }
      }
    }
    return res;
  }, [token, logout]);

  const value = useMemo(() => ({ token, setToken, logout, authFetch }), [token, setToken, logout, authFetch]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth outside AuthProvider");
  return ctx;
}
