import { LitElement, html, nothing } from "lit";
import { fetchReadings, type ReadingsPage, type Tab } from "../api";
import { dayName, fmt, monthName, shortDate } from "../format";
import { shared } from "../styles";

const TABS: [Tab, string][] = [["daily", "Denné"], ["weekly", "Týždenné"], ["monthly", "Mesačné"]];

export class SvbReadings extends LitElement {
  static styles = [shared];
  static properties = { year: { type: Number }, tab: { state: true }, page: { state: true }, data: { state: true } };

  year = 0;
  tab: Tab = "daily";
  page = 0;
  data?: ReadingsPage;

  connectedCallback() {
    super.connectedCallback();
    void this.load(this.tab, this.page);
  }

  private async load(tab: Tab, page: number) {
    try {
      const data = await fetchReadings(tab, page);
      this.tab = data.tab;
      this.page = data.page;
      this.data = data;
    } catch (e) {
      console.error(e);
    }
  }

  private go(e: Event, tab: Tab, page: number) {
    e.preventDefault();
    void this.load(tab, page);
  }

  render() {
    return html`
      <div class="readings-section">
        <h2 class="section-title">Záznamy (${this.year})</h2>
        <div class="readings-tabs">
          ${TABS.map(([tab, label]) => html`
            <a class="readings-tab ${tab === this.tab ? "readings-tab--active" : ""}"
               href="?tab=${tab}" @click=${(e: Event) => this.go(e, tab, 0)}>${label}</a>`)}
        </div>
        ${this.renderTable()}
      </div>`;
  }

  private renderTable() {
    const d = this.data;
    if (!d) return nothing;
    const amounts = (r: { exported: number; imported: number; difference: number }) => html`
      <td class="text-right">${fmt(r.exported)}</td>
      <td class="text-right">${fmt(r.imported)}</td>
      <td class="text-right">${fmt(r.difference)}</td>`;
    const table = (first: string, rows: unknown) => html`
      <table class="readings-table">
        <thead>
          <tr>
            <th>${first}</th>
            <th class="text-right">Odoslané (kWh)</th>
            <th class="text-right">Prijaté (kWh)</th>
            <th class="text-right">Rozdiel (kWh)</th>
          </tr>
        </thead>
        <tbody>${rows}</tbody>
      </table>`;

    switch (d.tab) {
      case "weekly":
        return table("Týždeň", (d.weekly ?? []).map((r) => html`
          <tr><td>${shortDate(r.start_date)} – ${shortDate(r.end_date)}</td>${amounts(r)}</tr>`));
      case "monthly":
        return table("Mesiac", (d.monthly ?? []).map((r) => html`
          <tr><td>${monthName(r.month)}</td>${amounts(r)}</tr>`));
      default:
        return html`
          ${table("Dátum", (d.daily ?? []).map((r) => html`
            <tr><td>${dayName(r.date)} ${shortDate(r.date)}</td>${amounts(r)}</tr>`))}
          <div class="readings-pagination">
            ${d.has_older ? html`<a class="pagination-link" href="?tab=daily&page=${d.page + 1}"
                @click=${(e: Event) => this.go(e, "daily", d.page + 1)}>← Staršie</a>` : nothing}
            <div class="pagination-spacer"></div>
            ${d.has_newer ? html`<a class="pagination-link" href="?tab=daily&page=${d.page - 1}"
                @click=${(e: Event) => this.go(e, "daily", d.page - 1)}>Novšie →</a>` : nothing}
          </div>`;
    }
  }
}

customElements.define("svb-readings", SvbReadings);
