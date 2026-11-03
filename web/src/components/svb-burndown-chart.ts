import { LitElement, html } from "lit";
import Chart from "chart.js/auto";
import { chartConfig } from "../chart-config";
import { buildChart, type ChartPoint, type DailyValue } from "../projection";
import { shared } from "../styles";

export class SvbBurndownChart extends LitElement {
  static styles = [shared];
  static properties = {
    readings: { attribute: false },
    previousYear: { attribute: false },
    maxCapacity: { type: Number },
    year: { type: Number },
  };

  readings: ChartPoint[] = [];
  previousYear: DailyValue[] = [];
  maxCapacity = 6000;
  year = 0;
  private chart?: Chart<"line">;

  render() {
    return html`
      <div class="burndown-chart">
        <h2 class="section-title">Stav virtuálnej batérie (${this.year})</h2>
        <div class="chart-wrapper"><canvas></canvas></div>
      </div>`;
  }

  updated() {
    const canvas = this.renderRoot.querySelector("canvas");
    if (!canvas) return;
    this.chart?.destroy();
    const series = buildChart(this.readings, this.maxCapacity, this.previousYear);
    try {
      this.chart = new Chart(canvas, chartConfig(series, this.maxCapacity, window.innerWidth < 600));
    } catch (e) {
      console.error("Failed to render chart", e);
      return;
    }
    // Exposed for the e2e parity suite.
    (canvas as HTMLCanvasElement & { chart?: Chart }).chart = this.chart;
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this.chart?.destroy();
    this.chart = undefined;
  }
}

customElements.define("svb-burndown-chart", SvbBurndownChart);
