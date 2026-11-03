import { expect, test } from "@playwright/test";
import { resetReadings } from "../support/db";
import { resetSsd } from "../support/ssd";
import { today, yearOf } from "../support/dates";
import { fmt, standardSeed, summary } from "../support/model";
import { openDashboard } from "../support/page";

test.use({ serviceWorkers: "block" });

test.beforeEach(async () => {
  await resetSsd();
});

test("page shell and title", async ({ page }) => {
  await resetReadings(standardSeed());
  await openDashboard(page);

  await expect(page).toHaveTitle("SSD Virtuálna batéria");
  const year = yearOf(today());
  await expect(page.locator(".section-title")).toHaveText([
    `Ročný súhrn (${year})`,
    `Stav virtuálnej batérie (${year})`,
    `Záznamy (${year})`,
  ]);
});

test("year summary shows current-year totals and net flow", async ({ page }) => {
  const seed = standardSeed();
  await resetReadings(seed);
  await openDashboard(page);

  const s = summary(seed);
  await expect(page.locator(".metric-label")).toHaveText([
    "Kumulatívne odoslané do siete (aktuálny rok)",
    "Kumulatívne prijaté zo siete (aktuálny rok)",
    "Kumulatívne saldo siete (aktuálny rok)",
  ]);
  await expect(page.locator(".metric-value")).toHaveText([fmt(s.exported), fmt(s.imported), fmt(s.net)]);
  await expect(page.locator(".metric-unit")).toHaveText(["kWh", "kWh", "kWh"]);
});

test("net flow caps exported energy at the 6000 kWh battery capacity", async ({ page }) => {
  const seed = standardSeed({ exportScale: 10 });
  const s = summary(seed);
  test.skip(s.exported <= 6000, "not enough days this year to exceed capacity");
  await resetReadings(seed);
  await openDashboard(page);

  await expect(page.locator(".metric-value").nth(0)).toHaveText(fmt(s.exported));
  await expect(page.locator(".metric-value").nth(2)).toHaveText(fmt(6000 - s.imported));
});

test("empty database shows zeros, the no-data notice and no chart", async ({ page }) => {
  await resetReadings([]);
  await openDashboard(page);

  await expect(page.locator(".metric-value")).toHaveText(["0.00", "0.00", "0.00"]);
  await expect(page.locator(".no-data")).toContainText("Zatiaľ nie sú k dispozícii žiadne údaje.");
  await expect(page.locator(".burndown-chart")).toHaveCount(0);
  await expect(page.locator(".readings-table")).toHaveCount(0);
});
