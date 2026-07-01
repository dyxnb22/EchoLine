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
};

export type MessagePage = {
  messages: Message[];
  next_before: number | null;
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

export async function sendMessage(token: string, conversationId: string, body: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/conversations/${conversationId}/messages`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ type: "text", body, client_msg_id: crypto.randomUUID() }),
  });
  if (!res.ok) throw new Error("send failed");
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
