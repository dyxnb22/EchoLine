import {
  authedBlob,
  authedJSON,
  authedJSONOr,
  authedVoid,
  publicJSON,
} from "./api/http";

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

export type Notification = {
  id: string;
  type: string;
  payload: Record<string, unknown>;
  created_at: string;
  read_at?: string | null;
};

export type Reaction = { user_id: string; emoji: string; created_at?: string };

export type AdminUser = { id: string; username: string; display_name: string; is_admin: boolean };
export type AdminReport = { id: string; reason: string; message_id: string; conversation_id: string };
export type DLQEvent = { id: string; event_type: string; status: string; attempts: number };

export type WSStatus = "connecting" | "open" | "closed";

export async function login(username: string, password: string): Promise<TokenPair> {
  return publicJSON<TokenPair>("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  }, "login failed");
}

export async function register(username: string, password: string, displayName: string): Promise<void> {
  await publicJSON("/api/auth/register", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password, display_name: displayName }),
  }, "register failed");
}

export async function refreshToken(refresh: string): Promise<TokenPair> {
  return publicJSON<TokenPair>("/api/auth/refresh", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refresh }),
  }, "refresh failed");
}

export async function listNotifications(token: string): Promise<Notification[]> {
  const data = await authedJSON<{ notifications?: Notification[] }>(
    token, "/api/notifications", {}, "notifications failed",
  );
  return data.notifications ?? [];
}

export async function markConversationRead(token: string, conversationId: string, lastReadSeq: number): Promise<void> {
  await authedVoid(token, `/api/conversations/${conversationId}/read`, {
    method: "POST",
    body: JSON.stringify({ last_read_seq: lastReadSeq }),
  }, "mark read failed");
}

export async function listConversations(token: string): Promise<Conversation[]> {
  const data = await authedJSON<{ conversations?: Conversation[] }>(
    token, "/api/conversations", {}, "list conversations failed",
  );
  return data.conversations ?? [];
}

export async function listMessages(
  token: string,
  conversationId: string,
  beforeSeq?: number,
): Promise<MessagePage> {
  const params = new URLSearchParams({ limit: "50" });
  if (beforeSeq != null) params.set("before_seq", String(beforeSeq));
  const data = await authedJSON<{ messages?: Message[]; next_before?: number | null }>(
    token,
    `/api/conversations/${conversationId}/messages?${params}`,
    {},
    "list messages failed",
  );
  return { messages: data.messages ?? [], next_before: data.next_before ?? null };
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
  await authedVoid(token, `/api/conversations/${conversationId}/messages`, {
    method: "POST",
    body: JSON.stringify(payload),
  }, "send failed");
}

export async function presignUpload(
  token: string,
  file: File,
): Promise<{ upload_url: string; object_key: string }> {
  const data = await authedJSON<{ upload_url: string; object_key: string }>(
    token,
    "/api/media/upload-url",
    {
      method: "POST",
      body: JSON.stringify({
        mime_type: file.type || "application/octet-stream",
        size_bytes: file.size,
      }),
    },
    "presign upload failed",
  );
  const put = await fetch(data.upload_url, {
    method: "PUT",
    body: file,
    headers: { "Content-Type": file.type || "application/octet-stream" },
  });
  if (!put.ok) throw new Error("upload failed");
  return data;
}

export async function searchMessages(token: string, query: string): Promise<SearchHit[]> {
  const params = new URLSearchParams({ q: query, limit: "20" });
  const data = await authedJSON<{ results?: SearchHit[] }>(
    token, `/api/search/messages?${params}`, {}, "search failed",
  );
  return data.results ?? [];
}

export async function addReaction(token: string, messageId: string, emoji: string): Promise<void> {
  await authedVoid(token, `/api/messages/${messageId}/reactions`, {
    method: "POST",
    body: JSON.stringify({ emoji }),
  }, "reaction failed");
}

export async function reportMessage(
  token: string,
  conversationId: string,
  messageId: string,
  reason: string,
): Promise<void> {
  await authedVoid(token, `/api/conversations/${conversationId}/messages/${messageId}/report`, {
    method: "POST",
    body: JSON.stringify({ reason }),
  }, "report failed");
}

export async function blockUser(token: string, userId: string): Promise<void> {
  await authedVoid(token, `/api/blocks/${userId}`, { method: "POST" }, "block failed");
}

export async function listRecommendations(token: string): Promise<{ channel_id: string; title: string }[]> {
  const data = await authedJSONOr<{ channels?: { id?: string; channel_id?: string; title: string }[] }>(
    token, "/api/recommendations/channels", {}, { channels: [] },
  );
  return (data.channels ?? []).map((c) => ({
    channel_id: c.id ?? c.channel_id ?? "",
    title: c.title,
  }));
}

export async function listReactions(token: string, messageId: string): Promise<Reaction[]> {
  const data = await authedJSONOr<{ reactions?: Reaction[] }>(
    token, `/api/messages/${messageId}/reactions`, {}, { reactions: [] },
  );
  return data.reactions ?? [];
}

export async function removeReaction(token: string, messageId: string, emoji: string): Promise<void> {
  await authedVoid(
    token,
    `/api/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`,
    { method: "DELETE" },
    "remove reaction failed",
  );
}

