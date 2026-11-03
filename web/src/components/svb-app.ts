import { LitElement, html, nothing } from "lit";
import { fetchDashboard, type Dashboard } from "../api";
import { shared } from "../styles";
import "./svb-year-summary";
import "./svb-burndown-chart";
import "./svb-readings";

export class SvbApp extends LitElement {
  static styles = [shared];
  static properties = { data: { state: true }, failed: { state: true } };

  data?: Dashboard;
  failed = false;

  connectedCallback() {
    super.connectedCallback();
    fetchDashboard().then(
      (data) => (this.data = data),
      (e) => {
        console.error(e);
        this.failed = true;
      },
    );
  }

  render() {
    return html`
      <div class="dashboard-container">
        <h1 class="dashboard-title">Prehľad virtuálnej batérie</h1>
        ${this.failed ? html`<div class="load-error">Nepodarilo sa načítať údaje.</div>` : nothing}
        ${this.data ? this.renderDashboard(this.data) : nothing}
      </div>`;
  }

  private renderDashboard(d: Dashboard) {
    return html`
      <svb-year-summary .summary=${d.summary} year=${d.year}></svb-year-summary>
      ${d.chart_data.length > 0
        ? html`<svb-burndown-chart .readings=${d.chart_data} .previousYear=${d.previous_year}
            maxCapacity=${d.max_capacity} year=${d.year}></svb-burndown-chart>`
        : nothing}
      ${d.has_readings
        ? html`<svb-readings year=${d.year}></svb-readings>`
        : html`
          <div class="no-data">
            <p><strong>Zatiaľ nie sú k dispozícii žiadne údaje.</strong> Údaje budú zhromažďované na pozadí.</p>
          </div>`}`;
  }
}

customElements.define("svb-app", SvbApp);

declare global {
  interface HTMLElementTagNameMap {
    "svb-app": SvbApp;
  }
}
