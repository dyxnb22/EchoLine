import { useState } from "react";
import {
  createChannelConversation,
  createDirectConversation,
  createGroupConversation,
} from "../api";

type Props = {
  token: string;
  onCreated: (id: string) => void;
  onClose: () => void;
};

export function CreateConversationModal({ token, onCreated, onClose }: Props) {
  const [mode, setMode] = useState<"direct" | "group" | "channel">("direct");
  const [peerId, setPeerId] = useState("");
  const [title, setTitle] = useState("");
  const [members, setMembers] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      if (mode === "direct") {
        const conv = await createDirectConversation(token, peerId.trim());
        onCreated(String(conv.id));
      } else if (mode === "group") {
        const memberIds = members.split(",").map((s) => s.trim()).filter(Boolean);
        const conv = await createGroupConversation(token, title.trim() || "Group", memberIds);
        onCreated(String(conv.id));
      } else {
        const conv = await createChannelConversation(token, title.trim() || "Channel");
        onCreated(String(conv.id));
      }
      onClose();
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="modal-overlay">
      <form className="modal" onSubmit={handleSubmit}>
        <h3>New conversation</h3>
        <div className="tabs">
          {(["direct", "group", "channel"] as const).map((m) => (
            <button key={m} type="button" className={mode === m ? "active" : ""} onClick={() => setMode(m)}>
              {m}
            </button>
          ))}
        </div>
        {mode === "direct" && (
          <input value={peerId} onChange={(e) => setPeerId(e.target.value)} placeholder="Peer user UUID" required />
        )}
        {(mode === "group" || mode === "channel") && (
          <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Title" />
        )}
        {mode === "group" && (
          <input value={members} onChange={(e) => setMembers(e.target.value)} placeholder="Member UUIDs (comma-separated)" />
        )}
        {error && <p className="error">{error}</p>}
        <div className="modal-actions">
          <button type="button" onClick={onClose}>Cancel</button>
          <button type="submit" disabled={loading}>{loading ? "Creating..." : "Create"}</button>
        </div>
      </form>
    </div>
  );
}
