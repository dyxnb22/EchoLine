import { test, expect } from "@playwright/test";

/**
 * Extended smoke: verifies composer interaction (still mocked API).
 * Full stack integration requires `make dev-up` — see e2e/integration.spec.ts placeholder.
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
      if (url.includes("/api/me")) {
        return route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            id: "550e8400-e29b-41d4-a716-446655440099",
            username: "mock-user",
            display_name: "Mock User",
          }),
        });
      }
      if (url.includes("/api/auth/refresh")) {
        return route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({ access_token: "mock-token", refresh_token: "mock-refresh" }),
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
});
