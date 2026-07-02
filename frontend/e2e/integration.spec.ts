import { test } from "@playwright/test";

/**
 * Full-stack integration tests require a running compose stack.
 *
 * Run locally:
 *   make dev-up && make dev-app
 *   cd frontend && npx playwright test e2e/integration.spec.ts
 *
 * This placeholder keeps extended.spec.ts reference valid until CI gains a compose job.
 */
test.describe("Full stack integration (placeholder)", () => {
  test.skip("requires make dev-up stack", async () => {
    // Implement: register two users, send message, assert WS delivery
  });
});
