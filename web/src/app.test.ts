import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import "./components/svb-app";
import type { SvbApp } from "./components/svb-app";

type Json = Record<string, unknown>;

function mockApi(routes: Record<string, Json | ((url: URL) => Json)>) {
  const calls: string[] = [];
  vi.stubGlobal("fetch", vi.fn(async (input: string) => {
    const url = new URL(input, "http://localhost");
    calls.push(url.pathname + url.search);
    const route = routes[url.pathname];
    if (!route) return new Response("not found", { status: 404 });
    const body = typeof route === "function" ? route(url) : route;
    return new Response(JSON.stringify(body), { headers: { "Content-Type": "application/json" } });
  }));
  return calls;
}

const dashboard = (over: Json = {}) => ({
  year: 2026,
  max_capacity: 6000,
  has_readings: true,
  summary: { total_exported_to_grid: 1523.75, total_imported_from_grid: 458, net_grid_flow: 1065.75 },
  chart_data: [
    { date: "2026-11-01", cumulative_exported: 10, cumulative_imported: 4 },
    { date: "2026-11-02", cumulative_exported: 30, cumulative_imported: 10 },
  ],
  previous_year: [],
  ...over,
});

const daily = (page: number, over: Json = {}) => ({
  tab: "daily", page, has_newer: page > 0, has_older: true,
  daily: [
    { date: "2026-11-02", exported: 25.5, imported: 8.75, difference: 16.75 },
    { date: "2026-11-01", exported: 2, imported: 5.5, difference: -3.5 },
  ],
  ...over,
});

const settle = async (root: ParentNode = document) => {
  for (let i = 0; i < 10; i++) {
    await new Promise((r) => setTimeout(r, 0));
    const pending = deepAll(root, "*").filter((e) => "updateComplete" in e);
    await Promise.all(pending.map((e) => (e as unknown as { updateComplete: Promise<unknown> }).updateComplete));
  }
};

/** querySelectorAll that descends into open shadow roots, in tree order. */
function deepAll(root: ParentNode, selector: string): Element[] {
  const out: Element[] = [];
  const walk = (node: ParentNode) => {
    for (const el of Array.from(node.querySelectorAll("*"))) {
      if (el.matches(selector)) out.push(el);
      if (el.shadowRoot) walk(el.shadowRoot);
    }
  };
  walk(root);
  return out;
}

const texts = (selector: string) => deepAll(document, selector).map((e) => (e.textContent ?? "").replace(/\s+/g, " ").trim());

async function mount() {
  const el = document.createElement("svb-app") as SvbApp;
  document.body.append(el);
  await settle();
  return el;
}

beforeEach(() => {
  document.body.innerHTML = "";
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("svb-app", () => {
  it("renders title, section titles and the year summary", async () => {
    mockApi({ "/api/dashboard": dashboard(), "/api/readings": daily(0) });
    await mount();

    expect(texts(".dashboard-title")).toEqual(["Prehľad virtuálnej batérie"]);
    expect(texts(".section-title")).toEqual([
      "Ročný súhrn (2026)", "Stav virtuálnej batérie (2026)", "Záznamy (2026)",
    ]);
    expect(texts(".metric-label")).toEqual([
      "Kumulatívne odoslané do siete (aktuálny rok)",
      "Kumulatívne prijaté zo siete (aktuálny rok)",
      "Kumulatívne saldo siete (aktuálny rok)",
    ]);
    expect(texts(".metric-value")).toEqual(["1523.75", "458.00", "1065.75"]);
    expect(texts(".metric-unit")).toEqual(["kWh", "kWh", "kWh"]);
    expect(deepAll(document, ".burndown-chart canvas")).toHaveLength(1);
  });

  it("shows the no-data notice without chart or readings", async () => {
    const calls = mockApi({
      "/api/dashboard": dashboard({
        has_readings: false, chart_data: [],
        summary: { total_exported_to_grid: 0, total_imported_from_grid: 0, net_grid_flow: 0 },
      }),
    });
    await mount();

    expect(texts(".metric-value")).toEqual(["0.00", "0.00", "0.00"]);
    expect(texts(".no-data")[0]).toContain("Zatiaľ nie sú k dispozícii žiadne údaje.");
    expect(deepAll(document, ".burndown-chart")).toHaveLength(0);
    expect(deepAll(document, ".readings-table")).toHaveLength(0);
    expect(calls).toEqual(["/api/dashboard"]);
  });

  it("hides the chart when the current year has no readings", async () => {
    mockApi({ "/api/dashboard": dashboard({ chart_data: [] }), "/api/readings": daily(0) });
    await mount();
    expect(deepAll(document, ".burndown-chart")).toHaveLength(0);
    expect(deepAll(document, ".readings-table")).toHaveLength(1);
  });

  it("shows an error when the dashboard cannot load", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response("boom", { status: 500 })));
    await mount();
    expect(texts(".load-error")).toEqual(["Nepodarilo sa načítať údaje."]);
  });
});

