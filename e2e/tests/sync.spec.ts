import { expect, test } from "@playwright/test";
import { resetReadings, storedReadings } from "../support/db";
import { resetSsd, ssdRequests } from "../support/ssd";
import { addDays, today, yearOf, yesterday } from "../support/dates";
import { fmt, standardSeed, summary } from "../support/model";
import { openDashboard } from "../support/page";

test.use({ serviceWorkers: "block" });

const quarterHours = (incoming: number, outgoing: number, n = 4) =>
  Array.from({ length: n }, () => ({ incoming, outgoing }));

test("loading the dashboard fetches missing days from SSD and shows them", async ({ page }) => {
  const seed = standardSeed().filter((r) => r.date <= addDays(yesterday(), -3));
  await resetReadings(seed);
  await resetSsd({
    [addDays(yesterday(), -2)]: quarterHours(2, 4),
    [addDays(yesterday(), -1)]: [{ incoming: 1.5, outgoing: null }, { incoming: null, outgoing: 3 }],
    [yesterday()]: quarterHours(0.333, 1.111, 96),
  });

  await openDashboard(page);

  await expect.poll(async () => (await ssdRequests()).map((r) => r.date))
    .toEqual([addDays(yesterday(), -2), addDays(yesterday(), -1), yesterday()]);
  const requests = await ssdRequests();
  expect(requests.every((r) => r.pod === "4242")).toBe(true);
  expect(requests[0].from).toBe(`${addDays(yesterday(), -2)}T00:00:00.000Z`);
  expect(requests[0].to).toBe(`${addDays(yesterday(), -1)}T00:00:00.000Z`);

  const added = [
    { date: addDays(yesterday(), -2), exported: 4, imported: 2 },
    { date: addDays(yesterday(), -1), exported: 0.75, imported: 0.38 },
    { date: yesterday(), exported: 26.66, imported: 7.99 },
  ];
  await expect.poll(async () => (await storedReadings()).slice(-3)).toEqual(added);
  await expect(page.locator(".metric-value").first()).toHaveText(fmt(summary([...seed, ...added]).exported));
});

test("sync stops at the first day SSD has not published yet", async ({ page }) => {
  await resetReadings(standardSeed().filter((r) => r.date <= addDays(yesterday(), -3)));
  await resetSsd({ [addDays(yesterday(), -2)]: quarterHours(1, 1) }, "empty");

  await openDashboard(page);

  expect((await ssdRequests()).map((r) => r.date)).toEqual([addDays(yesterday(), -2), addDays(yesterday(), -1)]);
  const stored = await storedReadings();
  expect(stored[stored.length - 1].date).toBe(addDays(yesterday(), -2));
});

test("no SSD calls when yesterday is already stored", async ({ page }) => {
  await resetReadings(standardSeed());
  await resetSsd({}, "error");

  await openDashboard(page);
  await page.reload();

  expect(await ssdRequests()).toEqual([]);
});

test("empty database syncs from January 1st", async ({ page }) => {
  test.skip(today() === `${yearOf(today())}-01-01`, "nothing to sync on January 1st");
  await resetReadings([]);
  await resetSsd({}, "empty");

  await openDashboard(page);

  expect((await ssdRequests()).map((r) => r.date)).toEqual([`${yearOf(today())}-01-01`]);
});

test("dashboard still renders stored data when SSD fails", async ({ page }) => {
  const seed = standardSeed().filter((r) => r.date <= addDays(yesterday(), -3));
  await resetReadings(seed);
  await resetSsd({}, "error");

  await openDashboard(page);

  await expect(page.locator(".metric-value").first()).toHaveText(fmt(summary(seed).exported));
  expect((await ssdRequests()).map((r) => r.date)).toEqual([addDays(yesterday(), -2)]);
});
