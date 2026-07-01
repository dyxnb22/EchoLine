const API_BASE = import.meta.env.VITE_API_BASE ?? "";

export type TokenPair = {
  access_token: string;
  refresh_token: string;
};

export type Conversation = {
  id: string;
  type: string;
  title: string;
  unread: number;
  latest_seq: number;
};

export type Message = {
  id: string;
  body: string;
  seq: number;
  sender_id: string;
  type?: string;
  pending?: boolean;
  failed?: boolean;
  attachment?: { object_key: string; mime_type?: string };
};

export type MessagePage = {
  messages: Message[];
  next_before: number | null;
};

export type SearchHit = {
  message_id: string;
  conversation_id: string;
  body: string;
  seq: number;
};

function authHeaders(token: string) {
  return {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };
}

export async function login(username: string, password: string): Promise<TokenPair> {
  const res = await fetch(`${API_BASE}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) throw new Error("login failed");
  return res.json();
}

export async function register(username: string, password: string, displayName: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password, display_name: displayName }),
  });
  if (!res.ok) throw new Error("register failed");
}

export async function refreshToken(refresh: string): Promise<TokenPair> {
  const res = await fetch(`${API_BASE}/api/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refresh }),
  });
  if (!res.ok) throw new Error("refresh failed");
  return res.json();
}

export type Notification = {
  id: string;
  type: string;
  payload: Record<string, unknown>;
  created_at: string;
  read_at?: string | null;
};

export async function listNotifications(token: string): Promise<Notification[]> {
  const res = await fetch(`${API_BASE}/api/notifications`, { headers: authHeaders(token) });
  if (!res.ok) throw new Error("notifications failed");
  const data = await res.json();
  return data.notifications ?? [];
}

export async function markConversationRead(token: string, conversationId: string, lastReadSeq: number): Promise<void> {
  const res = await fetch(`${API_BASE}/api/conversations/${conversationId}/read`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ last_read_seq: lastReadSeq }),
  });
  if (!res.ok) throw new Error("mark read failed");
}

