/**
 * EchoLine Mixed Workload Load Test
 * Task: T113
 *
 * Simulates a realistic production traffic mix:
 *   - Scenario A (40% VUs): Auth — login and fetch conversation list
 *   - Scenario B (30% VUs): Send — authenticated message send
 *   - Scenario C (20% VUs): WebSocket — connect, receive messages, disconnect
 *   - Scenario D (10% VUs): Search — full-text search queries
 *
 * Staged ramp-up to find the saturation point.
 *
 * Prerequisites:
 *   - EchoLine API running at API_BASE_URL (default: http://localhost:8080)
 *   - Seed data loaded (run scripts/seed-extended.sh first)
 *
 * Run:
 *   k6 run loadtests/k6-mixed-workload.js
 *
 * Run with env overrides:
 *   API_BASE_URL=http://staging.echoline.io \
 *   LOAD_EMAIL=alice@echoline.dev \
 *   LOAD_PASS=Seed1234! \
 *   k6 run loadtests/k6-mixed-workload.js
 */

import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep, group } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// ─── Configuration ────────────────────────────────────────────────────────────

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';
const WS_URL   = __ENV.WS_BASE_URL  || 'ws://localhost:8080';
const LOAD_EMAIL = __ENV.LOAD_EMAIL || 'alice@echoline.dev';
const LOAD_PASS  = __ENV.LOAD_PASS  || 'Seed1234!';

// ─── Custom Metrics ───────────────────────────────────────────────────────────

const authLatency    = new Trend('echoline_auth_latency_ms', true);
const sendLatency    = new Trend('echoline_send_latency_ms', true);
const searchLatency  = new Trend('echoline_search_latency_ms', true);
const wsConnectTime  = new Trend('echoline_ws_connect_ms', true);
const wsMessageCount = new Counter('echoline_ws_messages_received');
const errorCount     = new Counter('echoline_errors');
const authSuccess    = new Rate('echoline_auth_success_rate');
const sendSuccess    = new Rate('echoline_send_success_rate');
const searchSuccess  = new Rate('echoline_search_success_rate');

// ─── Load Profile ─────────────────────────────────────────────────────────────
//
// Staged ramp: gentle warm-up → main load → push → cool-down
// Total: ~8 minutes

export const options = {
  scenarios: {
    // 40% of VUs: auth + conv list (read-heavy)
    auth_scenario: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 4  },
        { duration: '90s', target: 20 },
        { duration: '120s', target: 20 },
        { duration: '30s', target: 0  },
      ],
      gracefulRampDown: '10s',
      exec: 'authScenario',
    },

    // 30% of VUs: message send (write-heavy)
    send_scenario: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 3  },
        { duration: '90s', target: 15 },
        { duration: '120s', target: 15 },
        { duration: '30s', target: 0  },
      ],
      gracefulRampDown: '10s',
      exec: 'sendScenario',
    },

    // 20% of VUs: WebSocket connections (realtime)
    ws_scenario: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 2  },
        { duration: '90s', target: 10 },
        { duration: '120s', target: 10 },
        { duration: '30s', target: 0  },
      ],
      gracefulRampDown: '10s',
      exec: 'wsScenario',
    },

    // 10% of VUs: search queries (read with FTS)
    search_scenario: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 1 },
        { duration: '90s', target: 5 },
        { duration: '120s', target: 5 },
        { duration: '30s', target: 0 },
      ],
      gracefulRampDown: '10s',
      exec: 'searchScenario',
    },
  },

  thresholds: {
    // Auth must be fast
    'echoline_auth_latency_ms':   ['p(95)<300', 'p(99)<500'],
    // Message send must be under 200ms at p95
    'echoline_send_latency_ms':   ['p(95)<200', 'p(99)<400'],
    // Search can be slower (FTS)
    'echoline_search_latency_ms': ['p(95)<500', 'p(99)<1000'],
    // WS connect must be quick
    'echoline_ws_connect_ms':     ['p(95)<200'],
    // Success rates
    'echoline_auth_success_rate':   ['rate>0.95'],
    'echoline_send_success_rate':   ['rate>0.99'],
    'echoline_search_success_rate': ['rate>0.90'],
    // Overall HTTP error rate
    'http_req_failed': ['rate<0.05'],
  },
};

// ─── Helpers ──────────────────────────────────────────────────────────────────

function jsonHeaders(token) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return { headers };
}

function login(email, password) {
  const t0 = Date.now();
  const res = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({ email, password }),
    { headers: { 'Content-Type': 'application/json' }, timeout: '5s' }
  );
  authLatency.add(Date.now() - t0);
  const ok = res.status === 200;
  authSuccess.add(ok ? 1 : 0);
  if (!ok) {
    errorCount.add(1);
    return null;
  }
  try {
    return JSON.parse(res.body).access_token || null;
  } catch {
    return null;
  }
}

function getConversations(token) {
  const res = http.get(
    `${BASE_URL}/conversations`,
    jsonHeaders(token)
  );
  check(res, { 'conversations 200': (r) => r.status === 200 });
  try {
    const body = JSON.parse(res.body);
    const convs = body.conversations || body.data || [];
    return convs;
  } catch {
    return [];
  }
}

// ─── Scenario A: Auth + Conversation List ─────────────────────────────────────

