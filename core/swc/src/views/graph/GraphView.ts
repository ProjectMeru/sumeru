import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { onWillStart } from "../../runtime/lifecycle.js";

interface GraphViewProps {
  payload: SwcWorkspacePayload;
}

/** Graph view over read_group RPC (bar / line / pie). */
export class GraphView extends SwcComponent<GraphViewProps> {
  private groups: Record<string, unknown>[] = [];
  private measureField = "id";
  private groupField = "create_date";
  private chart = "bar";

  override setup(): void {
    onWillStart(() => this.load());
  }

  private async load(): Promise<void> {
    const payload = this.props.payload;
    this.chart = (payload.arch.graph?.chart || "bar").toLowerCase();
    this.groupField = payload.arch.fields.find((f) => f.pivotType === "row")?.name ?? "create_date";
    this.measureField = payload.arch.fields.find((f) => f.pivotType === "measure")?.name ?? "id";
    this.groups = await this.env.services.rpc.readGroup(
      payload.model,
      [],
      [this.measureField],
      [this.groupField],
      40,
    );
    this.rerender();
  }

  private labelOf(group: Record<string, unknown>): string {
    const nameKey = `${this.groupField}_name`;
    if (group[nameKey] != null) return String(group[nameKey]);
    if (group[this.groupField] != null) return String(group[this.groupField]);
    return String(group.name ?? "");
  }

  override template() {
    const max = Math.max(...this.groups.map((g) => Number(g[this.measureField] ?? 0)), 1);
    if (this.chart === "pie") {
      let accumulatedPercent = 0;
      const total = this.groups.reduce((sum, g) => sum + Number(g[this.measureField] ?? 0), 0) || 1;
      const stops: string[] = [];
      const palette = ["#2563eb", "#16a34a", "#f59e0b", "#dc2626", "#7c3aed", "#0891b2"];
      this.groups.forEach((group, index) => {
        const fieldValue = Number(group[this.measureField] ?? 0);
        const start = accumulatedPercent;
        accumulatedPercent += (fieldValue / total) * 100;
        stops.push(`${palette[index % palette.length]} ${start}% ${accumulatedPercent}%`);
      });
      return html`
        <div class="sum-graph-view">
          <div class="sum-graph-pie" style=${`background:conic-gradient(${stops.join(",")})`}></div>
          <ul class="sum-graph-legend">
            ${this.groups.map(
              (group, index) => html`<li>
                <span class="sum-graph-swatch" style=${`background:${["#2563eb", "#16a34a", "#f59e0b", "#dc2626", "#7c3aed", "#0891b2"][index % 6]}`}></span>
                ${this.labelOf(group)} (${String(group[this.measureField] ?? 0)})
              </li>`,
            )}
          </ul>
        </div>
      `;
    }
    return html`
      <div class="sum-graph-view">
        ${this.groups.map((group) => {
          const label = this.labelOf(group);
          const fieldValue = Number(group[this.measureField] ?? 0);
          const pct = Math.round((fieldValue / max) * 100);
          return html`<div class="sum-graph-bar-row">
            <span class="sum-graph-label">${label}</span>
            <div class=${this.chart === "line" ? "sum-graph-bar sum-graph-bar--line" : "sum-graph-bar"} style="width:${pct}%"></div>
            <span class="sum-graph-value">${fieldValue}</span>
          </div>`;
        })}
      </div>
    `;
  }
}
