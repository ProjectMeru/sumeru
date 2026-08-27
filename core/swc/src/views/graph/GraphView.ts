import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { useState, useEffect } from "../../runtime/hooks.js";

interface GraphViewProps {
  payload: SwcWorkspacePayload;
}

/** Graph view over read_group RPC (bar / line / pie). */
export class GraphView extends SwcComponent<GraphViewProps> {
  private groups: Record<string, unknown>[] = [];
  private measureField = "id";
  private groupField = "create_date";
  private chart = "bar";

  setup(): void {
    const [, bump] = useState(0);
    this.bump = () => bump((n) => n + 1);
    useEffect(() => {
      void this.load();
    });
  }

  private bump: (() => void) | null = null;

  private async load(): Promise<void> {
    const p = this.props.payload;
    this.chart = (p.arch.graph?.chart || "bar").toLowerCase();
    this.groupField = p.arch.fields.find((f) => f.pivotType === "row")?.name ?? "create_date";
    this.measureField = p.arch.fields.find((f) => f.pivotType === "measure")?.name ?? "id";
    this.groups = await this.env.services.rpc.readGroup(
      p.model,
      [],
      [this.measureField],
      [this.groupField],
      40,
    );
    this.bump?.();
  }

  private labelOf(g: Record<string, unknown>): string {
    const nameKey = `${this.groupField}_name`;
    if (g[nameKey] != null) return String(g[nameKey]);
    if (g[this.groupField] != null) return String(g[this.groupField]);
    return String(g.name ?? "");
  }

  template() {
    const max = Math.max(...this.groups.map((g) => Number(g[this.measureField] ?? 0)), 1);
    if (this.chart === "pie") {
      let acc = 0;
      const total = this.groups.reduce((s, g) => s + Number(g[this.measureField] ?? 0), 0) || 1;
      const stops: string[] = [];
      const palette = ["#2563eb", "#16a34a", "#f59e0b", "#dc2626", "#7c3aed", "#0891b2"];
      this.groups.forEach((g, i) => {
        const val = Number(g[this.measureField] ?? 0);
        const start = acc;
        acc += (val / total) * 100;
        stops.push(`${palette[i % palette.length]} ${start}% ${acc}%`);
      });
      return html`
        <div class="sum-graph-view">
          <div class="sum-graph-pie" style=${`background:conic-gradient(${stops.join(",")})`}></div>
          <ul class="sum-graph-legend">
            ${this.groups.map(
              (g, i) => html`<li>
                <span class="sum-graph-swatch" style=${`background:${["#2563eb", "#16a34a", "#f59e0b", "#dc2626", "#7c3aed", "#0891b2"][i % 6]}`}></span>
                ${this.labelOf(g)} (${String(g[this.measureField] ?? 0)})
              </li>`,
            )}
          </ul>
        </div>
      `;
    }
    return html`
      <div class="sum-graph-view">
        ${this.groups.map((g) => {
          const label = this.labelOf(g);
          const val = Number(g[this.measureField] ?? 0);
          const pct = Math.round((val / max) * 100);
          return html`<div class="sum-graph-bar-row">
            <span class="sum-graph-label">${label}</span>
            <div class=${this.chart === "line" ? "sum-graph-bar sum-graph-bar--line" : "sum-graph-bar"} style="width:${pct}%"></div>
            <span class="sum-graph-value">${val}</span>
          </div>`;
        })}
      </div>
    `;
  }
}
