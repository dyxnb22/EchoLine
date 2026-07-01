import { useEffect, useState } from "react";
import { listNotifications, markNotificationRead, Notification } from "../api";

type Props = {
  token: string;
  onClose: () => void;
};

export function NotificationPanel({ token, onClose }: Props) {
  const [items, setItems] = useState<Notification[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listNotifications(token).then(setItems).catch((e) => setError(String(e)));
  }, [token]);

  return (
    <aside className="notif-panel">
      <header>
        <strong>Notifications</strong>
        <button type="button" onClick={onClose}>Close</button>
      </header>
      {error && <p className="error">{error}</p>}
      <ul>
        {items.map((n) => (
          <li key={n.id}>
            <button
              type="button"
              onClick={() => markNotificationRead(token, n.id).then(() => setItems((prev) => prev.filter((x) => x.id !== n.id)))}
            >
              {n.type} {n.read_at ? "(read)" : ""}
            </button>
          </li>
        ))}
      </ul>
    </aside>
  );
}
