import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  connectWS,
  Conversation,
  listConversations,
  listMessages,
  login,
  Message,
  sendMessage,
  WSStatus,
} from "./api";

export default function App() {
  const [username, setUsername] = useState("alice");
  const [password, setPassword] = useState("secret123");
  const [token, setToken] = useState<string | null>(localStorage.getItem("echoline_token"));
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [nextBefore, setNextBefore] = useState<number | null>(null);
  const [loadingMore, setLoadingMore] = useState(false);
  const [draft, setDraft] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [wsStatus, setWsStatus] = useState<WSStatus>("closed");
  const activeIdRef = useRef<string | null>(null);

  const deviceId = useMemo(() => localStorage.getItem("echoline_device") ?? crypto.randomUUID(), []);

  useEffect(() => {
    localStorage.setItem("echoline_device", deviceId);
  }, [deviceId]);

  useEffect(() => {
    activeIdRef.current = activeId;
  }, [activeId]);

  useEffect(() => {
    if (!token) return;
    listConversations(token)
      .then(setConversations)
      .catch((e) => setError(String(e)));
  }, [token]);

  const loadMessages = useCallback(async (conversationId: string, beforeSeq?: number) => {
    if (!token) return;
    const page = await listMessages(token, conversationId, beforeSeq);
    const ordered = [...page.messages].reverse();
    if (beforeSeq == null) {
      setMessages(ordered);
    } else {
      setMessages((prev) => [...ordered, ...prev]);
    }
    setNextBefore(page.next_before);
  }, [token]);

  useEffect(() => {
    if (!token || !activeId) return;
    setNextBefore(null);
    loadMessages(activeId).catch((e) => setError(String(e)));
  }, [token, activeId, loadMessages]);

  useEffect(() => {
    if (!token) return;
    const conn = connectWS(token, deviceId, (payload) => {
      const env = payload as {
        type?: string;
        payload?: {
          conversation_id?: string;
          seq?: number;
          body?: string;
          id?: string;
          sender_id?: string;
        };
      };
      if (env.type !== "message.created") return;
      if (env.payload?.conversation_id !== activeIdRef.current) return;
      setMessages((prev) => {
        const seq = env.payload!.seq ?? 0;
        if (prev.some((m) => m.seq === seq)) return prev;
        return [...prev, {
          id: env.payload!.id ?? crypto.randomUUID(),
          seq,
          body: env.payload!.body ?? "",
          sender_id: env.payload!.sender_id ?? "",
        }];
      });
    }, setWsStatus);
    return () => conn.close();
  }, [token, deviceId]);

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      const tokens = await login(username, password);
      localStorage.setItem("echoline_token", tokens.access_token);
      setToken(tokens.access_token);
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleSend(e: React.FormEvent) {
    e.preventDefault();
    if (!token || !activeId || !draft.trim()) return;
    try {
      await sendMessage(token, activeId, draft.trim());
      setDraft("");
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleLoadMore() {
    if (!activeId || nextBefore == null || loadingMore) return;
    setLoadingMore(true);
    try {
      await loadMessages(activeId, nextBefore);
    } catch (err) {
      setError(String(err));
    } finally {
      setLoadingMore(false);
    }
  }

  if (!token) {
    return (
      <main className="shell">
        <h1>EchoLine</h1>
        <form onSubmit={handleLogin} className="card">
          <input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="username" />
          <input value={password} onChange={(e) => setPassword(e.target.value)} type="password" placeholder="password" />
          <button type="submit">Login</button>
          {error && <p className="error">{error}</p>}
        </form>
      </main>
    );
  }

  const active = conversations.find((c) => c.id === activeId);

  return (
    <main className="layout">
      <aside>
        <header>
          <h1>EchoLine</h1>
          <span className={`ws-status ws-${wsStatus}`}>{wsStatus}</span>
          <button onClick={() => { localStorage.removeItem("echoline_token"); setToken(null); }}>Logout</button>
        </header>
        <ul>
          {conversations.map((c) => (
            <li key={c.id}>
              <button className={c.id === activeId ? "active" : ""} onClick={() => setActiveId(c.id)}>
                {c.title || c.type} {c.unread ? `(${c.unread})` : ""}
              </button>
            </li>
          ))}
        </ul>
      </aside>
      <section className="chat">
        <header>{active ? (active.title || active.type) : "Select a conversation"}</header>
        {nextBefore != null && (
          <button className="load-more" onClick={handleLoadMore} disabled={loadingMore}>
            {loadingMore ? "Loading..." : "Load older messages"}
          </button>
        )}
        <div className="messages">
          {messages.map((m) => (
            <div key={`${m.id}-${m.seq}`} className="message">
              <strong>#{m.seq}</strong> {m.body}
            </div>
          ))}
        </div>
        {activeId && (
          <form onSubmit={handleSend} className="composer">
            <input value={draft} onChange={(e) => setDraft(e.target.value)} placeholder="Type a message" />
            <button type="submit">Send</button>
          </form>
        )}
        {error && <p className="error">{error}</p>}
      </section>
    </main>
  );
}
