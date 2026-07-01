import { useState } from "react";
import { Link } from "react-router-dom";
import { registerPushToken } from "../api";
import { useAuth } from "../context/AuthContext";
import { registerDeviceKey } from "../lib/e2ee";

export function SettingsPage() {
  const { token } = useAuth();
  const [pushToken, setPushToken] = useState("");
  const [msg, setMsg] = useState<string | null>(null);
  const deviceId = localStorage.getItem("echoline_device") ?? "web";

  if (!token) return null;

  return (
    <main className="shell settings-page">
      <h1>Settings</h1>
      <p><Link to="/">← Back to chat</Link></p>
      <section>
        <h3>Push notifications</h3>
        <input value={pushToken} onChange={(e) => setPushToken(e.target.value)} placeholder="FCM/APNs token" />
        <button
          type="button"
          onClick={() => registerPushToken(token, deviceId, pushToken, "web").then(() => setMsg("push registered")).catch((e) => setMsg(String(e)))}
        >
          Register push token
        </button>
      </section>
      <section>
        <h3>E2EE key bundle</h3>
        <button
          type="button"
          onClick={() => registerDeviceKey(token, deviceId, `pk-${deviceId}`).then(() => setMsg("key registered")).catch((e) => setMsg(String(e)))}
        >
          Register device public key
        </button>
      </section>
      {msg && <p className="hint">{msg}</p>}
    </main>
  );
}
