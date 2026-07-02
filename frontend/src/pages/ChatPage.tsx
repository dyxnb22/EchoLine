import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { getOrCreateDeviceId } from "../lib/deviceId";
import {
  addReaction,
  adminListUsers,
  blockUser,
  connectWS,
  Conversation,
  createPaymentLedger,
  editMessage,
  fetchMe,
  listArchived,
  listConversations,
  listMessages,
  listNotifications,
  listPins,
  listReactions,
  listFriendRecommendations,
  listRecommendations,
  markConversationRead,
  Message,
  Notification,
  presignDownload,
  presignUpload,
  Reaction,
  recallMessage,
  refreshTokenOnce,
  removeReaction,
  reportMessage,
  searchMessages,
  SearchHit,
  sendMessage,
  settlePaymentLedger,
  subscribeChannel,
  syncConversations,
  touchLastSeen,
  unarchiveConversation,
  WSStatus,
} from "../api";
import { CreateConversationModal } from "../components/CreateConversationModal";
import { AdminPanel } from "../components/AdminPanel";
import { ConversationActions } from "../components/ConversationActions";
import { GroupSettings } from "../components/GroupSettings";
import { NotificationPanel } from "../components/NotificationPanel";
import { ThreadPanel } from "../components/ThreadPanel";
import { isPaymentRequired } from "../api/http";
import { useAuth } from "../context/AuthContext";

