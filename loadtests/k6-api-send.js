/**
 * EchoLine API Load Test — Message Send
 * Task: I006
 *
 * Tests the message send hot path under sustained load.
 * Stages: ramp-up → sustained → ramp-down
 *
 * Prerequisites:
 *   - EchoLine API running at API_BASE_URL
 *   - At least 2 test users registered (seeded via `make seed`)
 *   - TOKEN_A and TOKEN_B: JWT tokens for the two test users
 *   - CONV_ID: ID of a DM conversation between the two users
 *
 * Run:
 *   k6 run loadtests/k6-api-send.js
 *
 * Run with env overrides:
 *   API_BASE_URL=http://staging.echoline.io \
 *   TOKEN_A=eyJ... \
 *   CONV_ID=abc123 \
 *   k6 run loadtests/k6-api-send.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// ─── Configuration ────────────────────────────────────────────────────────────

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';
const TOKEN_A  = __ENV.TOKEN_A || 'test-token-a';
const TOKEN_B  = __ENV.TOKEN_B || 'test-token-b';
const CONV_ID  = __ENV.CONV_ID  || 'test-conv-id';

// ─── Custom Metrics ───────────────────────────────────────────────────────────

const messageSendErrors   = new Counter('echoline_message_send_errors');
const messageSendLatency  = new Trend('echoline_message_send_latency_ms', true);
const idempotencyRetries  = new Counter('echoline_idempotency_retries');
const successRate         = new Rate('echoline_message_send_success_rate');

// ─── Load Profile ─────────────────────────────────────────────────────────────

export const options = {
  stages: [
    { duration: '30s', target: 10  },  // ramp up to 10 VUs
    { duration: '60s', target: 50  },  // ramp up to 50 VUs (~ 100 RPS)
    { duration: '120s', target: 100 }, // sustained 100 VUs (~ 200 RPS)
    { duration: '30s', target: 0   },  // ramp down
  ],
  thresholds: {
    // 95th percentile message send must be under 200ms
    'echoline_message_send_latency_ms': ['p(95)<200'],
    // At least 99% of sends must succeed
    'echoline_message_send_success_rate': ['rate>0.99'],
    // Overall HTTP failure rate below 1%
    'http_req_failed': ['rate<0.01'],
  },
};

// ─── Helpers ──────────────────────────────────────────────────────────────────

function randomClientMsgId() {
  // Generate a UUID-like string for idempotency key
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

function authHeaders(token) {
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
}

// ─── Main Scenario ─────────────────────────────────────────────────────────────

export default function () {
  const token = (__VU % 2 === 0) ? TOKEN_A : TOKEN_B;
  const clientMsgId = randomClientMsgId();

  const payload = JSON.stringify({
    body: `Load test message from VU ${__VU} iter ${__ITER}`,
    client_msg_id: clientMsgId,
    type: 'text',
  });

  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/conversations/${CONV_ID}/messages`,
    payload,
    { headers: authHeaders(token), timeout: '5s' }
  );
  const latency = Date.now() - start;

  messageSendLatency.add(latency);

  const ok = check(res, {
    'message send 200': (r) => r.status === 200,
    'response has message id': (r) => {
      try { return JSON.parse(r.body).id !== undefined; } catch { return false; }
    },
    'response has seq': (r) => {
      try { return JSON.parse(r.body).seq > 0; } catch { return false; }
    },
  });

  if (!ok) {
    messageSendErrors.add(1);
    successRate.add(0);
  } else {
    successRate.add(1);
  }

  // Test idempotency: re-send the same client_msg_id
  if (__ITER % 20 === 0) {
    const retryRes = http.post(
      `${BASE_URL}/api/conversations/${CONV_ID}/messages`,
      payload,
      { headers: authHeaders(token), timeout: '5s' }
    );
    const retryOk = check(retryRes, {
      'idempotent retry returns 200': (r) => r.status === 200,
      'idempotent retry returns same message id': (r) => {
        try {
          const orig = JSON.parse(res.body).id;
          const retry = JSON.parse(r.body).id;
          return orig === retry;
        } catch { return false; }
      },
    });
    idempotencyRetries.add(1);
    if (!retryOk) messageSendErrors.add(1);
  }

  sleep(0.1); // 100ms think time → ~10 RPS per VU
}

// ─── Setup: Health Check ───────────────────────────────────────────────────────

export function setup() {
  const res = http.get(`${BASE_URL}/health`);
  if (res.status !== 200) {
    throw new Error(`API health check failed: ${res.status} ${res.body}`);
  }
  console.log(`[setup] API healthy at ${BASE_URL}`);
  return { base_url: BASE_URL };
}

// ─── Teardown: Summary ─────────────────────────────────────────────────────────

export function teardown(data) {
  console.log(`[teardown] Load test complete. Base URL: ${data.base_url}`);
}
