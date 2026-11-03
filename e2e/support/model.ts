import type { Reading } from "./db";
import {
  addDays, dayName, isoWeek, minusOneYear, MONTH_NAMES, range, shortDate, today, yearOf, yesterday,
} from "./dates";

export const MAX_CAPACITY = 6000;

const round2 = (n: number) => Math.round(n * 100) / 100;
export const fmt = (n: number) => round2(n).toFixed(2);

/** Every day of the current year up to yesterday plus the whole previous year. */
export function standardSeed(opts: { exportScale?: number } = {}): Reading[] {
  const scale = opts.exportScale ?? 1;
  const year = yearOf(today());
  const current = range(`${year}-01-01`, yesterday()).map((date, n) => ({
    date,
    exported: round2((((n % 7) + 1) * 1.25) * scale),
    imported: ((n % 5) + 1) * 0.5,
  }));
  const previous = range(`${year - 1}-01-01`, `${year - 1}-12-31`).map((date) => ({
    date,
    exported: 1 + (Number(date.slice(8, 10)) % 4) * 0.5,
    imported: 0.75,
  }));
  return [...previous, ...current];
}

export function currentYear(readings: Reading[]): Reading[] {
  const year = yearOf(today());
  return readings.filter((r) => yearOf(r.date) === year).sort((a, b) => a.date.localeCompare(b.date));
}

export function summary(readings: Reading[]) {
  const cy = currentYear(readings);
  const exported = cy.reduce((s, r) => s + r.exported, 0);
  const imported = cy.reduce((s, r) => s + r.imported, 0);
  return { exported, imported, net: Math.min(MAX_CAPACITY, exported) - imported };
}

export function cumulative(readings: Reading[]) {
  let e = 0;
  let i = 0;
  return currentYear(readings).map((r) => {
    e += r.exported;
    i += r.imported;
    return { date: r.date, exported: round2(e), imported: round2(i) };
  });
}

function previousYearWindow(readings: Reading[]): Reading[] {
  const from = minusOneYear(today());
  const to = `${yearOf(today()) - 1}-12-31`;
  return readings.filter((r) => r.date >= from && r.date <= to).sort((a, b) => a.date.localeCompare(b.date));
}

/** Independent re-statement of the burndown projection rules. */
export function projection(readings: Reading[], key: "exported" | "imported") {
  const cum = cumulative(readings);
  if (cum.length < 2) return [] as { date: string; value: number }[];

  const prev = previousYearWindow(readings);
  const lookup = new Map(prev.map((r) => [r.date.slice(5), r[key]]));
  const series = cum.map((c) => c[key]);
  const window = Math.min(14, series.length);
  const recent = series.slice(-window);
  const avgDaily = (recent[recent.length - 1] - recent[0]) / (window - 1);
  const prevAvg = prev.length ? prev.reduce((s, r) => s + r[key], 0) / prev.length : avgDaily;

  const last = cum[cum.length - 1];
  let value = last[key];
  const out: { date: string; value: number }[] = [];
  for (let d = addDays(last.date, 1); d <= `${yearOf(last.date)}-12-31`; d = addDays(d, 1)) {
    value += prev.length ? (lookup.get(d.slice(5)) ?? prevAvg) : avgDaily;
    out.push({ date: d, value: round2(Math.min(value, MAX_CAPACITY)) });
  }
  return out;
}

type Row = [string, string, string, string];

export function dailyRows(readings: Reading[], page: number): Row[] {
  const byDate = new Map(readings.map((r) => [r.date, r]));
  const end = addDays(yesterday(), -7 * page);
  return range(addDays(end, -6), end).reverse().map((date) => {
    const r = byDate.get(date) ?? { exported: 0, imported: 0 };
    return [`${dayName(date)} ${shortDate(date)}`, fmt(r.exported), fmt(r.imported), fmt(r.exported - r.imported)];
  });
}

export function weeklyRows(readings: Reading[]): Row[] {
  const groups = new Map<number, Reading[]>();
  for (const r of currentYear(readings)) {
    const w = isoWeek(r.date);
    groups.set(w, [...(groups.get(w) ?? []), r]);
  }
  return [...groups.values()].reverse().map((rs) => {
    const e = rs.reduce((s, r) => s + r.exported, 0);
    const i = rs.reduce((s, r) => s + r.imported, 0);
    return [`${shortDate(rs[0].date)} – ${shortDate(rs[rs.length - 1].date)}`, fmt(e), fmt(i), fmt(e - i)];
  });
}

export function monthlyRows(readings: Reading[]): Row[] {
  const groups = new Map<number, Reading[]>();
  for (const r of currentYear(readings)) {
    const m = Number(r.date.slice(5, 7));
    groups.set(m, [...(groups.get(m) ?? []), r]);
  }
  return [...groups.entries()].reverse().map(([m, rs]) => {
    const e = rs.reduce((s, r) => s + r.exported, 0);
    const i = rs.reduce((s, r) => s + r.imported, 0);
    return [MONTH_NAMES[m - 1], fmt(e), fmt(i), fmt(e - i)];
  });
}
