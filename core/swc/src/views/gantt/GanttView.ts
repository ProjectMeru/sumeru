import { html } from "../../template/html.js";
import { forEach } from "../../template/helpers.js";
import { VIEW_GANTT } from "../../constants/routes.js";
import { CollectionView } from "../shared/collection-view.js";
import { parseArchDate, resolveArchDateField } from "../shared/arch-date.js";
import { openWorkspaceRecord } from "../shared/collection-navigation.js";
import { recordDisplayLabel } from "../shared/field-display.js";

type GanttScale = "day" | "week" | "month";

/** Gantt view — horizontal timeline bars from date_start / date_stop arch fields. */
export class GanttView extends CollectionView {
  protected readonly collectionViewType = VIEW_GANTT;
  private scale: GanttScale = "week";

  private setScale(next: GanttScale): void {
    this.scale = next;
    this.rerender();
  }

  private dateStartField(): string {
    return (
      this.props.payload.arch.gantt?.dateStart ||
      resolveArchDateField(this.props.payload.arch, ["gantt", "calendar"], "date_start")
    );
  }

  private dateStopField(): string {
    return (
      this.props.payload.arch.gantt?.dateStop ||
      this.props.payload.arch.calendar?.dateStop ||
      this.dateStartField()
    );
  }

  private openRecord(row: Record<string, unknown>): void {
    openWorkspaceRecord(this.env, this.props.payload, row);
  }

  private range(): { start: number; end: number } {
    const startField = this.dateStartField();
    const stopField = this.dateStopField();
    let min = Infinity;
    let max = -Infinity;
    for (const row of this.props.payload.records ?? []) {
      const start = parseArchDate(row[startField])?.getTime();
      const stop = parseArchDate(row[stopField])?.getTime() ?? start;
      if (start == null) continue;
      min = Math.min(min, start);
      max = Math.max(max, stop ?? start);
    }
    if (!Number.isFinite(min)) {
      const now = Date.now();
      return { start: now, end: now + 86400000 * 7 };
    }
    const pad = this.scale === "day" ? 86400000 : this.scale === "week" ? 86400000 * 7 : 86400000 * 30;
    return { start: min - pad, end: max + pad };
  }

  override template() {
    const startField = this.dateStartField();
    const stopField = this.dateStopField();
    const { start, end } = this.range();
    const span = Math.max(end - start, 1);
    const rows = this.props.payload.records ?? [];

    return this.renderShell(html`
      <div class="sum-gantt-toolbar">
        <h2>${this.props.payload.arch.title ?? "Gantt"}</h2>
        <div class="sum-gantt-scale">
          ${(["day", "week", "month"] as GanttScale[]).map(
            (scale) => html`<button
              type="button"
              class=${this.scale === scale ? "sum-btn sum-btn--secondary" : "sum-btn sum-btn--ghost"}
              @click=${() => this.setScale(scale)}
            >${scale}</button>`,
          )}
        </div>
      </div>
      <ul class="sum-gantt-rows">
        ${forEach(rows, (row) => Number(row.id ?? 0), (row) => {
          const from = parseArchDate(row[startField])?.getTime();
          const to = parseArchDate(row[stopField])?.getTime() ?? from;
          if (from == null || to == null) {
            return html`<li class="sum-gantt-row">
              <span class="sum-gantt-label">${recordDisplayLabel(row, Number(row.id ?? 0))}</span>
            </li>`;
          }
          const left = ((from - start) / span) * 100;
          const width = Math.max(((to - from) / span) * 100, 0.8);
          return html`<li class="sum-gantt-row" @click=${() => this.openRecord(row)}>
            <span class="sum-gantt-label">${recordDisplayLabel(row, Number(row.id ?? 0))}</span>
            <div class="sum-gantt-track">
              <div class="sum-gantt-bar" style=${`left:${left}%;width:${width}%`}></div>
            </div>
          </li>`;
        })}
      </ul>
    `);
  }
}
