const DAY_NAMES = ["Nedeľa", "Pondelok", "Utorok", "Streda", "Štvrtok", "Piatok", "Sobota"];
const MONTH_NAMES = [
  "Január", "Február", "Marec", "Apríl", "Máj", "Jún",
  "Júl", "August", "September", "Október", "November", "December",
];

export function fmt(value: number): string {
  return value.toFixed(2);
}

export function dayName(isoDate: string): string {
  return DAY_NAMES[new Date(`${isoDate}T00:00:00Z`).getUTCDay()];
}

export function shortDate(isoDate: string): string {
  return `${isoDate.slice(8, 10)}. ${isoDate.slice(5, 7)}.`;
}

export function monthName(month: number): string {
  return MONTH_NAMES[month - 1];
}
