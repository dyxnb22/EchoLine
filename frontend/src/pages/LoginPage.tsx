import { useState } from "react";
import { login, register } from "../api";

type Props = {
  onLogin: (token: string) => void;
};

export function LoginPage({ onLogin }: Props) {
  const [username, setUsername] = useState("alice");
  const [password, setPassword] = useState("secret123");
  const [displayName, setDisplayName] = useState("");
  const [authMode, setAuthMode] = useState<"login" | "register">("login");
  const [error, setError] = useState<string | null>(null);

  async function handleAuth(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      if (authMode === "register") {
        await register(username, password, displayName || username);
      }
      const tokens = await login(username, password);
      localStorage.setItem("echoline_token", tokens.access_token);
      localStorage.setItem("echoline_refresh", tokens.refresh_token);
      onLogin(tokens.access_token);
    } catch (err) {
      setError(String(err));
    }
  }

  return (
    <main className="shell">
      <h1>EchoLine</h1>
      <form onSubmit={handleAuth} className="card">
        <div className="auth-tabs">
          <button type="button" className={authMode === "login" ? "active" : ""} onClick={() => setAuthMode("login")}>Login</button>
          <button type="button" className={authMode === "register" ? "active" : ""} onClick={() => setAuthMode("register")}>Register</button>
        </div>
        <input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="username" />
        {authMode === "register" && (
          <input value={displayName} onChange={(e) => setDisplayName(e.target.value)} placeholder="display name" />
        )}
        <input value={password} onChange={(e) => setPassword(e.target.value)} type="password" placeholder="password" />
        <button type="submit">{authMode === "login" ? "Login" : "Register & Login"}</button>
        {error && <p className="error toast">{error}</p>}
      </form>
    </main>
  );
}
