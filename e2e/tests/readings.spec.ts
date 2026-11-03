import { expect, test } from "@playwright/test";
import { resetReadings } from "../support/db";
import { resetSsd } from "../support/ssd";
import { addDays, yesterday } from "../support/dates";
import { dailyRows, monthlyRows, standardSeed, weeklyRows } from "../support/model";
import { openDashboard, tableHeaders, tableRows } from "../support/page";

test.use({ serviceWorkers: "block" });

test.beforeEach(async () => {
  await resetSsd();
});

test("daily tab is active by default and shows the last 7 days", async ({ page }) => {
  const seed = standardSeed();
  await resetReadings(seed);
  await openDashboard(page);

  await expect(page.locator(".readings-tab")).toHaveText(["Denné", "Týždenné", "Mesačné"]);
  await expect(page.locator(".readings-tab--active")).toHaveText("Denné");
  expect(await tableHeaders(page)).toEqual(["Dátum", "Odoslané (kWh)", "Prijaté (kWh)", "Rozdiel (kWh)"]);
  expect(await tableRows(page)).toEqual(dailyRows(seed, 0));
});

test("daily pagination walks back and forth one week at a time", async ({ page }) => {
  const seed = standardSeed();
  await resetReadings(seed);
  await openDashboard(page);

  await expect(page.getByText("Novšie →")).toHaveCount(0);
  await page.getByText("← Staršie").click();
  await expect(page.locator(".readings-table tbody tr").first()).toContainText(dailyRows(seed, 1)[0][0]);
  expect(await tableRows(page)).toEqual(dailyRows(seed, 1));

  await page.getByText("← Staršie").click();
  await expect(page.locator(".readings-table tbody tr").first()).toContainText(dailyRows(seed, 2)[0][0]);
  expect(await tableRows(page)).toEqual(dailyRows(seed, 2));

  await page.getByText("Novšie →").click();
  await expect(page.locator(".readings-table tbody tr").first()).toContainText(dailyRows(seed, 1)[0][0]);
  expect(await tableRows(page)).toEqual(dailyRows(seed, 1));
});

test("daily view fills gaps with zeros and hides older link at the oldest page", async ({ page }) => {
  const seed = [
    { date: yesterday(), exported: 2, imported: 5.5 },
    { date: addDays(yesterday(), -3), exported: 4.25, imported: 1 },
  ];
  await resetReadings(seed);
  await openDashboard(page);

  expect(await tableRows(page)).toEqual(dailyRows(seed, 0));
  await expect(page.getByText("← Staršie")).toHaveCount(0);
  await expect(page.getByText("Novšie →")).toHaveCount(0);
});

test("weekly tab aggregates by ISO week, newest first", async ({ page }) => {
  const seed = standardSeed();
  await resetReadings(seed);
  await openDashboard(page);

  await page.locator(".readings-tab", { hasText: "Týždenné" }).click();
  await expect(page.locator(".readings-tab--active")).toHaveText("Týždenné");
  expect(await tableHeaders(page)).toEqual(["Týždeň", "Odoslané (kWh)", "Prijaté (kWh)", "Rozdiel (kWh)"]);
  expect(await tableRows(page)).toEqual(weeklyRows(seed));
  await expect(page.locator(".readings-pagination")).toHaveCount(0);
});

test("monthly tab aggregates by month, newest first", async ({ page }) => {
  const seed = standardSeed();
  await resetReadings(seed);
  await openDashboard(page);

  await page.locator(".readings-tab", { hasText: "Mesačné" }).click();
  await expect(page.locator(".readings-tab--active")).toHaveText("Mesačné");
  expect(await tableHeaders(page)).toEqual(["Mesiac", "Odoslané (kWh)", "Prijaté (kWh)", "Rozdiel (kWh)"]);
  expect(await tableRows(page)).toEqual(monthlyRows(seed));
  await expect(page.locator(".readings-pagination")).toHaveCount(0);

  await page.locator(".readings-tab", { hasText: /^Denné$/ }).click();
  await expect(page.locator(".readings-tab--active")).toHaveText("Denné");
  expect(await tableRows(page)).toEqual(dailyRows(seed, 0));
});
