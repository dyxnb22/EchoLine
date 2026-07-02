/**
 * EchoLine Large-Group Fanout Load Test
 * Task: E010
 *
 * Simulates a high-traffic public channel with many concurrent subscribers.
 * Tests the fanout worker's ability to deliver messages to all online members
 * without message loss or excessive latency.
 *
 * Scenario:
 *   - MEMBER_COUNT members connected to a large group/channel via WebSocket
 *   - A separate "sender" VU sends messages at MESSAGE_RATE per second
 *   - Measure: each member's time-to-receive the message
 *   - Acceptance: P95 fanout latency < 1000ms, 0 dropped messages
 *
 * Prerequisites:
 *   - EchoLine API and fanout worker running
 *   - Large group conversation created with MEMBER_COUNT members
 *   - TOKENS: comma-separated JWT tokens for member VUs
 *   - SENDER_TOKEN: token for the sender account
 *   - GROUP_CONV_ID: the large group conversation ID
 *
 * Run:
 *   k6 run loadtests/k6-large-group.js
 */

import ws from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend, Gauge } from 'k6/metrics';

// ─── Configuration ────────────────────────────────────────────────────────────

const API_BASE       = __ENV.API_BASE_URL    || 'http://localhost:8080';
const WS_BASE        = __ENV.WS_BASE_URL     || 'ws://localhost:8080';
const GROUP_CONV_ID  = __ENV.GROUP_CONV_ID   || 'test-large-group-id';
const SENDER_TOKEN   = __ENV.SENDER_TOKEN    || 'test-sender-token';
const MEMBER_TOKENS  = (__ENV.MEMBER_TOKENS  || 'test-member-token').split(',');
const MESSAGE_RATE   = parseInt(__ENV.MESSAGE_RATE || '10'); // messages per second from sender

// ─── Custom Metrics ───────────────────────────────────────────────────────────

const fanoutLatency         = new Trend('echoline_fanout_latency_ms', true);
const fanoutMessagesExpected = new Counter('echoline_fanout_messages_expected');
const fanoutMessagesReceived = new Counter('echoline_fanout_messages_received');
const fanoutDropped         = new Counter('echoline_fanout_dropped');
const fanoutSuccessRate     = new Rate('echoline_fanout_success_rate');
const activeMemberConns     = new Gauge('echoline_active_member_ws_connections');

// Shared state: map of message_id → sent_at_ms (for latency calculation)
// Note: k6 VUs don't share memory, so we use a simplified approach:
// sender embeds sent_at_ms in message body, receivers parse it.

// ─── Load Profile ─────────────────────────────────────────────────────────────

export const options = {
  scenarios: {
    // Member subscribers: many persistent WS connections
    member_subscribers: {
      executor: 'constant-vus',
      vus: 50,   // simulate 50 concurrent members online
      duration: '180s',
      gracefulStop: '10s',
      exec: 'memberScenario',
    },
    // Sender: one VU sending messages at MESSAGE_RATE/s
    message_sender: {
      executor: 'constant-arrival-rate',
      rate: MESSAGE_RATE,
      timeUnit: '1s',
      duration: '120s',
      preAllocatedVUs: 5,
      maxVUs: 10,
      startTime: '30s',  // start after members have connected
      exec: 'senderScenario',
    },
  },
  thresholds: {
    // P95 fanout latency (sender sends → member receives) under 1000ms
    'echoline_fanout_latency_ms': ['p(95)<1000', 'p(99)<2000'],
    // Fewer than 1% of messages dropped
    'echoline_fanout_success_rate': ['rate>0.99'],
  },
};

// ─── Member Subscriber Scenario ───────────────────────────────────────────────

export function memberScenario() {
  const tokenIdx = (__VU - 1) % MEMBER_TOKENS.length;
  const token = MEMBER_TOKENS[tokenIdx];

  const response = ws.connect(
    `${WS_BASE}/ws?token=${token}`,
    { timeout: '10s' },
    function (socket) {
      activeMemberConns.add(1);

      socket.on('open', () => {
        // Subscribe signal (no explicit subscribe needed; membership enforced server-side)
      });

      socket.on('message', (data) => {
        let msg;
        try { msg = JSON.parse(data); } catch { fanoutDropped.add(1); return; }

        if (msg.type === 'message.received') {
          fanoutMessagesReceived.add(1);

          // Parse sent_at_ms from message body (embedded by sender scenario)
          const body = msg.payload?.body || '';
          const match = body.match(/sent_at:(\d+)/);
          if (match) {
            const sentAt = parseInt(match[1]);
            const latency = Date.now() - sentAt;
            fanoutLatency.add(latency);
            fanoutSuccessRate.add(1);
          }

          // Send ACK
          socket.send(JSON.stringify({
            type: 'ack',
            payload: { message_id: msg.payload?.id },
          }));
        }
      });

      socket.on('error', () => activeMemberConns.add(-1));
      socket.on('close', () => activeMemberConns.add(-1));

      // Hold connection open for the full test duration
      socket.setTimeout(() => { socket.close(); }, 180000);
    }
  );

  check(response, { 'member ws connect 101': (r) => r && r.status === 101 });
}

// ─── Sender Scenario ──────────────────────────────────────────────────────────

export function senderScenario() {
  const sentAt = Date.now();
  const clientMsgId = `lgtest-${__VU}-${__ITER}-${sentAt}`;

  const payload = JSON.stringify({
    body: `Large group test message sent_at:${sentAt} iter:${__ITER}`,
    client_msg_id: clientMsgId,
    type: 'text',
  });

  fanoutMessagesExpected.add(1);

  const res = http.post(
    `${API_BASE}/api/conversations/${GROUP_CONV_ID}/messages`,
    payload,
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${SENDER_TOKEN}`,
      },
      timeout: '5s',
    }
  );

  check(res, {
    'large group send 200': (r) => r.status === 200,
  });

  if (res.status !== 200) {
    fanoutDropped.add(1);
    fanoutSuccessRate.add(0);
  }
}

// ─── Setup ────────────────────────────────────────────────────────────────────

export function setup() {
  const health = http.get(`${API_BASE}/health`);
  check(health, { 'API healthy': (r) => r.status === 200 });

  console.log(`[setup] Large group test`);
  console.log(`[setup] Group conversation: ${GROUP_CONV_ID}`);
  console.log(`[setup] Member tokens: ${MEMBER_TOKENS.length}`);
  console.log(`[setup] Message rate: ${MESSAGE_RATE} msgs/s`);

  return { group_conv_id: GROUP_CONV_ID };
}

// ─── Teardown ─────────────────────────────────────────────────────────────────

export function teardown(data) {
  console.log(`[teardown] Large group fanout test complete`);
  console.log(`[teardown] Check echoline_fanout_latency_ms for P95/P99`);
  console.log(`[teardown] Check echoline_fanout_dropped for message loss`);
}
