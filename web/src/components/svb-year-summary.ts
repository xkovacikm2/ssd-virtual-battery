import { LitElement, html } from "lit";
import type { Summary } from "../api";
import { fmt } from "../format";
import { shared } from "../styles";

export class SvbYearSummary extends LitElement {
  static styles = [shared];
  static properties = { summary: { attribute: false }, year: { type: Number } };

  summary?: Summary;
  year = 0;

  render() {
    const s = this.summary;
    if (!s) return null;
    const card = (modifier: string, label: string, value: number) => html`
      <div class="metric-card metric-card--${modifier}">
        <div class="metric-label">${label}</div>
        <div class="metric-value">${fmt(value)}</div>
        <div class="metric-unit">kWh</div>
      </div>`;
    return html`
      <div class="year-summary">
        <h2 class="section-title">Ročný súhrn (${this.year})</h2>
        <div class="metrics-grid">
          ${card("current-charge", "Kumulatívne odoslané do siete (aktuálny rok)", s.total_exported_to_grid)}
          ${card("exported", "Kumulatívne prijaté zo siete (aktuálny rok)", s.total_imported_from_grid)}
          ${card("imported-grid", "Kumulatívne saldo siete (aktuálny rok)", s.net_grid_flow)}
        </div>
      </div>`;
  }
}

customElements.define("svb-year-summary", SvbYearSummary);
