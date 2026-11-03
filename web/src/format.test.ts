import { describe, expect, it } from "vitest";
import { dayName, fmt, monthName, shortDate } from "./format";

describe("fmt", () => {
  it("always shows two decimals like number_with_precision", () => {
    expect(fmt(0)).toBe("0.00");
    expect(fmt(12.3)).toBe("12.30");
    expect(fmt(1523.75)).toBe("1523.75");
    expect(fmt(-3.5)).toBe("-3.50");
    expect(fmt(6000)).toBe("6000.00");
  });
});

describe("Slovak dates", () => {
  it("names weekdays", () => {
    expect(dayName("2026-11-02")).toBe("Pondelok");
    expect(dayName("2026-11-01")).toBe("Nedeľa");
    expect(dayName("2026-11-05")).toBe("Štvrtok");
  });

  it("formats day and month like strftime('%d. %m.')", () => {
    expect(shortDate("2026-03-05")).toBe("05. 03.");
  });

  it("names months", () => {
    expect(monthName(1)).toBe("Január");
    expect(monthName(3)).toBe("Marec");
    expect(monthName(12)).toBe("December");
  });
});