export async function listConversations(token: string): Promise<Conversation[]> {
  const res = await fetch(`${API_BASE}/api/conversations`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("list conversations failed");
  const data = await res.json();
  return data.conversations ?? [];
}

export async function listMessages(
  token: string,
  conversationId: string,
  beforeSeq?: number,
): Promise<MessagePage> {
  const params = new URLSearchParams({ limit: "50" });
  if (beforeSeq != null) params.set("before_seq", String(beforeSeq));
  const res = await fetch(`${API_BASE}/api/conversations/${conversationId}/messages?${params}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("list messages failed");
  const data = await res.json();
  return {
    messages: data.messages ?? [],
    next_before: data.next_before ?? null,
  };
}

export async function sendMessage(
  token: string,
  conversationId: string,
  body: string,
  attachmentObjectKey?: string,
): Promise<void> {
  const payload: Record<string, unknown> = {
    type: attachmentObjectKey ? "file" : "text",
    body,
    client_msg_id: crypto.randomUUID(),
  };
  if (attachmentObjectKey) {
    payload.attachment = { object_key: attachmentObjectKey };
  }
  const res = await fetch(`${API_BASE}/api/conversations/${conversationId}/messages`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
  if (!res.ok) throw new Error("send failed");
}

export async function presignUpload(
  token: string,
  file: File,
): Promise<{ upload_url: string; object_key: string }> {
  const res = await fetch(`${API_BASE}/api/media/upload-url`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({
      mime_type: file.type || "application/octet-stream",
      size_bytes: file.size,
    }),
  });
  if (!res.ok) throw new Error("presign upload failed");
  const data = await res.json();
  const put = await fetch(data.upload_url, {
    method: "PUT",
    body: file,
    headers: { "Content-Type": file.type || "application/octet-stream" },
  });
  if (!put.ok) throw new Error("upload failed");
  return { upload_url: data.upload_url, object_key: data.object_key };
}

export async function searchMessages(token: string, query: string): Promise<SearchHit[]> {
  const params = new URLSearchParams({ q: query, limit: "20" });
  const res = await fetch(`${API_BASE}/api/search/messages?${params}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("search failed");
  const data = await res.json();
  return data.results ?? [];
}

export async function addReaction(token: string, messageId: string, emoji: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/messages/${messageId}/reactions`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ emoji }),
  });
  if (!res.ok) throw new Error("reaction failed");
}

export async function reportMessage(
  token: string,
  conversationId: string,
  messageId: string,
  reason: string,
): Promise<void> {
  const res = await fetch(`${API_BASE}/api/conversations/${conversationId}/messages/${messageId}/report`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ reason }),
  });
  if (!res.ok) throw new Error("report failed");
}

export async function blockUser(token: string, userId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/blocks/${userId}`, {
    method: "POST",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("block failed");
}

export async function listRecommendations(token: string): Promise<{ channel_id: string; title: string }[]> {
  const res = await fetch(`${API_BASE}/api/recommendations/channels`, { headers: authHeaders(token) });
  if (!res.ok) return [];
  const data = await res.json();
  return (data.channels ?? []).map((c: { id?: string; channel_id?: string; title: string }) => ({
    channel_id: c.id ?? c.channel_id ?? "",
    title: c.title,
  }));
}

export type Reaction = { user_id: string; emoji: string; created_at?: string };

export async function listReactions(token: string, messageId: string): Promise<Reaction[]> {
  const res = await fetch(`${API_BASE}/api/messages/${messageId}/reactions`, { headers: authHeaders(token) });
  if (!res.ok) return [];
  const data = await res.json();
  return data.reactions ?? [];
}

export async function removeReaction(token: string, messageId: string, emoji: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("remove reaction failed");
}

export async function listReplies(token: string, convId: string, messageId: string): Promise<Message[]> {
  const res = await fetch(`${API_BASE}/api/conversations/${convId}/messages/${messageId}/replies`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("list replies failed");
  const data = await res.json();
  return data.replies ?? [];
}

export async function sendReply(token: string, convId: string, messageId: string, body: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/conversations/${convId}/messages/${messageId}/replies`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ body }),
  });
  if (!res.ok) throw new Error("send reply failed");
}

export type AdminUser = { id: string; username: string; display_name: string; is_admin: boolean };
export type AdminReport = { id: string; reason: string; message_id: string; conversation_id: string };

export async function adminListUsers(token: string): Promise<AdminUser[]> {
  const res = await fetch(`${API_BASE}/api/admin/users`, { headers: authHeaders(token) });
  if (!res.ok) throw new Error("admin users failed");
  const data = await res.json();
  return data.users ?? [];
}

export async function adminListReports(token: string): Promise<AdminReport[]> {
  const res = await fetch(`${API_BASE}/api/admin/reports`, { headers: authHeaders(token) });
  if (!res.ok) throw new Error("admin reports failed");
  const data = await res.json();
  return data.reports ?? [];
}

export async function pinMessage(token: string, conversationId: string, messageId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/conversations/${conversationId}/pins/${messageId}`, {
    method: "POST",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("pin failed");
}

export async function archiveConversation(token: string, conversationId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/conversations/${conversationId}/archive`, {
    method: "POST",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("archive failed");
}

export async function exportConversation(token: string, conversationId: string): Promise<Blob> {
  const res = await fetch(`${API_BASE}/api/conversations/${conversationId}/export`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("export failed");
  return res.blob();
}

export async function forwardMessage(token: string, messageId: string, targetConversationId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/messages/${messageId}/forward`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ target_conversation_id: targetConversationId }),
  });
  if (!res.ok) throw new Error("forward failed");
}

export async function subscribeChannel(token: string, channelId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/conversations/${channelId}/subscribe`, {
    method: "POST",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("subscribe failed");
}

export async function listFriendRecommendations(token: string): Promise<{ id: string; username: string; display_name: string }[]> {
  const res = await fetch(`${API_BASE}/api/recommendations/friends`, { headers: authHeaders(token) });
  if (!res.ok) return [];
  const data = await res.json();
  return data.friends ?? [];
}

export async function touchLastSeen(token: string): Promise<void> {
  await fetch(`${API_BASE}/api/presence/last-seen`, {
    method: "POST",
    headers: authHeaders(token),
  });
}

export type WSStatus = "connecting" | "open" | "closed";

export function connectWS(
  token: string,
  deviceId: string,
  onMessage: (data: unknown) => void,
  onStatus?: (status: WSStatus) => void,
): { close: () => void; send: (payload: unknown) => void } {
  let ws: WebSocket | null = null;
  let closed = false;
  let attempt = 0;
  let timer: number | undefined;

  const proto = window.location.protocol === "https:" ? "wss" : "ws";
  const host = window.location.host;
  const url = `${proto}://${host}/ws?token=${encodeURIComponent(token)}&device_id=${encodeURIComponent(deviceId)}`;

  function scheduleReconnect() {
    if (closed) return;
    attempt += 1;
    const delay = Math.min(30_000, 500 * 2 ** Math.min(attempt, 6));
    timer = window.setTimeout(connect, delay);
  }

  function connect() {
    if (closed) return;
    onStatus?.("connecting");
    ws = new WebSocket(url);
    ws.onopen = () => {
      attempt = 0;
      onStatus?.("open");
    };
    ws.onmessage = (evt) => {
      try {
        onMessage(JSON.parse(evt.data));
      } catch {
        onMessage(evt.data);
      }
    };
    ws.onclose = () => {
      onStatus?.("closed");
      scheduleReconnect();
    };
    ws.onerror = () => {
      ws?.close();
    };
  }

  connect();

  return {
    close: () => {
      closed = true;
      if (timer) window.clearTimeout(timer);
      ws?.close();
    },
    send: (payload: unknown) => {
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(payload));
      }
    },
  };
}
