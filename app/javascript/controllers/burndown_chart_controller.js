import { Controller } from "@hotwired/stimulus"
import Chart from "chart.js/auto"

export default class extends Controller {
  static targets = ["canvas"]
  static values  = {
    readings:    Array,
    maxCapacity: Number
  }

  connect() {
    this.#renderChart()
  }

  disconnect() {
    this.chart?.destroy()
  }

  // private

  #renderChart() {
    const isMobile    = window.innerWidth < 600
    const raw         = this.readingsValue
    const maxCapacity = this.maxCapacityValue
    const labels      = raw.map(d => d.date)
    const exportData  = raw.map(d => parseFloat(d.cumulative_exported))
    const importData  = raw.map(d => parseFloat(d.cumulative_imported))

    const { projLabels: exportProjLabels, projValues: exportProjValues } =
      this.#projectSeries(exportData, labels, maxCapacity)
    const { projLabels: importProjLabels, projValues: importProjValues } =
      this.#projectSeries(importData, labels, maxCapacity)

    const projLabelSet = Array.from(new Set([...exportProjLabels, ...importProjLabels])).sort()
    const allLabels    = labels.concat(projLabelSet)
    const projCount    = projLabelSet.length
    const pad          = arr => arr.concat(new Array(projCount).fill(null))

    const alignProj = (pLabels, pValues) =>
      projLabelSet.map(lbl => {
        const idx = pLabels.indexOf(lbl)
        return idx !== -1 ? pValues[idx] : null
      })

    const exportProjAligned = new Array(labels.length).fill(null).concat(alignProj(exportProjLabels, exportProjValues))
    const importProjAligned = new Array(labels.length).fill(null).concat(alignProj(importProjLabels, importProjValues))
    const capLine           = allLabels.map(() => maxCapacity)

    this.chart = new Chart(this.canvasTarget, {
      type: "line",
      data: {
        labels: allLabels,
        datasets: [
          {
            label: "Kumulatívny export do siete (kWh)",
            data: pad(exportData),
            borderColor: "#4facfe",
            backgroundColor: "rgba(252,74,26,0.07)",
            fill: false,
            tension: 0.3,
            pointRadius: 2,
            spanGaps: false
          },
          {
            label: "Projekcia exportu (kWh)",
            data: exportProjAligned,
            borderColor: "#4facfe",
            backgroundColor: "transparent",
            borderDash: [6, 4],
            fill: false,
            tension: 0.3,
            pointRadius: 0,
            spanGaps: false
          },
          {
            label: "Kumulatívny import zo siete (kWh)",
            data: pad(importData),
            borderColor: "#fc4a1a",
            backgroundColor: "rgba(79,172,254,0.07)",
            fill: false,
            tension: 0.3,
            pointRadius: 2,
            spanGaps: false
          },
          {
            label: "Projekcia importu (kWh)",
            data: importProjAligned,
            borderColor: "#fc4a1a",
            backgroundColor: "transparent",
            borderDash: [6, 4],
            fill: false,
            tension: 0.3,
            pointRadius: 0,
            spanGaps: false
          },
          {
            label: "Maximálna kapacita VB (kWh)",
            data: capLine,
            borderColor: "#2ecc71",
            backgroundColor: "transparent",
            borderDash: [4, 4],
            fill: false,
            tension: 0,
            pointRadius: 0
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: "index", intersect: false },
        plugins: {
          legend: {
            position: "bottom",
            labels: {
              boxWidth: isMobile ? 8 : 12,
              font: { size: isMobile ? 10 : 12 }
            }
          },
          tooltip: {
            callbacks: {
              label: ctx => `${ctx.dataset.label}: ${ctx.parsed.y ?? "-"} kWh`
            }
          }
        },
        scales: {
          x: { ticks: { maxTicksLimit: isMobile ? 6 : 12, maxRotation: 45 } },
          y: {
            min: 0,
            suggestedMax: maxCapacity * 1.05,
            title: { display: true, text: "kWh" }
          }
        }
      }
    })
  }

  #projectSeries(values, labels, maxCapacity) {
    if (values.length < 2) return { projLabels: [], projValues: [] }

    const windowSize = Math.min(14, values.length)
    const recent     = values.slice(-windowSize)
    const avgDaily   = (recent[recent.length - 1] - recent[0]) / (windowSize - 1)
    const lastDate   = new Date(labels[labels.length - 1])
    const lastValue  = parseFloat(values[values.length - 1])
    const yearEnd    = new Date(lastDate.getFullYear(), 11, 31)
    const projLabels = []
    const projValues = []

    let d    = new Date(lastDate)
    let step = 0
    d.setDate(d.getDate() + 1)

    while (d <= yearEnd) {
      step++
      projLabels.push(d.toISOString().slice(0, 10))
      const val = lastValue + avgDaily * step
      projValues.push(isFinite(val) ? parseFloat(Math.min(val, maxCapacity).toFixed(2)) : null)
      d.setDate(d.getDate() + 1)
    }

    return { projLabels, projValues }
  }
}
