import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { VIEW_COHORT } from "../../constants/routes.js";
import { CollectionBarHost, mountCollectionBar } from "../shared/collection-bar-host.js";

interface CohortViewProps {
  payload: SwcWorkspacePayload;
}

type CohortInterval = "week" | "month";

function parseDate(raw: unknown): Date | null {
  const text = String(raw ?? "").trim();
  if (!text) return null;
  const date = new Date(text);
  return Number.isNaN(date.getTime()) ? null : date;
}

function bucketKey(date: Date, interval: CohortInterval): string {
  if (interval === "month") {
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;
  }
  const tmp = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
  const day = tmp.getUTCDay() || 7;
  tmp.setUTCDate(tmp.getUTCDate() + 4 - day);
  const yearStart = new Date(Date.UTC(tmp.getUTCFullYear(), 0, 1));
  const week = Math.ceil(((tmp.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
  return `${tmp.getUTCFullYear()}-W${String(week).padStart(2, "0")}`;
}

export class CohortView extends SwcComponent<CohortViewProps> {
  private collectionBar!: CollectionBarHost;

  override setup(): void {
    this.collectionBar = mountCollectionBar(this.props.payload, VIEW_COHORT, this.env);
  }

  override onPropsChanged(props: CohortViewProps): void {
    this.collectionBar.updateProps({ payload: props.payload, viewType: VIEW_COHORT });
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
  }

  private dateField(): string {
    return this.props.payload.arch.cohort?.dateStart
      || this.props.payload.arch.calendar?.dateStart
      || this.props.payload.arch.gantt?.dateStart
      || this.props.payload.arch.fields.find((f) => f.type === "date" || f.type === "datetime")?.name
      || "create_date";
  }

  private measureField(): string {
    return this.props.payload.arch.cohort?.measure
      || this.props.payload.arch.fields.find((f) => f.pivotType === "measure")?.name
      || "";
  }

  private interval(): CohortInterval {
    const raw = (this.props.payload.arch.cohort?.interval ?? "month").toLowerCase();
    return raw === "week" ? "week" : "month";
  }

  private table(): { periods: string[]; rows: Array<{ cohort: string; values: number[] }> } {
    const dateField = this.dateField();
    const measureField = this.measureField();
    const interval = this.interval();
    const groups = new Map<string, number>();
    for (const row of this.props.payload.records ?? []) {
      const date = parseDate(row[dateField]);
      if (!date) continue;
      const key = bucketKey(date, interval);
      const amount = measureField ? Number(row[measureField] ?? 0) : 1;
      groups.set(key, (groups.get(key) ?? 0) + (Number.isFinite(amount) ? amount : 0));
    }
    const periods = [...groups.keys()].sort();
    const rows = periods.map((cohort, index) => {
      const values = periods.map((_, col) => {
        if (col < index) return 0;
        const later = periods[col];
        return groups.get(later) ?? 0;
      });
      return { cohort, values };
    });
    return { periods, rows };
  }

  override template() {
    const { periods, rows } = this.table();
    return html`
      <div class="sum-collection-view sum-cohort-view">
        ${this.collectionBar.renderOrPatch()}
        <h2>${this.props.payload.arch.title ?? "Cohort"}</h2>
        <table class="sum-cohort-table">
          <thead>
            <tr>
              <th>Cohort</th>
              ${periods.map((p) => html`<th>${p}</th>`)}
            </tr>
          </thead>
          <tbody>
            ${rows.map(
              (row) => html`<tr>
                <th>${row.cohort}</th>
                ${row.values.map((value) => html`<td>${value === 0 ? "" : String(value)}</td>`)}
              </tr>`,
            )}
          </tbody>
        </table>
      </div>
    `;
  }
}