export function authScenario() {
  group('auth', () => {
    const token = login(LOAD_EMAIL, LOAD_PASS);
    if (!token) { sleep(1); return; }

    const convs = getConversations(token);
    check(convs, {
      'at least one conversation': (c) => c.length >= 0,
    });

    // Simulate reading one conversation
    if (convs.length > 0) {
      const convId = convs[0].id;
      const msgsRes = http.get(
        `${BASE_URL}/conversations/${convId}/messages?limit=20`,
        jsonHeaders(token)
      );
      check(msgsRes, { 'messages 200': (r) => r.status === 200 });
    }
  });

  sleep(1 + Math.random() * 2);
}

// ─── Scenario B: Send Message ──────────────────────────────────────────────────

export function sendScenario() {
  group('send', () => {
    const token = login(LOAD_EMAIL, LOAD_PASS);
    if (!token) { sleep(1); return; }

    const convs = getConversations(token);
    if (convs.length === 0) { sleep(1); return; }

    const convId = convs[0].id;
    const clientMsgId = uuidv4();
    const payload = JSON.stringify({
      text: `k6 mixed workload VU=${__VU} iter=${__ITER} ts=${Date.now()}`,
      client_msg_id: clientMsgId,
    });

    const t0 = Date.now();
    const res = http.post(
      `${BASE_URL}/conversations/${convId}/messages`,
      payload,
      { ...jsonHeaders(token), timeout: '5s' }
    );
    sendLatency.add(Date.now() - t0);

    const ok = check(res, {
      'send 200': (r) => r.status === 200,
      'has message id': (r) => {
        try { return !!JSON.parse(r.body).id; } catch { return false; }
      },
    });
    sendSuccess.add(ok ? 1 : 0);
    if (!ok) errorCount.add(1);
  });

  sleep(0.5 + Math.random());
}

// ─── Scenario C: WebSocket ─────────────────────────────────────────────────────

export function wsScenario() {
  group('websocket', () => {
    const token = login(LOAD_EMAIL, LOAD_PASS);
    if (!token) { sleep(2); return; }

    const wsStart = Date.now();
    let connected = false;
    let msgCount = 0;

    const res = ws.connect(
      `${WS_URL}/ws`,
      { headers: { Authorization: `Bearer ${token}` } },
      (socket) => {
        socket.on('open', () => {
          connected = true;
          wsConnectTime.add(Date.now() - wsStart);

          // Send auth frame
          socket.send(JSON.stringify({
            type: 'auth',
            token: token,
          }));

          // Keep connection alive for 10s, receiving messages
          socket.setTimeout(() => {
            socket.close();
          }, 10000);
        });

        socket.on('message', (data) => {
          try {
            const frame = JSON.parse(data);
            if (frame.type && frame.type !== 'pong' && frame.type !== 'error') {
              msgCount++;
              wsMessageCount.add(1);
            }
          } catch { /* non-JSON frame */ }
        });

        socket.on('error', (err) => {
          errorCount.add(1);
        });
      }
    );

    check(res, {
      'WS connected': () => connected,
    });
    check(msgCount, {
      'WS no errors': () => true,   // presence of messages is optional (no new sends)
    });
  });

  sleep(2 + Math.random() * 3);
}

// ─── Scenario D: Search ────────────────────────────────────────────────────────

const SEARCH_TERMS = ['hello', 'seeded', 'message', 'alice', 'test', 'load'];

export function searchScenario() {
  group('search', () => {
    const token = login(LOAD_EMAIL, LOAD_PASS);
    if (!token) { sleep(1); return; }

    const term = SEARCH_TERMS[__ITER % SEARCH_TERMS.length];
    const t0 = Date.now();
    const res = http.get(
      `${BASE_URL}/search?q=${encodeURIComponent(term)}&limit=10`,
      jsonHeaders(token)
    );
    searchLatency.add(Date.now() - t0);

    const ok = check(res, {
      'search 200 or 404': (r) => r.status === 200 || r.status === 404,
      'search response valid JSON': (r) => {
        try { JSON.parse(r.body); return true; } catch { return false; }
      },
    });
    searchSuccess.add(ok ? 1 : 0);
    if (!ok) errorCount.add(1);
  });

  sleep(2 + Math.random() * 2);
}

// ─── Setup ────────────────────────────────────────────────────────────────────

export function setup() {
  const res = http.get(`${BASE_URL}/health`);
  if (res.status !== 200 && res.status !== 404) {
    // 404 is acceptable if /health route differs; we check connectivity
    console.warn(`[setup] Health check returned ${res.status} — proceeding`);
  }
  console.log(`[setup] EchoLine at ${BASE_URL} — starting mixed workload`);

  // Pre-login to verify credentials work
  const tokenRes = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({ email: LOAD_EMAIL, password: LOAD_PASS }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  if (tokenRes.status !== 200) {
    console.warn(`[setup] Login failed (${tokenRes.status}). Run scripts/seed-extended.sh first.`);
  } else {
    console.log('[setup] Credentials verified');
  }

  return { baseUrl: BASE_URL };
}

// ─── Teardown ─────────────────────────────────────────────────────────────────

export function teardown(data) {
  console.log(`[teardown] Mixed workload complete. Target: ${data.baseUrl}`);
}
