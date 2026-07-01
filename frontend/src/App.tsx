import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  connectWS,
  Conversation,
  listConversations,
  listMessages,
  listNotifications,
  login,
  markConversationRead,
  Message,
  Notification,
  presignUpload,
  register,
  searchMessages,
  SearchHit,
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
  const [searchQ, setSearchQ] = useState("");
  const [searchHits, setSearchHits] = useState<SearchHit[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [wsStatus, setWsStatus] = useState<WSStatus>("closed");
  const [uploading, setUploading] = useState(false);
  const [authMode, setAuthMode] = useState<"login" | "register">("login");
  const [displayName, setDisplayName] = useState("");
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  const activeIdRef = useRef<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const wsRef = useRef<{ close: () => void; send: (p: unknown) => void } | null>(null);
  const typingTimer = useRef<number | undefined>(undefined);

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
    if (!token) return [];
    const page = await listMessages(token, conversationId, beforeSeq);
    const ordered = [...page.messages].reverse();
    if (beforeSeq == null) {
      setMessages(ordered);
    } else {
      setMessages((prev) => [...ordered, ...prev]);
    }
    setNextBefore(page.next_before);
    return ordered;
  }, [token]);

  useEffect(() => {
    if (!token || !activeId) return;
    setNextBefore(null);
    setTypingUsers([]);
    loadMessages(activeId).then((ordered) => {
      const last = ordered[ordered.length - 1];
      if (last?.seq) {
        markConversationRead(token, activeId, last.seq).catch(() => undefined);
      }
    }).catch((e) => setError(String(e)));
  }, [token, activeId, loadMessages]);

  useEffect(() => {
    if (!token) return;
    listNotifications(token).then(setNotifications).catch(() => undefined);
  }, [token, messages.length]);

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
          user_id?: string;
        };
      };
      if (env.type === "typing.start" && env.payload?.conversation_id === activeIdRef.current) {
        const uid = env.payload.user_id ?? "someone";
        setTypingUsers((prev) => (prev.includes(uid) ? prev : [...prev, uid]));
        window.setTimeout(() => {
          setTypingUsers((prev) => prev.filter((u) => u !== uid));
        }, 3000);
        return;
      }
      if (env.type !== "message.created") return;
      if (env.payload?.conversation_id !== activeIdRef.current) return;
      setMessages((prev) => {
        const seq = env.payload!.seq ?? 0;
        const withoutPending = prev.filter((m) => !(m.pending && m.body === (env.payload!.body ?? "")));
        if (withoutPending.some((m) => m.seq === seq)) return withoutPending;
        return [...withoutPending, {
          id: env.payload!.id ?? crypto.randomUUID(),
          seq,
          body: env.payload!.body ?? "",
          sender_id: env.payload!.sender_id ?? "",
        }];
      });
    }, setWsStatus);
    wsRef.current = conn;
    return () => conn.close();
  }, [token, deviceId]);

  function emitTyping() {
    if (!activeId || !wsRef.current) return;
    wsRef.current.send({
      type: "typing.start",
      payload: { conversation_id: activeId },
    });
  }

  async function handleAuth(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      if (authMode === "register") {
        await register(username, password, displayName || username);
      }
      const tokens = await login(username, password);
      localStorage.setItem("echoline_token", tokens.access_token);
      localStorage.setItem("echoline_refresh", tokens.refresh_token);
      setToken(tokens.access_token);
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleSend(e: React.FormEvent) {
    e.preventDefault();
    if (!token || !activeId || !draft.trim()) return;
    const body = draft.trim();
    const tempId = crypto.randomUUID();
    const optimistic: Message = {
      id: tempId,
      seq: Date.now(),
      body,
      sender_id: "me",
      pending: true,
    };
    setMessages((prev) => [...prev, optimistic]);
    setDraft("");
    try {
      await sendMessage(token, activeId, body);
    } catch (err) {
      setMessages((prev) => prev.map((m) => (m.id === tempId ? { ...m, pending: false, failed: true } : m)));
      setError(String(err));
    }
  }

  async function handleUpload(file: File) {
    if (!token || !activeId) return;
    setUploading(true);
    setError(null);
    try {
      const { object_key } = await presignUpload(token, file);
      const tempId = crypto.randomUUID();
      setMessages((prev) => [...prev, {
        id: tempId,
        seq: Date.now(),
        body: file.name,
        sender_id: "me",
        pending: true,
        attachment: { object_key, mime_type: file.type },
      }]);
      await sendMessage(token, activeId, file.name, object_key);
    } catch (err) {
      setError(String(err));
    } finally {
      setUploading(false);
    }
  }

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (!token || !searchQ.trim()) return;
    try {
      const hits = await searchMessages(token, searchQ.trim());
      setSearchHits(hits);
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
        <form onSubmit={handleAuth} className="card">
          <div className="auth-tabs">
            <button type="button" className={authMode === "login" ? "active" : ""} onClick={() => setAuthMode("login")}>Login</button>
            <button type="button" className={authMode === "register" ? "active" : ""} onClick={() => setAuthMode("register")}>Register</button>
          </div>
          <input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="username" />
          {authMode === "register" && (
            <input value={displayName} onChange={(e) => setDisplayName(e.target.value)} placeholder="display name" />
          )}
          <input value={password} onChange={(e) => setPassword(e.target.value)} type="password" placeholder="password" />
          <button type="submit">{authMode === "login" ? "Login" : "Register & Login"}</button>
          {error && <p className="error toast">{error}</p>}
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
          {notifications.length > 0 && <span className="notif-badge">{notifications.length}</span>}
          <button onClick={() => { localStorage.removeItem("echoline_token"); setToken(null); }}>Logout</button>
        </header>
        <form onSubmit={handleSearch} className="search">
          <input value={searchQ} onChange={(e) => setSearchQ(e.target.value)} placeholder="Search messages" />
          <button type="submit">Search</button>
        </form>
        {searchHits.length > 0 && (
          <ul className="search-results">
            {searchHits.map((h) => (
              <li key={h.message_id}>
                <button onClick={() => setActiveId(h.conversation_id)}>#{h.seq} {h.body}</button>
              </li>
            ))}
          </ul>
        )}
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
        <header>
          {active ? (active.title || active.type) : "Select a conversation"}
          {typingUsers.length > 0 && <span className="typing">typing...</span>}
        </header>
        {nextBefore != null && (
          <button className="load-more" onClick={handleLoadMore} disabled={loadingMore}>
            {loadingMore ? "Loading..." : "Load older messages"}
          </button>
        )}
        <div className="messages">
          {messages.map((m) => (
            <div key={`${m.id}-${m.seq}`} className={`message ${m.pending ? "pending" : ""} ${m.failed ? "failed" : ""}`}>
              <strong>#{m.seq}</strong> {m.body}
              {m.pending && <em> sending...</em>}
              {m.failed && <em> failed</em>}
            </div>
          ))}
        </div>
        {activeId && (
          <form onSubmit={handleSend} className="composer">
            <input
              value={draft}
              onChange={(e) => {
                setDraft(e.target.value);
                if (typingTimer.current) window.clearTimeout(typingTimer.current);
                typingTimer.current = window.setTimeout(emitTyping, 300);
              }}
              placeholder="Type a message"
            />
            <input
              ref={fileInputRef}
              type="file"
              hidden
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) void handleUpload(file);
                e.target.value = "";
              }}
            />
            <button type="button" disabled={uploading} onClick={() => fileInputRef.current?.click()}>
              {uploading ? "..." : "Attach"}
            </button>
            <button type="submit">Send</button>
          </form>
        )}
        {error && <p className="error toast">{error}</p>}
      </section>
    </main>
  );
}
