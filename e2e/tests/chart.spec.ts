import { expect, test } from "@playwright/test";
import { resetReadings } from "../support/db";
import { resetSsd } from "../support/ssd";
import { addDays, range, today, yearOf, yesterday } from "../support/dates";
import { cumulative, MAX_CAPACITY, projection, standardSeed } from "../support/model";
import { chartSnapshot, openDashboard } from "../support/page";

test.use({ serviceWorkers: "block" });

test.beforeEach(async () => {
  await resetSsd();
});

const LABELS = [
  "Kumulatívny export do siete (kWh)",
  "Projekcia exportu (kWh)",
  "Kumulatívny import zo siete (kWh)",
  "Projekcia importu (kWh)",
  "Maximálna kapacita VB (kWh)",
];

function numbers(data: (number | string | null)[]) {
  return data.map((v) => (v === null ? null : Number(v)));
}

function expectClose(actual: (number | null)[], expected: (number | null)[]) {
  expect(actual.length).toBe(expected.length);
  actual.forEach((v, i) => {
    if (expected[i] === null) expect(v, `index ${i}`).toBeNull();
    else expect(v, `index ${i}`).toBeCloseTo(expected[i] as number, 2);
  });
}

test("burndown chart plots cumulative values, projections and capacity", async ({ page }) => {
  const seed = standardSeed();
  await resetReadings(seed);
  await openDashboard(page);
  const chart = await chartSnapshot(page);

  const cum = cumulative(seed);
  const expProj = projection(seed, "exported");
  const impProj = projection(seed, "imported");
  const projDates = range(addDays(yesterday(), 1), `${yearOf(today())}-12-31`);

  expect(chart.datasets.map((d) => d.label)).toEqual(LABELS);
  expect(chart.labels).toEqual([...cum.map((c) => c.date), ...projDates]);

  const pad = new Array(projDates.length).fill(null);
  const lead = new Array(cum.length).fill(null);
  expectClose(numbers(chart.datasets[0].data), [...cum.map((c) => c.exported), ...pad]);
  expectClose(numbers(chart.datasets[2].data), [...cum.map((c) => c.imported), ...pad]);
  expectClose(numbers(chart.datasets[1].data), [...lead, ...expProj.map((p) => p.value)]);
  expectClose(numbers(chart.datasets[3].data), [...lead, ...impProj.map((p) => p.value)]);
  expect(numbers(chart.datasets[4].data)).toEqual(chart.labels.map(() => MAX_CAPACITY));
  expect(chart.datasets[1].borderDash).toEqual([6, 4]);
});

test("projection never exceeds capacity", async ({ page }) => {
  const seed = standardSeed({ exportScale: 10 });
  await resetReadings(seed);
  await openDashboard(page);
  const chart = await chartSnapshot(page);

  const proj = numbers(chart.datasets[1].data).filter((v): v is number => v !== null);
  expectClose(proj, projection(seed, "exported").map((p) => p.value));
  proj.forEach((v) => expect(v).toBeLessThanOrEqual(MAX_CAPACITY));
});

test("projection falls back to the recent average without previous-year data", async ({ page }) => {
  const seed = standardSeed().filter((r) => yearOf(r.date) === yearOf(today()));
  await resetReadings(seed);
  await openDashboard(page);
  const chart = await chartSnapshot(page);

  const proj = numbers(chart.datasets[1].data).filter((v): v is number => v !== null);
  expectClose(proj, projection(seed, "exported").map((p) => p.value));
});

test("no projection with a single reading", async ({ page }) => {
  await resetReadings([{ date: yesterday(), exported: 3, imported: 1 }]);
  await openDashboard(page);
  const chart = await chartSnapshot(page);

  expect(chart.labels).toEqual([yesterday()]);
  expect(numbers(chart.datasets[0].data)).toEqual([3]);
  expect(chart.datasets[1].data.every((v) => v === null)).toBe(true);
});
