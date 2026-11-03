const DAY_MS = 86_400_000;

export function iso(d: Date): string {
  return d.toISOString().slice(0, 10);
}

export function parse(s: string): Date {
  return new Date(`${s}T00:00:00Z`);
}

export function addDays(s: string, n: number): string {
  return iso(new Date(parse(s).getTime() + n * DAY_MS));
}

export function today(): string {
  return iso(new Date());
}

export function yesterday(): string {
  return addDays(today(), -1);
}

export function range(from: string, to: string): string[] {
  const out: string[] = [];
  for (let d = from; d <= to; d = addDays(d, 1)) out.push(d);
  return out;
}

export function yearOf(s: string): number {
  return Number(s.slice(0, 4));
}

/** Same calendar day one year earlier (Ruby `Date - 1.year` semantics: Feb 29 -> Feb 28). */
export function minusOneYear(s: string): string {
  const [y, m, d] = s.split("-").map(Number);
  const last = new Date(Date.UTC(y - 1, m, 0)).getUTCDate();
  return `${y - 1}-${String(m).padStart(2, "0")}-${String(Math.min(d, last)).padStart(2, "0")}`;
}

export function isoWeek(s: string): number {
  const d = parse(s);
  const day = d.getUTCDay() || 7;
  d.setUTCDate(d.getUTCDate() + 4 - day);
  const yearStart = Date.UTC(d.getUTCFullYear(), 0, 1);
  return Math.ceil(((d.getTime() - yearStart) / DAY_MS + 1) / 7);
}

const DAY_NAMES = ["Nedeľa", "Pondelok", "Utorok", "Streda", "Štvrtok", "Piatok", "Sobota"];
export const MONTH_NAMES = [
  "Január", "Február", "Marec", "Apríl", "Máj", "Jún",
  "Júl", "August", "September", "Október", "November", "December",
];

export function dayName(s: string): string {
  return DAY_NAMES[parse(s).getUTCDay()];
}

/** strftime('%d. %m.') */
export function shortDate(s: string): string {
  return `${s.slice(8, 10)}. ${s.slice(5, 7)}.`;
}
