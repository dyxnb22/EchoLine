import { useEffect, useState } from "react";
import { AdminReport, AdminUser, adminListDLQ, adminListReports, adminListUsers, adminReplayDLQ, DLQEvent } from "../api";

type Props = {
  token: string;
  onClose: () => void;
};

export function AdminPanel({ token, onClose }: Props) {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [reports, setReports] = useState<AdminReport[]>([]);
  const [dlq, setDlq] = useState<DLQEvent[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([adminListUsers(token), adminListReports(token), adminListDLQ(token)])
      .then(([u, r, d]) => {
        setUsers(u);
        setReports(r);
        setDlq(d);
      })
      .catch((e) => setError(String(e)));
  }, [token]);

  return (
    <aside className="admin-panel">
      <header>
        <strong>Admin</strong>
        <button type="button" onClick={onClose}>Close</button>
      </header>
      {error && <p className="error toast">{error}</p>}
      {status && <p className="hint">{status}</p>}
      <section>
        <h3>Users ({users.length})</h3>
        <ul>
          {users.map((u) => (
            <li key={u.id}>{u.username} {u.is_admin ? "(admin)" : ""}</li>
          ))}
        </ul>
      </section>
      <section>
        <h3>Reports ({reports.length})</h3>
        <ul>
          {reports.map((r) => (
            <li key={r.id}>{r.reason} — msg {r.message_id.slice(0, 8)}</li>
          ))}
        </ul>
      </section>
      <section>
        <h3>DLQ ({dlq.length})</h3>
        <ul>
          {dlq.map((e) => (
            <li key={e.id}>
              {e.event_type} ({e.status}, {e.attempts} attempts)
              <button
                type="button"
                onClick={() => adminReplayDLQ(token, e.id)
                  .then(() => setStatus(`replayed ${e.id.slice(0, 8)}`))
                  .catch((err) => setError(String(err)))}
              >
                Replay
              </button>
            </li>
          ))}
        </ul>
      </section>
    </aside>
  );
}
