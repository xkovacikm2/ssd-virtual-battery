import { describe, expect, it } from "vitest";
import { buildChart } from "./projection";

const cum = (rows: [string, number, number][]) =>
  rows.map(([date, e, i]) => ({ date, cumulative_exported: e, cumulative_imported: i }));

describe("buildChart", () => {
  it("has no projection with fewer than two readings", () => {
    const c = buildChart(cum([["2026-12-29", 3, 1]]), 6000, []);
    expect(c.labels).toEqual(["2026-12-29"]);
    expect(c.exported).toEqual([3]);
    expect(c.exportProjection).toEqual([null]);
    expect(c.capacity).toEqual([6000]);
  });

  it("projects to Dec 31 using previous-year daily values by month-day", () => {
    const c = buildChart(
      cum([["2026-12-28", 10, 1], ["2026-12-29", 20, 2]]),
      6000,
      [
        { date: "2025-12-30", exported_to_grid: 5, imported_from_grid: 1 },
        { date: "2025-12-31", exported_to_grid: 7, imported_from_grid: 3 },
      ],
    );
    expect(c.labels).toEqual(["2026-12-28", "2026-12-29", "2026-12-30", "2026-12-31"]);
    expect(c.exported).toEqual([10, 20, null, null]);
    expect(c.imported).toEqual([1, 2, null, null]);
    expect(c.exportProjection).toEqual([null, null, 25, 32]);
    expect(c.importProjection).toEqual([null, null, 3, 6]);
    expect(c.capacity).toEqual([6000, 6000, 6000, 6000]);
  });

  it("uses the previous-year average for days missing last year", () => {
    const c = buildChart(
      cum([["2026-12-28", 10, 0], ["2026-12-29", 20, 0]]),
      6000,
      [
        { date: "2025-12-30", exported_to_grid: 4, imported_from_grid: 0 },
        { date: "2025-11-01", exported_to_grid: 8, imported_from_grid: 0 },
      ],
    );
    expect(c.exportProjection.slice(2)).toEqual([24, 30]);
  });

  it("falls back to the average daily change over the last 14 days", () => {
    const rows: [string, number, number][] = [];
    for (let d = 1; d <= 20; d++) rows.push([`2026-12-${String(d).padStart(2, "0")}`, d * d, d]);
    const c = buildChart(cum(rows), 6000, []);
    // last 14 values are 7^2..20^2, avg = (400 - 49) / 13 = 27
    expect(c.exportProjection[20]).toBe(427);
    expect(c.exportProjection[30]).toBe(400 + 27 * 11);
    expect(c.importProjection[30]).toBe(31);
  });

  it("caps the projection at max capacity", () => {
    const c = buildChart(cum([["2026-12-28", 5990, 0], ["2026-12-29", 5995, 0]]), 6000, []);
    expect(c.exportProjection.slice(2)).toEqual([6000, 6000]);
  });

  it("rounds projected values to two decimals", () => {
    const c = buildChart(
      cum([["2026-12-29", 1, 0], ["2026-12-30", 1.1, 0]]),
      6000,
      [{ date: "2025-12-31", exported_to_grid: 0.333, imported_from_grid: 0 }],
    );
    expect(c.exportProjection[2]).toBe(1.43);
  });

  it("accepts numeric strings like the Rails JSON did", () => {
    const c = buildChart([{ date: "2026-12-30", cumulative_exported: "1.5", cumulative_imported: "0.5" }], 6000, []);
    expect(c.exported).toEqual([1.5]);
  });
});
