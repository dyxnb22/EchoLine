import { useEffect, useState } from "react";
import { AdminReport, AdminUser, adminListReports, adminListUsers } from "../api";

type Props = {
  token: string;
  onClose: () => void;
};

export function AdminPanel({ token, onClose }: Props) {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [reports, setReports] = useState<AdminReport[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([adminListUsers(token), adminListReports(token)])
      .then(([u, r]) => {
        setUsers(u);
        setReports(r);
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
    </aside>
  );
}
