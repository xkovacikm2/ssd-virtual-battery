import type { ChartConfiguration } from "chart.js";
import type { ChartSeries } from "./projection";

export function chartConfig(series: ChartSeries, maxCapacity: number, isMobile: boolean): ChartConfiguration<"line"> {
  return {
    type: "line",
    data: {
      labels: series.labels,
      datasets: [
        {
          label: "Kumulatívny export do siete (kWh)",
          data: series.exported,
          borderColor: "#4facfe",
          backgroundColor: "rgba(252,74,26,0.07)",
          fill: false,
          tension: 0.3,
          pointRadius: 2,
          spanGaps: false,
        },
        {
          label: "Projekcia exportu (kWh)",
          data: series.exportProjection,
          borderColor: "#4facfe",
          backgroundColor: "transparent",
          borderDash: [6, 4],
          fill: false,
          tension: 0.3,
          pointRadius: 0,
          spanGaps: false,
        },
        {
          label: "Kumulatívny import zo siete (kWh)",
          data: series.imported,
          borderColor: "#fc4a1a",
          backgroundColor: "rgba(79,172,254,0.07)",
          fill: false,
          tension: 0.3,
          pointRadius: 2,
          spanGaps: false,
        },
        {
          label: "Projekcia importu (kWh)",
          data: series.importProjection,
          borderColor: "#fc4a1a",
          backgroundColor: "transparent",
          borderDash: [6, 4],
          fill: false,
          tension: 0.3,
          pointRadius: 0,
          spanGaps: false,
        },
        {
          label: "Maximálna kapacita VB (kWh)",
          data: series.capacity,
          borderColor: "#2ecc71",
          backgroundColor: "transparent",
          borderDash: [4, 4],
          fill: false,
          tension: 0,
          pointRadius: 0,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { mode: "index", intersect: false },
      plugins: {
        legend: {
          position: "bottom",
          labels: { boxWidth: isMobile ? 8 : 12, font: { size: isMobile ? 10 : 12 } },
        },
        tooltip: {
          callbacks: {
            label: (ctx) => `${ctx.dataset.label}: ${ctx.parsed.y ?? "-"} kWh`,
          },
        },
      },
      scales: {
        x: { ticks: { maxTicksLimit: isMobile ? 6 : 12, maxRotation: 45 } },
        y: { min: 0, suggestedMax: maxCapacity * 1.05, title: { display: true, text: "kWh" } },
      },
    },
  };
}
