import { test, expect } from "@playwright/test";

/**
 * Extended smoke: verifies composer interaction (still mocked API).
 * Full stack integration requires `make dev-up` — see e2e/integration.spec.ts.
 */
test.describe("Extended UI smoke", () => {
  test("new chat button visible when logged in", async ({ page }) => {
    await page.goto("/");
    await page.evaluate(() => {
      localStorage.setItem("echoline_token", "mock-token");
      localStorage.setItem("echoline_refresh", "mock-refresh");
      localStorage.setItem("echoline_device", "playwright-device");
    });
    await page.route("**/api/**", async (route) => {
      const url = route.request().url();
      if (url.includes("/api/conversations/archived")) {
        return route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({ archived: [] }),
        });
      }
      if (url.includes("/api/conversations") && route.request().method() === "GET") {
        return route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            conversations: [{
              id: "550e8400-e29b-41d4-a716-446655440001",
              type: "group",
              title: "Test",
              unread: 0,
              latest_seq: 1,
              can_publish: true,
            }],
          }),
        });
      }
      if (url.includes("/api/recommendations")) {
        return route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
      }
      if (url.includes("/api/notifications")) {
        return route.fulfill({ status: 200, contentType: "application/json", body: '{"notifications":[]}' });
      }
      return route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
    });
    await page.route("**/ws**", (route) => route.abort());
    await page.reload();
    await expect(page.getByRole("button", { name: "New chat" })).toBeVisible();
  });

  test("archived list parses backend shape", async ({ page }) => {
    await page.goto("/");
    await page.evaluate(() => {
      localStorage.setItem("echoline_token", "mock-token");
      localStorage.setItem("echoline_refresh", "mock-refresh");
      localStorage.setItem("echoline_device", "playwright-device");
    });
    await page.route("**/api/**", async (route) => {
      const url = route.request().url();
      if (url.includes("/api/conversations/archived")) {
        return route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            archived: [{
              conversation_id: "550e8400-e29b-41d4-a716-446655440099",
              type: "group",
              title: "Archived Test",
              archived_at: "2026-07-01T00:00:00Z",
            }],
          }),
        });
      }
      return route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
    });
    await page.route("**/ws**", (route) => route.abort());
    await page.reload();
    await page.getByRole("button", { name: "archived" }).click();
    await expect(page.getByText("Archived Test (archived)")).toBeVisible();
  });
});
