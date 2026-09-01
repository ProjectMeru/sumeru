import Chart from "chart.js/auto";
import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { onWillStart, onWillUnmount } from "../../runtime/lifecycle.js";
import { onMount, useTemplateRef } from "../../runtime/hooks.js";
import { VIEW_GRAPH, VIEW_LIST } from "../../constants/routes.js";
import { CollectionBarHost, mountCollectionBar } from "../shared/collection-bar-host.js";
import { navigateCollectionQuery } from "../shared/collection-query.js";
import { renderGraphExportLink } from "../shared/view-toolbar.js";

interface GraphViewProps {
  payload: SwcWorkspacePayload;
}

const CHART_PALETTE = ["#2563eb", "#16a34a", "#f59e0b", "#dc2626", "#7c3aed", "#0891b2"];

/** Graph view over read_group RPC (bar / line / pie) via Chart.js. */
export class GraphView extends SwcComponent<GraphViewProps> {
  private groups: Record<string, unknown>[] = [];
  private measureField = "id";
  private groupField = "create_date";
  private chart = "bar";
  private collectionBar!: CollectionBarHost;
  private chartInstance: Chart | null = null;
  private chartInstanceType: "bar" | "line" | "pie" | "" = "";
  private canvasRef!: { current: Element | null };
  private chartDataKey = "";
  private loadSeq = 0;

  override setup(): void {
    this.canvasRef = useTemplateRef("graph-canvas");
    this.collectionBar = mountCollectionBar(this.props.payload, VIEW_GRAPH, this.env);
    onWillStart(() => this.load());
    onMount(() => this.syncChart());
    onWillUnmount(() => this.destroyChart());
  }

  override afterPatch(): void {
    const key = this.buildChartDataKey();
    if (key === this.chartDataKey) {
      return;
    }
    this.chartDataKey = key;
    this.syncChart();
  }

  override onPropsChanged(props: GraphViewProps): void {
    this.collectionBar.updateProps({ payload: props.payload, viewType: VIEW_GRAPH });
    void this.load();
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
  }

  private async load(): Promise<void> {
    const seq = ++this.loadSeq;
    const payload = this.props.payload;
    this.chart = (payload.arch.graph?.chart || "bar").toLowerCase();
    this.groupField = payload.arch.fields.find((f) => f.pivotType === "row")?.name ?? "create_date";
    this.measureField = payload.arch.fields.find((f) => f.pivotType === "measure")?.name ?? "id";
    const groups = await this.env.services.rpc.readGroup(
      payload.model,
      [],
      [this.measureField],
      [this.groupField],
      40,
    );
    if (seq !== this.loadSeq) {
      return;
    }
    this.groups = groups;
    this.chartDataKey = "";
    this.rerender();
  }

  private labelOf(group: Record<string, unknown>): string {
    const nameKey = `${this.groupField}_name`;
    if (group[nameKey] != null) return String(group[nameKey]);
    if (group[this.groupField] != null) return String(group[this.groupField]);
    return String(group.name ?? "");
  }

  private chartType(): "bar" | "line" | "pie" {
    if (this.chart === "pie") return "pie";
    if (this.chart === "line") return "line";
    return "bar";
  }

  private buildChartDataKey(): string {
    const labels = this.groups.map((group) => this.labelOf(group));
    const values = this.groups.map((group) => Number(group[this.measureField] ?? 0));
    return `${this.chartType()}:${this.measureField}:${labels.join("\u001e")}:${values.join(",")}`;
  }

  private destroyChart(): void {
    this.chartInstance?.destroy();
    this.chartInstance = null;
    this.chartInstanceType = "";
  }

  private syncChart(): void {
    const canvas = this.canvasRef.current as HTMLCanvasElement | null;
    if (!canvas) return;

    const labels = this.groups.map((group) => this.labelOf(group));
    const values = this.groups.map((group) => Number(group[this.measureField] ?? 0));
    const type = this.chartType();

    if (this.chartInstance && this.chartInstanceType === type) {
      this.chartInstance.data.labels = labels;
      const dataset = this.chartInstance.data.datasets[0];
      dataset.data = values;
      dataset.label = this.measureField;
      if (type === "pie") {
        dataset.backgroundColor = CHART_PALETTE;
        dataset.borderColor = CHART_PALETTE;
      } else {
        dataset.backgroundColor = CHART_PALETTE[0];
        dataset.borderColor = type === "line" ? CHART_PALETTE[0] : CHART_PALETTE;
      }
      this.chartInstance.update();
      return;
    }

    this.destroyChart();
    this.chartInstanceType = type;
    this.chartInstance = new Chart(canvas, {
      type,
      data: {
        labels,
        datasets: [
          {
            label: this.measureField,
            data: values,
            backgroundColor: type === "pie" ? CHART_PALETTE : CHART_PALETTE[0],
            borderColor: type === "line" ? CHART_PALETTE[0] : CHART_PALETTE,
            borderWidth: type === "line" ? 2 : 1,
            tension: type === "line" ? 0.25 : 0,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        onClick: (_event, elements) => {
          if (elements.length === 0) return;
          const idx = elements[0].index;
          const group = this.groups[idx];
          if (!group) return;
          const value = group[this.groupField];
          navigateCollectionQuery(this.env, this.props.payload, VIEW_LIST, {
            customDomain: JSON.stringify([[this.groupField, "=", value]]),
            listOffset: 0,
          });
        },
        plugins: {
          legend: { display: type === "pie" },
        },
        scales:
          type === "pie"
            ? undefined
            : {
                y: { beginAtZero: true },
              },
      },
    });
  }

  override template() {
    const exportLink = renderGraphExportLink(this.props.payload, this.groupField, this.measureField);
    return html`
      <div class="sum-collection-view sum-graph-view">
        ${this.collectionBar.renderOrPatch()}
        ${exportLink ?? ""}
        <div class="sum-graph-chart-wrap">
          <canvas data-ref="graph-canvas"></canvas>
        </div>
      </div>
    `;
  }
}
