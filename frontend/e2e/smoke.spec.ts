import { test, expect } from "@playwright/test";

test("login page renders", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "EchoLine" })).toBeVisible();
  await expect(page.getByPlaceholder("username")).toBeVisible();
});
