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

export type WSStatus = "connecting" | "open" | "closed";

export function connectWS(
  token: string,
  deviceId: string,
  onMessage: (data: unknown) => void,
  onStatus?: (status: WSStatus) => void,
): { close: () => void } {
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
  };
}
