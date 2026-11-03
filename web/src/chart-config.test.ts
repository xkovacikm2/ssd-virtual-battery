import { describe, expect, it } from "vitest";
import { chartConfig } from "./chart-config";
import { buildChart } from "./projection";

describe("chartConfig", () => {
  const series = buildChart(
    [
      { date: "2026-12-29", cumulative_exported: 10, cumulative_imported: 4 },
      { date: "2026-12-30", cumulative_exported: 20, cumulative_imported: 6 },
    ],
    6000,
    [],
  );

  it("plots the five Rails series in order with matching styles", () => {
    const cfg = chartConfig(series, 6000, false);
    expect(cfg.type).toBe("line");
    expect(cfg.data.labels).toEqual(series.labels);
    expect(cfg.data.datasets.map((d) => d.label)).toEqual([
      "Kumulatívny export do siete (kWh)",
      "Projekcia exportu (kWh)",
      "Kumulatívny import zo siete (kWh)",
      "Projekcia importu (kWh)",
      "Maximálna kapacita VB (kWh)",
    ]);
    expect(cfg.data.datasets.map((d) => d.borderColor)).toEqual(["#4facfe", "#4facfe", "#fc4a1a", "#fc4a1a", "#2ecc71"]);
    expect(cfg.data.datasets.map((d) => d.borderDash)).toEqual([undefined, [6, 4], undefined, [6, 4], [4, 4]]);
    expect(cfg.data.datasets[0].data).toEqual(series.exported);
    expect(cfg.data.datasets[1].data).toEqual(series.exportProjection);
    expect(cfg.data.datasets[4].data).toEqual(series.capacity);
  });

  it("scales the y axis to 105% of capacity", () => {
    const y = chartConfig(series, 6000, false).options!.scales!.y as { min: number; suggestedMax: number };
    expect(y.min).toBe(0);
    expect(y.suggestedMax).toBe(6300);
  });

  it("uses compact legend and fewer ticks on mobile", () => {
    const desktop = chartConfig(series, 6000, false).options!;
    const mobile = chartConfig(series, 6000, true).options!;
    expect((desktop.scales!.x as any).ticks.maxTicksLimit).toBe(12);
    expect((mobile.scales!.x as any).ticks.maxTicksLimit).toBe(6);
    expect(desktop.plugins!.legend!.labels!.boxWidth).toBe(12);
    expect(mobile.plugins!.legend!.labels!.boxWidth).toBe(8);
  });
});