export async function listReplies(token: string, convId: string, messageId: string): Promise<Message[]> {
  const data = await authedJSON<{ replies?: Message[] }>(
    token,
    `/api/conversations/${convId}/messages/${messageId}/replies`,
    {},
    "list replies failed",
  );
  return data.replies ?? [];
}

export async function sendReply(token: string, convId: string, messageId: string, body: string): Promise<void> {
  await authedVoid(token, `/api/conversations/${convId}/messages/${messageId}/replies`, {
    method: "POST",
    body: JSON.stringify({ body }),
  }, "send reply failed");
}

export async function adminListUsers(token: string): Promise<AdminUser[]> {
  const data = await authedJSON<{ users?: AdminUser[] }>(token, "/api/admin/users", {}, "admin users failed");
  return data.users ?? [];
}

export async function adminListReports(token: string): Promise<AdminReport[]> {
  const data = await authedJSON<{ reports?: AdminReport[] }>(token, "/api/admin/reports", {}, "admin reports failed");
  return data.reports ?? [];
}

export async function pinMessage(token: string, conversationId: string, messageId: string): Promise<void> {
  await authedVoid(token, `/api/conversations/${conversationId}/pins/${messageId}`, { method: "POST" }, "pin failed");
}

export async function archiveConversation(token: string, conversationId: string): Promise<void> {
  await authedVoid(token, `/api/conversations/${conversationId}/archive`, { method: "POST" }, "archive failed");
}

export async function exportConversation(token: string, conversationId: string): Promise<Blob> {
  return authedBlob(token, `/api/conversations/${conversationId}/export`, {}, "export failed");
}

export async function forwardMessage(token: string, messageId: string, targetConversationId: string): Promise<void> {
  await authedVoid(token, `/api/messages/${messageId}/forward`, {
    method: "POST",
    body: JSON.stringify({ target_conversation_id: targetConversationId }),
  }, "forward failed");
}

export async function subscribeChannel(token: string, channelId: string): Promise<void> {
  await authedVoid(token, `/api/conversations/${channelId}/subscribe`, { method: "POST" }, "subscribe failed");
}

export async function listFriendRecommendations(token: string): Promise<{ id: string; username: string; display_name: string }[]> {
  const data = await authedJSONOr<{ friends?: { id: string; username: string; display_name: string }[] }>(
    token, "/api/recommendations/friends", {}, { friends: [] },
  );
  return data.friends ?? [];
}

export async function touchLastSeen(token: string): Promise<void> {
  await authedVoid(token, "/api/presence/last-seen", { method: "POST" }, "last seen failed");
}

export async function markNotificationRead(token: string, id: string): Promise<void> {
  await authedVoid(token, `/api/notifications/${id}/read`, { method: "POST" }, "mark notification read failed");
}

export async function editMessage(token: string, convId: string, messageId: string, body: string): Promise<void> {
  await authedVoid(token, `/api/conversations/${convId}/messages/${messageId}`, {
    method: "PATCH",
    body: JSON.stringify({ body }),
  }, "edit failed");
}

export async function recallMessage(token: string, convId: string, messageId: string): Promise<void> {
  await authedVoid(token, `/api/conversations/${convId}/messages/${messageId}/recall`, { method: "POST" }, "recall failed");
}

export async function listPins(token: string, convId: string): Promise<{ message_id: string }[]> {
  const data = await authedJSONOr<{ pins?: { message_id: string }[] }>(
    token, `/api/conversations/${convId}/pins`, {}, { pins: [] },
  );
  return data.pins ?? [];
}

export async function listArchived(token: string): Promise<Conversation[]> {
  const data = await authedJSONOr<{ conversations?: Conversation[] }>(
    token, "/api/conversations/archived", {}, { conversations: [] },
  );
  return data.conversations ?? [];
}

export async function unarchiveConversation(token: string, convId: string): Promise<void> {
  await authedVoid(token, `/api/conversations/${convId}/unarchive`, { method: "POST" }, "unarchive failed");
}

export async function inviteMember(token: string, convId: string, userId: string): Promise<void> {
  await authedVoid(token, `/api/conversations/${convId}/members`, {
    method: "POST",
    body: JSON.stringify({ user_id: userId, role: "member" }),
  }, "invite failed");
}

export async function removeMember(token: string, convId: string, userId: string): Promise<void> {
  await authedVoid(token, `/api/conversations/${convId}/members/${userId}`, { method: "DELETE" }, "remove failed");
}

export async function registerPushToken(token: string, deviceId: string, pushToken: string, platform: string): Promise<void> {
  await authedVoid(token, "/api/push/tokens", {
    method: "POST",
    body: JSON.stringify({ device_id: deviceId, token: pushToken, platform }),
  }, "push register failed");
}

export async function adminListDLQ(token: string): Promise<DLQEvent[]> {
  const data = await authedJSON<{ events?: DLQEvent[] }>(token, "/api/admin/dlq", {}, "dlq list failed");
  return data.events ?? [];
}

export async function adminReplayDLQ(token: string, id: string): Promise<void> {
  await authedVoid(token, `/api/admin/dlq/${id}/replay`, { method: "POST" }, "dlq replay failed");
}

export async function grantChannelEntitlement(
  token: string,
  channelId: string,
  userId: string,
  reference: string,
): Promise<void> {
  await authedVoid(token, `/api/channels/${channelId}/entitlements/grant`, {
    method: "POST",
    body: JSON.stringify({ user_id: userId, reference }),
  }, "grant failed");
}

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
