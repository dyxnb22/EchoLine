import { test, expect } from "@playwright/test";

test("login page renders", async ({ page }) => {
  await page.goto("/login");
  await expect(page.getByRole("heading", { name: "EchoLine" })).toBeVisible();
  await expect(page.getByPlaceholder("username")).toBeVisible();
});

test("chat layout with mocked session", async ({ page }) => {
  await page.addInitScript(() => {
    localStorage.setItem("echoline_token", "test-token");
    localStorage.setItem("echoline_refresh", "test-refresh");
    localStorage.setItem("echoline_device", "e2e-device");
  });

  await page.route("**/api/conversations", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ conversations: [] }),
    });
  });
  await page.route("**/api/recommendations/**", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
  });
  await page.route("**/api/presence/last-seen", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
  });
  await page.route("**/api/notifications", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ notifications: [] }) });
  });

  await page.goto("/");
  await expect(page.getByText("Select a conversation")).toBeVisible();
  await expect(page.getByPlaceholder("Type a message")).toHaveCount(0);
});

test("composer visible when conversation selected (mocked)", async ({ page }) => {
  const convId = "00000000-0000-4000-8000-000000000001";

  await page.addInitScript(() => {
    localStorage.setItem("echoline_token", "test-token");
    localStorage.setItem("echoline_refresh", "test-refresh");
    localStorage.setItem("echoline_device", "e2e-device");
  });

  await page.route("**/api/conversations", async (route) => {
    if (route.request().url().includes("/messages")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ messages: [], next_before: null }),
      });
      return;
    }
    if (route.request().url().includes("/pins")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ pins: [] }) });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        conversations: [{ id: convId, type: "group", title: "E2E Group", unread: 0, latest_seq: 0 }],
      }),
    });
  });
  await page.route("**/api/recommendations/**", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
  });
  await page.route("**/api/presence/last-seen", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
  });
  await page.route("**/api/notifications", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ notifications: [] }) });
  });
  await page.route("**/ws**", async (route) => route.abort());

  await page.goto("/");
  await page.getByRole("button", { name: /E2E Group/ }).click();
  await expect(page.getByPlaceholder("Type a message")).toBeVisible();
  await page.getByPlaceholder("Type a message").fill("hello e2e");
  await expect(page.getByPlaceholder("Type a message")).toHaveValue("hello e2e");
});