describe("svb-readings", () => {
  it("defaults to the daily tab with Slovak day labels and formatted values", async () => {
    const calls = mockApi({ "/api/dashboard": dashboard(), "/api/readings": daily(0) });
    await mount();

    expect(calls).toContain("/api/readings?tab=daily&page=0");
    expect(texts(".readings-tab")).toEqual(["Denné", "Týždenné", "Mesačné"]);
    expect(texts(".readings-tab--active")).toEqual(["Denné"]);
    expect(texts(".readings-table th")).toEqual(["Dátum", "Odoslané (kWh)", "Prijaté (kWh)", "Rozdiel (kWh)"]);
    expect(deepAll(document, ".readings-table tbody tr").map((tr) =>
      Array.from(tr.querySelectorAll("td")).map((td) => td.textContent!.trim()))).toEqual([
      ["Pondelok 02. 11.", "25.50", "8.75", "16.75"],
      ["Nedeľa 01. 11.", "2.00", "5.50", "-3.50"],
    ]);
  });

  it("pages older and newer through the daily view", async () => {
    const calls = mockApi({
      "/api/dashboard": dashboard(),
      "/api/readings": (url) => daily(Number(url.searchParams.get("page"))),
    });
    await mount();

    expect(texts(".pagination-link")).toEqual(["← Staršie"]);
    (deepAll(document, ".pagination-link")[0] as HTMLElement).click();
    await settle();
    expect(calls).toContain("/api/readings?tab=daily&page=1");
    expect(texts(".pagination-link")).toEqual(["← Staršie", "Novšie →"]);

    (deepAll(document, ".pagination-link")[1] as HTMLElement).click();
    await settle();
    expect(calls.at(-1)).toBe("/api/readings?tab=daily&page=0");
  });

  it("hides the older link when there is no older data", async () => {
    mockApi({ "/api/dashboard": dashboard(), "/api/readings": daily(0, { has_older: false }) });
    await mount();
    expect(deepAll(document, ".pagination-link")).toHaveLength(0);
  });

  it("switches to weekly and monthly tabs without pagination", async () => {
    const calls = mockApi({
      "/api/dashboard": dashboard(),
      "/api/readings": (url) => {
        const tab = url.searchParams.get("tab");
        if (tab === "weekly") return { tab, page: 0, has_newer: false, has_older: false,
          weekly: [{ start_date: "2026-10-26", end_date: "2026-10-29", exported: 25, imported: 8, difference: 17 }] };
        if (tab === "monthly") return { tab, page: 0, has_newer: false, has_older: false,
          monthly: [{ month: 3, exported: 22, imported: 7, difference: 15 }] };
        return daily(0);
      },
    });
    await mount();

    const tab = (name: string) => deepAll(document, ".readings-tab").find((t) => t.textContent!.trim() === name) as HTMLElement;
    tab("Týždenné").click();
    await settle();
    expect(calls).toContain("/api/readings?tab=weekly&page=0");
    expect(texts(".readings-tab--active")).toEqual(["Týždenné"]);
    expect(texts(".readings-table th")[0]).toBe("Týždeň");
    expect(texts(".readings-table tbody td")).toEqual(["26. 10. – 29. 10.", "25.00", "8.00", "17.00"]);
    expect(deepAll(document, ".readings-pagination")).toHaveLength(0);

    tab("Mesačné").click();
    await settle();
    expect(texts(".readings-table th")[0]).toBe("Mesiac");
    expect(texts(".readings-table tbody td")).toEqual(["Marec", "22.00", "7.00", "15.00"]);
    expect(deepAll(document, ".readings-pagination")).toHaveLength(0);
  });
});
