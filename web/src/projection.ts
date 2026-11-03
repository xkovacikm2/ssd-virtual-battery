type Num = number | string;

export interface ChartPoint {
  date: string;
  cumulative_exported: Num;
  cumulative_imported: Num;
}

export interface DailyValue {
  date: string;
  exported_to_grid: Num;
  imported_from_grid: Num;
}

export interface ChartSeries {
  labels: string[];
  exported: (number | null)[];
  exportProjection: (number | null)[];
  imported: (number | null)[];
  importProjection: (number | null)[];
  capacity: number[];
}

type Daily = { date: string; value: number };

function project(values: number[], labels: string[], maxCapacity: number, prevYear: Daily[]) {
  const projLabels: string[] = [];
  const projValues: (number | null)[] = [];
  if (values.length < 2) return { projLabels, projValues };

  const lastDate = new Date(`${labels[labels.length - 1]}T00:00:00Z`);
  const yearEnd = new Date(Date.UTC(lastDate.getUTCFullYear(), 11, 31));
  const lookup = new Map(prevYear.map((d) => [d.date.slice(5, 10), d.value]));
  const hasPrevYear = prevYear.length > 0;

  // Without last year's data, extrapolate the average daily change of the last two weeks.
  const windowSize = Math.min(14, values.length);
  const recent = values.slice(-windowSize);
  const avgDaily = (recent[recent.length - 1] - recent[0]) / (windowSize - 1);
  const prevYearAvg = hasPrevYear ? prevYear.reduce((s, d) => s + d.value, 0) / prevYear.length : avgDaily;

  let cumulative = values[values.length - 1];
  const day = new Date(lastDate);
  day.setUTCDate(day.getUTCDate() + 1);
  while (day <= yearEnd) {
    const iso = day.toISOString().slice(0, 10);
    cumulative += hasPrevYear ? (lookup.get(iso.slice(5)) ?? prevYearAvg) : avgDaily;
    projLabels.push(iso);
    projValues.push(Number.isFinite(cumulative) ? parseFloat(Math.min(cumulative, maxCapacity).toFixed(2)) : null);
    day.setUTCDate(day.getUTCDate() + 1);
  }
  return { projLabels, projValues };
}

export function buildChart(readings: ChartPoint[], maxCapacity: number, previousYear: DailyValue[]): ChartSeries {
  const labels = readings.map((r) => r.date);
  const exported = readings.map((r) => Number(r.cumulative_exported));
  const imported = readings.map((r) => Number(r.cumulative_imported));

  const exp = project(exported, labels, maxCapacity,
    previousYear.map((d) => ({ date: d.date, value: Number(d.exported_to_grid) })));
  const imp = project(imported, labels, maxCapacity,
    previousYear.map((d) => ({ date: d.date, value: Number(d.imported_from_grid) })));

  const projLabels = Array.from(new Set([...exp.projLabels, ...imp.projLabels])).sort();
  const lead = (): null[] => new Array(labels.length).fill(null);
  const align = (p: { projLabels: string[]; projValues: (number | null)[] }) =>
    projLabels.map((l) => {
      const i = p.projLabels.indexOf(l);
      return i === -1 ? null : p.projValues[i];
    });
  const pad = (a: number[]) => [...a, ...new Array<null>(projLabels.length).fill(null)];
  const allLabels = [...labels, ...projLabels];

  return {
    labels: allLabels,
    exported: pad(exported),
    exportProjection: [...lead(), ...align(exp)],
    imported: pad(imported),
    importProjection: [...lead(), ...align(imp)],
    capacity: allLabels.map(() => maxCapacity),
  };
}
