import { expect, type Page } from "@playwright/test";

export async function openDashboard(page: Page) {
  await page.goto("/");
  await expect(page.locator(".dashboard-title")).toHaveText("Prehľad virtuálnej batérie");
}

export async function tableRows(page: Page): Promise<string[][]> {
  const rows = page.locator(".readings-table tbody tr");
  await expect(rows.first()).toBeVisible();
  return rows.evaluateAll((trs) =>
    trs.map((tr) => Array.from(tr.querySelectorAll("td")).map((td) => (td.textContent ?? "").replace(/\s+/g, " ").trim())),
  );
}

export async function tableHeaders(page: Page): Promise<string[]> {
  return (await page.locator(".readings-table th").allTextContents()).map((t) => t.trim());
}

export type ChartSnapshot = {
  labels: string[];
  datasets: { label: string; data: (number | string | null)[]; borderDash?: number[] }[];
};

export async function chartSnapshot(page: Page): Promise<ChartSnapshot> {
  const canvas = page.locator(".burndown-chart canvas");
  await expect.poll(() => canvas.evaluate((c) => Boolean((c as any).chart))).toBe(true);
  return canvas.evaluate((c) => {
    const chart = (c as any).chart;
    return {
      labels: [...chart.data.labels],
      datasets: chart.data.datasets.map((d: any) => ({ label: d.label, data: [...d.data], borderDash: d.borderDash })),
    };
  });
}
