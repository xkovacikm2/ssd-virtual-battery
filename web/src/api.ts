import type { ChartPoint, DailyValue } from "./projection";

export interface Summary {
  total_exported_to_grid: number;
  total_imported_from_grid: number;
  net_grid_flow: number;
}

export interface Dashboard {
  year: number;
  max_capacity: number;
  has_readings: boolean;
  summary: Summary;
  chart_data: ChartPoint[];
  previous_year: DailyValue[];
}

export type Tab = "daily" | "weekly" | "monthly";

interface Amounts {
  exported: number;
  imported: number;
  difference: number;
}

export interface ReadingsPage {
  tab: Tab;
  page: number;
  has_newer: boolean;
  has_older: boolean;
  daily?: (Amounts & { date: string })[];
  weekly?: (Amounts & { start_date: string; end_date: string })[];
  monthly?: (Amounts & { month: number })[];
}

async function getJSON<T>(url: string): Promise<T> {
  const res = await fetch(url, { headers: { Accept: "application/json" } });
  if (!res.ok) throw new Error(`${url}: HTTP ${res.status}`);
  return res.json() as Promise<T>;
}

export const fetchDashboard = () => getJSON<Dashboard>("/api/dashboard");

export const fetchReadings = (tab: Tab, page: number) =>
  getJSON<ReadingsPage>(`/api/readings?${new URLSearchParams({ tab, page: String(page) })}`);
