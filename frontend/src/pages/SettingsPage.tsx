import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listBlocks, registerPushToken, unblockUser } from "../api";
import { useAuth } from "../context/AuthContext";
import { registerDeviceKey } from "../lib/e2ee";

export function SettingsPage() {
  const { token } = useAuth();
  const [pushToken, setPushToken] = useState("");
  const [msg, setMsg] = useState<string | null>(null);
  const [blocks, setBlocks] = useState<{ blocked_id: string }[]>([]);
  const deviceId = localStorage.getItem("echoline_device") ?? "web";

  useEffect(() => {
    if (!token) return;
    listBlocks(token).then(setBlocks).catch(() => setBlocks([]));
  }, [token]);

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
      <section>
        <h3>Blocked users</h3>
        {blocks.length === 0 && <p className="hint">No blocked users</p>}
        <ul>
          {blocks.map((b) => (
            <li key={b.blocked_id}>
              {b.blocked_id.slice(0, 8)}…
              <button
                type="button"
                onClick={() => unblockUser(token, b.blocked_id)
                  .then(() => listBlocks(token).then(setBlocks))
                  .catch((e) => setMsg(String(e)))}
              >
                Unblock
              </button>
            </li>
          ))}
        </ul>
      </section>
      {msg && <p className="hint">{msg}</p>}
    </main>
  );
}
