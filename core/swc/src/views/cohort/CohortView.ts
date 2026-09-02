import { html } from "../../template/html.js";
import { VIEW_COHORT } from "../../constants/routes.js";
import { CollectionView } from "../shared/collection-view.js";
import { parseArchDate, resolveArchDateField } from "../shared/arch-date.js";

type CohortInterval = "week" | "month";

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

/** Cohort view — retention-style table grouped by date_start bucket. */
export class CohortView extends CollectionView {
  protected readonly collectionViewType = VIEW_COHORT;

  private dateField(): string {
    return resolveArchDateField(this.props.payload.arch, ["cohort", "calendar", "gantt"], "create_date");
  }

  private measureField(): string {
    return (
      this.props.payload.arch.cohort?.measure ||
      this.props.payload.arch.fields.find((f) => f.pivotType === "measure")?.name ||
      ""
    );
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
      const date = parseArchDate(row[dateField]);
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
