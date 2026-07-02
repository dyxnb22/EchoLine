import { useEffect, useState } from "react";
import { Message, listReplies, sendReply } from "../api";

type Props = {
  token: string;
  convId: string;
  parentMessage: Message;
  onClose: () => void;
};

export function ThreadPanel({ token, convId, parentMessage, onClose }: Props) {
  const [replies, setReplies] = useState<Message[]>([]);
  const [draft, setDraft] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listReplies(token, convId, parentMessage.id)
      .then(setReplies)
      .catch((e) => setError(String(e)));
  }, [token, convId, parentMessage.id]);

  async function handleReply(e: React.FormEvent) {
    e.preventDefault();
    if (!draft.trim()) return;
    try {
      await sendReply(token, convId, parentMessage.id, draft.trim());
      setDraft("");
      const next = await listReplies(token, convId, parentMessage.id);
      setReplies(next);
    } catch (err) {
      setError(String(err));
    }
  }

  return (
    <aside className="thread-panel">
      <header>
        <strong>Thread #{parentMessage.seq}</strong>
        <button type="button" onClick={onClose}>Close</button>
      </header>
      <p className="thread-parent">{parentMessage.body}</p>
      <div className="thread-replies">
        {replies.map((r) => (
          <div key={r.id} className="message">
            <strong>#{r.seq}</strong> {r.body}
          </div>
        ))}
      </div>
      <form onSubmit={handleReply} className="composer">
        <input value={draft} onChange={(e) => setDraft(e.target.value)} placeholder="Reply in thread" />
        <button type="submit">Reply</button>
      </form>
      {error && <p className="error toast">{error}</p>}
    </aside>
  );
}
