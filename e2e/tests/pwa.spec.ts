import { expect, test } from "@playwright/test";
import { resetReadings } from "../support/db";
import { resetSsd } from "../support/ssd";
import { standardSeed } from "../support/model";
import { openDashboard } from "../support/page";

test.beforeEach(async () => {
  await resetSsd();
  await resetReadings(standardSeed());
});

test("manifest is linked and served", async ({ page, request }) => {
  await openDashboard(page);
  await expect(page.locator('link[rel="manifest"]')).toHaveAttribute("href", "/manifest.webmanifest");
  await expect(page.locator('meta[name="theme-color"]')).toHaveAttribute("content", "#0d6efd");

  const res = await request.get("/manifest.webmanifest");
  expect(res.ok()).toBe(true);
  const manifest = await res.json();
  expect(manifest).toMatchObject({
    name: "SSD Virtuálna batéria",
    short_name: "SSDBattery",
    start_url: "/",
    display: "standalone",
    theme_color: "#0d6efd",
  });
  expect((await request.get("/icon.png")).headers()["content-type"]).toContain("image/png");
});

test("service worker registers and serves the offline page", async ({ page, context }) => {
  await openDashboard(page);
  const scriptURL = await page.evaluate(async () => (await navigator.serviceWorker.ready).active?.scriptURL);
  expect(scriptURL).toMatch(/\/service-worker\.js$/);

  await page.reload();
  await context.setOffline(true);
  await page.goto("/").catch(() => {});
  await expect(page.locator("h1")).toHaveText("You're offline");
  await context.setOffline(false);
});
