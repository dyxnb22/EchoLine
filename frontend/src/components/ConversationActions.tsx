import { useState } from "react";
import {
  archiveConversation,
  exportConversation,
  forwardMessage,
  pinMessage,
  subscribeChannel,
} from "../api";

type Props = {
  token: string;
  conversationId: string;
  conversationType: string;
  messageId?: string;
  onAction?: () => void;
};

export function ConversationActions({ token, conversationId, conversationType, messageId, onAction }: Props) {
  const [busy, setBusy] = useState(false);
  const [forwardTarget, setForwardTarget] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function run(fn: () => Promise<void>) {
    setBusy(true);
    setError(null);
    try {
      await fn();
      onAction?.();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="conv-actions">
      {conversationType === "channel" && (
        <button type="button" disabled={busy} onClick={() => run(() => subscribeChannel(token, conversationId))}>
          Subscribe
        </button>
      )}
      <button type="button" disabled={busy} onClick={() => run(() => archiveConversation(token, conversationId))}>
        Archive
      </button>
      <button type="button" disabled={busy} onClick={() => run(async () => {
        const blob = await exportConversation(token, conversationId);
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `echoline-${conversationId}.json`;
        a.click();
        URL.revokeObjectURL(url);
      })}>
        Export
      </button>
      {messageId && (
        <>
          <button type="button" disabled={busy} onClick={() => run(() => pinMessage(token, conversationId, messageId))}>
            Pin
          </button>
          <input
            value={forwardTarget}
            onChange={(e) => setForwardTarget(e.target.value)}
            placeholder="Forward to conv id"
            className="forward-input"
          />
          <button
            type="button"
            disabled={busy || !forwardTarget}
            onClick={() => run(() => forwardMessage(token, messageId, forwardTarget))}
          >
            Forward
          </button>
        </>
      )}
      {error && <span className="error">{error}</span>}
    </div>
  );
}
