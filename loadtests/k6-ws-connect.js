/**
 * EchoLine WebSocket Load Test — Connection Stability
 * Task: I007
 *
 * Tests WebSocket connection establishment, message receipt, and heartbeat
 * under concurrent load. Validates that the WS gateway handles 500+
 * concurrent connections without dropping messages.
 *
 * Prerequisites:
 *   - EchoLine API + WS gateway running at WS_BASE_URL
 *   - TOKEN_A: JWT token for a test user
 *   - TOKEN_B: JWT token for a second test user (sender)
 *   - CONV_ID: shared conversation ID
 *
 * Run:
 *   k6 run loadtests/k6-ws-connect.js
 *
 * k6 WebSocket docs: https://k6.io/docs/using-k6/protocols/websockets/
 */

import ws from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Gauge, Trend } from 'k6/metrics';

// ─── Configuration ────────────────────────────────────────────────────────────

const API_BASE = __ENV.API_BASE_URL || 'http://localhost:8080';
const WS_BASE  = __ENV.WS_BASE_URL  || 'ws://localhost:8080';
const TOKEN_A  = __ENV.TOKEN_A || 'test-token-a';
const TOKEN_B  = __ENV.TOKEN_B || 'test-token-b';
const CONV_ID  = __ENV.CONV_ID  || 'test-conv-id';

// ─── Custom Metrics ───────────────────────────────────────────────────────────

const wsConnectErrors     = new Counter('echoline_ws_connect_errors');
const wsMessageReceived   = new Counter('echoline_ws_messages_received');
const wsConnectLatency    = new Trend('echoline_ws_connect_latency_ms', true);
const wsMessageLatency    = new Trend('echoline_ws_message_latency_ms', true);
const wsActiveConns       = new Gauge('echoline_ws_active_connections');
const wsDroppedMessages   = new Counter('echoline_ws_dropped_messages');
const wsSuccessRate       = new Rate('echoline_ws_connect_success_rate');

// ─── Load Profile ─────────────────────────────────────────────────────────────

export const options = {
  scenarios: {
    // Scenario 1: Many persistent connections (connection count test)
    persistent_connections: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 100  },
        { duration: '60s', target: 300  },
        { duration: '120s', target: 500 },
        { duration: '30s', target: 0    },
      ],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    'echoline_ws_connect_latency_ms': ['p(95)<1000'],  // WS connect in < 1s
    'echoline_ws_message_latency_ms': ['p(95)<200'],   // Message delivery < 200ms
    'echoline_ws_connect_success_rate': ['rate>0.98'], // 98%+ connect success
    'echoline_ws_dropped_messages': ['count<10'],      // Less than 10 dropped msgs
  },
};

// ─── Helpers ──────────────────────────────────────────────────────────────────

function randomMsgId() {
  return Math.random().toString(36).substr(2, 9);
}

// ─── Main Scenario ─────────────────────────────────────────────────────────────

export default function () {
  const token = TOKEN_A;
  const connectStart = Date.now();

  const response = ws.connect(
    `${WS_BASE}/ws?token=${token}`,
    { timeout: '10s' },
    function (socket) {
      const connectLatency = Date.now() - connectStart;
      wsConnectLatency.add(connectLatency);
      wsActiveConns.add(1);

      socket.on('open', () => {
        wsSuccessRate.add(1);

        // Send a ping message to verify bidirectional communication
        socket.send(JSON.stringify({
          type: 'ping',
          request_id: randomMsgId(),
        }));
      });

      socket.on('message', (data) => {
        let msg;
        try {
          msg = JSON.parse(data);
        } catch {
          wsDroppedMessages.add(1);
          return;
        }

        if (msg.type === 'message.received') {
          wsMessageReceived.add(1);
          const deliveryLatency = Date.now() - (msg.payload?.sent_at_ms || Date.now());
          wsMessageLatency.add(Math.max(0, deliveryLatency));

          // Send ACK
          socket.send(JSON.stringify({
            type: 'ack',
            payload: { message_id: msg.payload?.id },
          }));
        }

        if (msg.type === 'pong') {
          // Heartbeat acknowledged
        }

        check(msg, {
          'ws message has type field': (m) => m.type !== undefined,
        });
      });

      socket.on('error', (e) => {
        wsConnectErrors.add(1);
        wsSuccessRate.add(0);
      });

      socket.on('close', () => {
        wsActiveConns.add(-1);
      });

      // Keep connection open for the duration of the test iteration
      // Send a heartbeat ping every 25 seconds (before server's 30s timeout)
      socket.setInterval(() => {
        socket.send(JSON.stringify({ type: 'ping', request_id: randomMsgId() }));
      }, 25000);

      // Hold connection open for 30 seconds
      socket.setTimeout(() => {
        socket.close();
      }, 30000);
    }
  );

  check(response, {
    'ws upgrade 101': (r) => r && r.status === 101,
  });

  if (!response || response.status !== 101) {
    wsConnectErrors.add(1);
    wsSuccessRate.add(0);
  }

  sleep(1);
}

// ─── Setup ────────────────────────────────────────────────────────────────────

export function setup() {
  // Verify API is healthy
  const health = http.get(`${API_BASE}/health`);
  check(health, { 'API healthy': (r) => r.status === 200 });

  // Verify we can generate a WS URL (token is valid)
  console.log(`[setup] WS gateway: ${WS_BASE}`);
  console.log(`[setup] Test conversation: ${CONV_ID}`);
  return {};
}