export function ChatPage() {
  const { token, logout } = useAuth();
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
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
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  const [filter, setFilter] = useState<"all" | "channel" | "group">("all");
  const [dark, setDark] = useState(localStorage.getItem("echoline_dark") === "1");
  const [recs, setRecs] = useState<{ channel_id: string; title: string }[]>([]);
  const [friends, setFriends] = useState<{ id: string; username: string; display_name: string }[]>([]);
  const [threadMsg, setThreadMsg] = useState<Message | null>(null);
  const [showAdmin, setShowAdmin] = useState(false);
  const [showNotifs, setShowNotifs] = useState(false);
  const [showArchived, setShowArchived] = useState(false);
  const [archived, setArchived] = useState<Conversation[]>([]);
  const [pins, setPins] = useState<{ message_id: string }[]>([]);
  const [reactions, setReactions] = useState<Record<string, Reaction[]>>({});
  const activeIdRef = useRef<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const wsRef = useRef<{ close: () => void; send: (p: unknown) => void } | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const seqCursorsRef = useRef<Record<string, number>>({});
  const conversationsRef = useRef<Conversation[]>([]);
  const pendingSearchSeqRef = useRef<{ conversationId: string; seq: number } | null>(null);
  const pendingSeqRef = useRef(Number.MAX_SAFE_INTEGER);
  const conversationsLoadedRef = useRef(false);
  const typingTimer = useRef<number | undefined>(undefined);

  const allocPendingSeq = useCallback(() => {
    pendingSeqRef.current -= 1;
    return pendingSeqRef.current;
  }, []);

  const deviceId = useMemo(() => getOrCreateDeviceId(), []);

  useEffect(() => {
    activeIdRef.current = activeId;
  }, [activeId]);

  useEffect(() => {
    document.documentElement.classList.toggle("dark", dark);
    localStorage.setItem("echoline_dark", dark ? "1" : "0");
  }, [dark]);

  useEffect(() => {
    conversationsRef.current = conversations;
  }, [conversations]);

  useEffect(() => {
    if (!token) {
      setCurrentUserId(null);
      return;
    }
    fetchMe(token).then((u) => setCurrentUserId(u.id)).catch(() => undefined);
  }, [token]);

  const refreshConversations = useCallback(() => {
    if (!token) return;
    const active = activeIdRef.current;
    listConversations(token).then((list) => {
      setConversations(() => {
        if (!active) return list;
        return list.map((c) => (c.id === active ? { ...c, unread: 0 } : c));
      });
    }).catch((e) => setError(String(e)));
  }, [token]);

  const clearUnread = useCallback((conversationId: string) => {
    setConversations((prev) => prev.map((c) => (
      c.id === conversationId ? { ...c, unread: 0 } : c
    )));
  }, []);

  const markActiveRead = useCallback((conversationId: string, seq: number) => {
    if (!token) return;
    clearUnread(conversationId);
    markConversationRead(token, conversationId, seq).catch(() => undefined);
  }, [token, clearUnread]);

  const activeConv = conversations.find((c) => c.id === activeId);
  const canPublish = activeConv
    ? (activeConv.can_publish ?? activeConv.type !== "channel")
    : false;

  const mergeSyncedMessages = useCallback((blocks: { conversation_id: string; messages?: Message[] }[]) => {
    for (const block of blocks) {
      if (!block.messages?.length) continue;
      const maxSeq = Math.max(...block.messages.map((m) => m.seq));
      seqCursorsRef.current[block.conversation_id] = maxSeq;
      if (block.conversation_id === activeIdRef.current) {
        setMessages((prev) => {
          const merged = [...prev];
          for (const m of block.messages!) {
            const dupIdx = merged.findIndex(
              (x) => x.seq === m.seq || (m.client_msg_id && x.client_msg_id === m.client_msg_id),
            );
            if (dupIdx >= 0) {
              merged[dupIdx] = { ...merged[dupIdx], ...m, pending: false, failed: false };
            } else {
              merged.push(m);
            }
          }
          return merged.sort((a, b) => a.seq - b.seq);
        });
      }
    }
  }, []);

  const runSync = useCallback(async () => {
    if (!token) return;
    const cursorMap = new Map<string, number>(
      Object.entries(seqCursorsRef.current).map(([id, seq]) => [id, seq]),
    );
    for (const c of conversationsRef.current) {
      if (!cursorMap.has(c.id)) {
        cursorMap.set(c.id, Math.max(0, c.latest_seq - 1));
      }
    }
    const cursors = [...cursorMap.entries()].map(([conversation_id, last_seq]) => ({
      conversation_id,
      last_seq,
    }));
    if (cursors.length === 0) return;
    try {
      const allBlocks: { conversation_id: string; messages?: Message[] }[] = [];
      for (const cursor of cursors) {
        let lastSeq = cursor.last_seq;
        let hasMore = true;
        while (hasMore) {
          const synced = await syncConversations(token, deviceId, [{
            conversation_id: cursor.conversation_id,
            last_seq: lastSeq,
          }]);
          const block = synced[0];
          if (!block) break;
          if (block.messages?.length) {
            allBlocks.push(block);
            lastSeq = Math.max(...block.messages.map((m) => m.seq));
          }
          hasMore = block.has_more ?? false;
          if (!block.messages?.length && !hasMore) break;
        }
        seqCursorsRef.current[cursor.conversation_id] = lastSeq;
      }
      mergeSyncedMessages(allBlocks);
      for (const block of allBlocks) {
        if (block.conversation_id === activeIdRef.current && block.messages?.length) {
          const maxSeq = Math.max(...block.messages.map((m) => m.seq));
          markActiveRead(block.conversation_id, maxSeq);
          break;
        }
      }
      refreshConversations();
    } catch {
      // sync is best-effort on reconnect
    }
  }, [token, deviceId, refreshConversations, mergeSyncedMessages, markActiveRead]);

  const runSyncRef = useRef(runSync);
  useEffect(() => {
    runSyncRef.current = runSync;
  }, [runSync]);

  useEffect(() => {
    if (conversations.length === 0) return;
    const justLoaded = !conversationsLoadedRef.current;
    conversationsLoadedRef.current = true;
    if (justLoaded && wsStatus === "open") {
      void runSyncRef.current();
    }
  }, [conversations.length, wsStatus]);

  const refreshAccessToken = useCallback(async () => {
    const refresh = localStorage.getItem("echoline_refresh");
    if (!refresh) return null;
    try {
      const pair = await refreshTokenOnce(refresh);
      localStorage.setItem("echoline_token", pair.access_token);
      localStorage.setItem("echoline_refresh", pair.refresh_token);
      return pair.access_token;
    } catch {
      logout();
      return null;
    }
  }, [logout]);

  const loggedIn = !!token;
  const prevWsStatusRef = useRef(wsStatus);

  useEffect(() => {
    if (prevWsStatusRef.current === "open" && wsStatus !== "open") {
      void runSyncRef.current();
    }
    prevWsStatusRef.current = wsStatus;
  }, [wsStatus]);

  useEffect(() => {
    if (!loggedIn || wsStatus === "open") return;
    const timer = window.setInterval(() => { void runSyncRef.current(); }, 30_000);
    return () => window.clearInterval(timer);
  }, [loggedIn, wsStatus]);

  useEffect(() => {
    if (!token) return;
    listRecommendations(token).then(setRecs).catch(() => undefined);
    listFriendRecommendations(token).then(setFriends).catch(() => undefined);
    touchLastSeen(token).catch(() => undefined);
    refreshConversations();
  }, [token, refreshConversations]);

  const prefetchReactions = useCallback(async (msgs: Message[]) => {
    if (!token) return;
    const recent = msgs.slice(-20);
    const results = await Promise.all(
      recent.map(async (m) => {
        try {
          const rx = await listReactions(token, m.id);
          return [m.id, rx] as const;
        } catch {
          return [m.id, []] as const;
        }
      }),
    );
    setReactions((prev) => {
      const next = { ...prev };
      for (const [id, rx] of results) next[id] = [...rx];
      return next;
    });
  }, [token]);

  const loadMessages = useCallback(async (conversationId: string, beforeSeq?: number) => {
    if (!token) return [];
    const page = await listMessages(token, conversationId, beforeSeq);
    const ordered = [...page.messages].reverse();
    if (beforeSeq == null) {
      setMessages(ordered);
      void prefetchReactions(ordered);
    } else {
      setMessages((prev) => [...ordered, ...prev]);
      void prefetchReactions(ordered);
    }
    setNextBefore(page.next_before);
    if (beforeSeq == null && ordered.length > 0) {
      const maxSeq = Math.max(...ordered.map((m) => m.seq));
      const prev = seqCursorsRef.current[conversationId] ?? 0;
      seqCursorsRef.current[conversationId] = Math.max(prev, maxSeq);
    }
    return ordered;
  }, [token, prefetchReactions]);

  useEffect(() => {
    if (!token || !activeId) {
      if (!activeId) setMessages([]);
      return;
    }
    const pending = pendingSearchSeqRef.current;
    if (pending?.conversationId === activeId) {
      pendingSearchSeqRef.current = null;
      setNextBefore(null);
      setTypingUsers([]);
      void loadMessages(activeId, pending.seq + 1).then((ordered) => {
        const target = ordered.find((m) => m.seq === pending.seq);
        if (target) {
          setMessages((prev) => {
            if (prev.some((m) => m.seq === pending.seq)) return prev;
            return [...prev, target].sort((a, b) => a.seq - b.seq);
          });
        }
        const last = ordered[ordered.length - 1];
        if (last?.seq) {
          markActiveRead(activeId, last.seq);
        }
      }).catch((e) => setError(String(e)));
      listPins(token, activeId).then(setPins).catch(() => setPins([]));
      return;
    }
    setNextBefore(null);
    setTypingUsers([]);
    loadMessages(activeId).then((ordered) => {
      const last = ordered[ordered.length - 1];
      if (last?.seq) {
        markActiveRead(activeId, last.seq);
      }
    }).catch((e) => setError(String(e)));
    listPins(token, activeId).then(setPins).catch(() => setPins([]));
  }, [token, activeId, loadMessages, markActiveRead]);

  useEffect(() => {
    if (!token) return;
    listNotifications(token).then(setNotifications).catch(() => undefined);
  }, [token, activeId]);

  useEffect(() => {
    if (!token || !showArchived) return;
    listArchived(token).then(setArchived).catch(() => setArchived([]));
  }, [token, showArchived]);

  useEffect(() => {
    if (!loggedIn) return;
    const conn = connectWS(
      () => localStorage.getItem("echoline_token") ?? "",
      deviceId,
      (payload) => {
      const env = payload as {
        type?: string;
        payload?: {
          conversation_id?: string;
          seq?: number;
          body?: string;
          id?: string;
          message_id?: string;
          sender_id?: string;
          user_id?: string;
          client_msg_id?: string;
          status?: string;
          attachment?: { object_key?: string; mime_type?: string };
        };
      };
      if (env.type === "typing.indicator" && env.payload?.conversation_id === activeIdRef.current) {
        const uid = env.payload.user_id ?? "someone";
        setTypingUsers((prev) => (prev.includes(uid) ? prev : [...prev, uid]));
        window.setTimeout(() => {
          setTypingUsers((prev) => prev.filter((u) => u !== uid));
        }, 3000);
        return;
      }
      if (env.type === "typing.stopped" && env.payload?.conversation_id === activeIdRef.current) {
        const uid = env.payload.user_id;
        if (uid) setTypingUsers((prev) => prev.filter((u) => u !== uid));
        return;
      }
      if (env.type === "message.edited" && env.payload?.conversation_id === activeIdRef.current) {
        const id = env.payload.id ?? env.payload.message_id;
        if (!id) return;
        setMessages((prev) => prev.map((m) => (
          m.id === id ? { ...m, body: env.payload!.body ?? m.body } : m
        )));
        return;
      }
      if (env.type === "message.recalled" && env.payload?.conversation_id === activeIdRef.current) {
        const id = env.payload.id ?? env.payload.message_id;
        if (!id) return;
        setMessages((prev) => prev.map((m) => (
          m.id === id ? { ...m, body: "", status: "recalled" } : m
        )));
        return;
      }
      if (env.type !== "message.created") return;
      const convId = env.payload?.conversation_id;
      if (convId && env.payload?.seq) {
        const prev = seqCursorsRef.current[convId] ?? 0;
        seqCursorsRef.current[convId] = Math.max(prev, env.payload.seq);
      }
      const isActive = convId === activeIdRef.current;
      if (isActive && env.payload?.seq) {
        markActiveRead(convId!, env.payload.seq);
      } else {
        refreshConversations();
      }
      if (isActive && env.payload?.id && wsRef.current) {
        wsRef.current.send({
          type: "message.ack",
          payload: {
            conversation_id: convId,
            message_id: env.payload.id,
            seq: env.payload.seq,
            status: "delivered",
          },
        });
      }
      if (convId !== activeIdRef.current) return;
      const wsAttachment = env.payload!.attachment?.object_key
        ? {
          object_key: env.payload!.attachment.object_key,
          mime_type: env.payload!.attachment.mime_type,
        }
        : undefined;
      setMessages((prev) => {
        const seq = env.payload!.seq ?? 0;
        const clientMsgId = env.payload!.client_msg_id;
        const withoutPending = prev.filter((m) => !(
          m.pending && (
            (clientMsgId && m.client_msg_id === clientMsgId)
            || (!clientMsgId && m.body === (env.payload!.body ?? ""))
          )
        ));
        const dupIdx = withoutPending.findIndex(
          (m) => m.seq === seq || (clientMsgId && m.client_msg_id === clientMsgId),
        );
        if (dupIdx >= 0) {
          const updated = [...withoutPending];
          updated[dupIdx] = {
            ...updated[dupIdx],
            id: env.payload!.id ?? updated[dupIdx].id,
            seq,
            body: env.payload!.body ?? "",
            sender_id: env.payload!.sender_id ?? "",
            attachment: wsAttachment ?? updated[dupIdx].attachment,
            pending: false,
          };
          return updated;
        }
        return [...withoutPending, {
          id: env.payload!.id ?? crypto.randomUUID(),
          seq,
          body: env.payload!.body ?? "",
          sender_id: env.payload!.sender_id ?? "",
          client_msg_id: clientMsgId,
          attachment: wsAttachment,
        }];
      });
    }, setWsStatus, () => { void runSyncRef.current(); }, refreshAccessToken);
    wsRef.current = conn;
    return () => conn.close();
  }, [loggedIn, deviceId, refreshConversations, refreshAccessToken, markActiveRead]);

  function emitTyping() {
    if (!activeId || !wsRef.current) return;
    wsRef.current.send({
      type: "typing.start",
      payload: { conversation_id: activeId },
    });
    if (typingTimer.current) window.clearTimeout(typingTimer.current);
    typingTimer.current = window.setTimeout(() => {
      wsRef.current?.send({
        type: "typing.stop",
        payload: { conversation_id: activeId },
      });
    }, 2000);
  }

  async function loadReactions(messageId: string) {
    if (!token) return;
    const rx = await listReactions(token, messageId);
    setReactions((prev) => ({ ...prev, [messageId]: rx }));
  }

  async function handleRecommendedChannel(channelId: string) {
    if (!token) return;
    try {
      await subscribeChannel(token, channelId);
      refreshConversations();
      setActiveId(channelId);
    } catch (err) {
      if (isPaymentRequired(err)) {
        try {
          const reference = `channel:${channelId}`;
          await createPaymentLedger(token, 999, reference);
          await settlePaymentLedger(token, reference);
          await subscribeChannel(token, channelId);
          refreshConversations();
          setActiveId(channelId);
          return;
        } catch (payErr) {
          setError(String(payErr));
          return;
        }
      }
      setError(String(err));
    }
  }

  async function toggleAdmin() {
    if (!token || showAdmin) {
      setShowAdmin(false);
      return;
    }
    try {
      await adminListUsers(token);
      setShowAdmin(true);
    } catch {
      setError("Admin access denied");
    }
  }

  async function handleSend(e: React.FormEvent) {
    e.preventDefault();
    if (!token || !activeId || !draft.trim()) return;
    const body = draft.trim();
    const tempId = crypto.randomUUID();
    const clientMsgId = crypto.randomUUID();
    const optimistic: Message = {
      id: tempId,
      seq: allocPendingSeq(),
      body,
      sender_id: "me",
      pending: true,
      client_msg_id: clientMsgId,
    };
    setMessages((prev) => [...prev, optimistic]);
    setDraft("");
    try {
      const created = await sendMessage(token, activeId, body, undefined, clientMsgId);
      setMessages((prev) => prev.map((m) => (
        m.id === tempId
          ? { ...m, ...created, pending: false, failed: false, sender_id: created.sender_id }
          : m
      )));
    } catch (err) {
      setMessages((prev) => prev.map((m) => (m.id === tempId ? { ...m, pending: false, failed: true } : m)));
      setError(String(err));
    }
  }

  async function handleUpload(file: File) {
    if (!token || !activeId) return;
    setUploading(true);
    setError(null);
    const tempId = crypto.randomUUID();
    const clientMsgId = crypto.randomUUID();
    try {
      const { object_key } = await presignUpload(token, file);
      setMessages((prev) => [...prev, {
        id: tempId,
        seq: allocPendingSeq(),
        body: file.name,
        sender_id: "me",
        pending: true,
        client_msg_id: clientMsgId,
        attachment: { object_key, mime_type: file.type },
      }]);
      const created = await sendMessage(token, activeId, file.name, object_key, clientMsgId);
      setMessages((prev) => prev.map((m) => (
        m.id === tempId
          ? { ...m, ...created, pending: false, failed: false, sender_id: created.sender_id }
          : m
      )));
    } catch (err) {
      setMessages((prev) => prev.map((m) => (m.id === tempId ? { ...m, pending: false, failed: true } : m)));
      setError(String(err));
    } finally {
      setUploading(false);
    }
  }

  async function handleAttachmentDownload(objectKey: string) {
    if (!token) return;
    try {
      const { download_url } = await presignDownload(token, objectKey);
      window.open(download_url, "_blank", "noopener,noreferrer");
    } catch (err) {
      setError(String(err));
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

  async function handleEditMessage(m: Message) {
    if (!token || !activeId) return;
    const next = window.prompt("Edit message", m.body);
    if (next == null || next === m.body) return;
    try {
      await editMessage(token, activeId, m.id, next);
      setMessages((prev) => prev.map((x) => (x.id === m.id ? { ...x, body: next } : x)));
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleRecallMessage(m: Message) {
    if (!token || !activeId) return;
    if (!window.confirm("Recall this message?")) return;
    try {
      await recallMessage(token, activeId, m.id);
      setMessages((prev) => prev.map((x) => (
        x.id === m.id ? { ...x, body: "", status: "recalled" } : x
      )));
    } catch (err) {
      setError(String(err));
    }
  }

  if (!token) return null;

  const active = conversations.find((c) => c.id === activeId);
  const filtered = conversations.filter((c) => filter === "all" || c.type === filter);

  return (
    <main className="layout">
      <aside>
        <header>
          <h1>EchoLine</h1>
          <button type="button" className="theme-toggle" onClick={() => setDark((d) => !d)}>{dark ? "☀" : "☾"}</button>
          <span className={`ws-status ws-${wsStatus}`}>{wsStatus}</span>
          <button type="button" className="notif-btn" onClick={() => setShowNotifs((v) => !v)}>
            🔔{notifications.filter((n) => !n.read_at).length > 0
              ? ` (${notifications.filter((n) => !n.read_at).length})`
              : ""}
          </button>
          <Link to="/settings">Settings</Link>
          <button type="button" onClick={logout}>Logout</button>
          <button type="button" onClick={() => setShowCreate(true)}>New chat</button>
          <button type="button" onClick={() => void toggleAdmin()}>Admin</button>
        </header>
        <div className="filter-tabs">
          {(["all", "channel", "group"] as const).map((f) => (
            <button key={f} type="button" className={filter === f ? "active" : ""} onClick={() => setFilter(f)}>{f}</button>
          ))}
          <button type="button" className={showArchived ? "active" : ""} onClick={() => setShowArchived((v) => !v)}>archived</button>
        </div>
        <form onSubmit={handleSearch} className="search">
          <input value={searchQ} onChange={(e) => setSearchQ(e.target.value)} placeholder="Search messages" />
          <button type="submit">Search</button>
        </form>
        {searchHits.length > 0 && (
          <ul className="search-results">
            {searchHits.map((h) => (
              <li key={h.message_id}>
                <button type="button" onClick={() => {
                  pendingSearchSeqRef.current = { conversationId: h.conversation_id, seq: h.seq };
                  setActiveId(h.conversation_id);
                }}>#{h.seq} {h.body}</button>
              </li>
            ))}
          </ul>
        )}
        {friends.length > 0 && (
          <div className="recs">
            <strong>Friends</strong>
            {friends.map((f) => (
              <span key={f.id} className="friend-chip">{f.display_name || f.username}</span>
            ))}
          </div>
        )}
        {recs.length > 0 && (
          <div className="recs">
            <strong>Recommended</strong>
            {recs.map((r) => (
              <button key={r.channel_id} type="button" onClick={() => void handleRecommendedChannel(r.channel_id)}>{r.title}</button>
            ))}
          </div>
        )}
        {showArchived && archived.length > 0 && (
          <ul className="archived-list">
            {archived.map((c) => (
              <li key={c.id}>
                <button
                  type="button"
                  onClick={() => {
                    if (!token) return;
                    unarchiveConversation(token, c.id)
                      .then(() => {
                        refreshConversations();
                        setActiveId(c.id);
                        setShowArchived(false);
                      })
                      .catch((e) => setError(String(e)));
                  }}
                >
                  {c.title || c.type} (archived)
                </button>
              </li>
            ))}
          </ul>
        )}
        <ul>
          {filtered.map((c) => (
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
          {token && activeId && active && (
            <ConversationActions
              token={token}
              conversationId={activeId}
              conversationType={active.type}
              messageId={messages[messages.length - 1]?.id}
              onAction={refreshConversations}
            />
          )}
        </header>
        {active?.type === "group" && token && activeId && (
          <GroupSettings token={token} conversationId={activeId} />
        )}
        {pins.length > 0 && (
          <div className="pins-bar">
            <strong>Pinned:</strong> {pins.map((p) => p.message_id.slice(0, 8)).join(", ")}
          </div>
        )}
        {nextBefore != null && (
          <button className="load-more" onClick={handleLoadMore} disabled={loadingMore}>
            {loadingMore ? "Loading..." : "Load older messages"}
          </button>
        )}
        <div className="messages">
          {messages.map((m) => (
            <div key={`${m.id}-${m.seq}`} className={`message ${m.pending ? "pending" : ""} ${m.failed ? "failed" : ""} ${m.status === "recalled" ? "recalled" : ""}`}>
              <strong>#{m.seq}</strong>{" "}
              {m.status === "recalled" ? (
                <em>(recalled)</em>
              ) : m.attachment?.object_key ? (
                <>
                  {m.body && <span>{m.body} </span>}
                  <button
                    type="button"
                    className="attachment-link"
                    onClick={() => void handleAttachmentDownload(m.attachment!.object_key)}
                  >
                    📎 Download{m.attachment.mime_type ? ` (${m.attachment.mime_type})` : ""}
                  </button>
                </>
              ) : (
                m.body
              )}
              {m.pending && <em> sending...</em>}
              {m.failed && <em> failed</em>}
              {(reactions[m.id] ?? []).length > 0 && (
                <span className="reactions">
                  {(reactions[m.id] ?? []).map((rx, i) => (
                    <button key={`${rx.emoji}-${i}`} type="button" onClick={() => token && removeReaction(token, m.id, rx.emoji).then(() => loadReactions(m.id)).catch((e) => setError(String(e)))}>
                      {rx.emoji}
                    </button>
                  ))}
                </span>
              )}
              {token && activeId && (
                <span className="msg-actions">
                  <button type="button" onClick={() => addReaction(token, m.id, "👍").then(() => loadReactions(m.id)).catch((e) => setError(String(e)))}>👍</button>
                  <button type="button" onClick={() => addReaction(token, m.id, "❤️").then(() => loadReactions(m.id)).catch((e) => setError(String(e)))}>❤️</button>
                  <button type="button" onClick={() => setThreadMsg(m)}>Reply</button>
                  <button type="button" onClick={() => void handleEditMessage(m)}>Edit</button>
                  <button type="button" onClick={() => void handleRecallMessage(m)}>Recall</button>
                  <button type="button" onClick={() => reportMessage(token, activeId, m.id, "spam").catch((e) => setError(String(e)))}>Report</button>
                  {currentUserId && m.sender_id !== currentUserId && (
                    <button type="button" onClick={() => blockUser(token, m.sender_id).catch((e) => setError(String(e)))}>Block</button>
                  )}
                </span>
              )}
            </div>
          ))}
        </div>
        {!canPublish && activeId && (
          <p className="channel-readonly">Subscribe-only channel — you cannot post messages here.</p>
        )}
        {activeId && canPublish && (
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
      {threadMsg && token && activeId && (
        <ThreadPanel token={token} convId={activeId} parentMessage={threadMsg} onClose={() => setThreadMsg(null)} />
      )}
      {showNotifs && token && (
        <NotificationPanel token={token} onClose={() => setShowNotifs(false)} />
      )}
      {showAdmin && token && (
        <AdminPanel token={token} onClose={() => setShowAdmin(false)} />
      )}
      {showCreate && token && (
        <CreateConversationModal
          token={token}
          onCreated={(id) => {
            refreshConversations();
            setActiveId(id);
          }}
          onClose={() => setShowCreate(false)}
        />
      )}
    </main>
  );
}
