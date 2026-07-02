import { useState } from "react";
import { inviteMember, removeMember } from "../api";

type Props = {
  token: string;
  conversationId: string;
};

export function GroupSettings({ token, conversationId }: Props) {
  const [userId, setUserId] = useState("");
  const [status, setStatus] = useState<string | null>(null);

  return (
    <div className="group-settings">
      <strong>Group settings</strong>
      <div className="row">
        <input value={userId} onChange={(e) => setUserId(e.target.value)} placeholder="User ID to invite" />
        <button
          type="button"
          onClick={() => inviteMember(token, conversationId, userId).then(() => setStatus("invited")).catch((e) => setStatus(String(e)))}
        >
          Invite
        </button>
        <button
          type="button"
          onClick={() => removeMember(token, conversationId, userId).then(() => setStatus("removed")).catch((e) => setStatus(String(e)))}
        >
          Kick
        </button>
      </div>
      {status && <p className="hint">{status}</p>}
    </div>
  );
}
