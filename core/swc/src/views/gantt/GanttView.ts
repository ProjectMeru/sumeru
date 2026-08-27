import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import { forEach } from "../../template/helpers.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { VIEW_FORM, VIEW_GANTT } from "../../constants/routes.js";
import { CollectionBarHost, mountCollectionBar } from "../shared/collection-bar-host.js";

interface GanttViewProps {
  payload: SwcWorkspacePayload;
}

type GanttScale = "day" | "week" | "month";

function parseDate(raw: unknown): Date | null {
  const text = String(raw ?? "").trim();
  if (!text) return null;
  const date = new Date(text);
  return Number.isNaN(date.getTime()) ? null : date;
}

export class GanttView extends SwcComponent<GanttViewProps> {
  private scale: GanttScale = "week";
  private collectionBar!: CollectionBarHost;

  override setup(): void {
    this.collectionBar = mountCollectionBar(this.props.payload, VIEW_GANTT, this.env);
  }

  override onPropsChanged(props: GanttViewProps): void {
    this.collectionBar.updateProps({ payload: props.payload, viewType: VIEW_GANTT });
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
  }

  private setScale(next: GanttScale): void {
    this.scale = next;
    this.rerender();
  }

  private dateStartField(): string {
    return this.props.payload.arch.gantt?.dateStart
      || this.props.payload.arch.calendar?.dateStart
      || this.props.payload.arch.fields.find((f) => f.type === "date" || f.type === "datetime")?.name
      || "date_start";
  }

  private dateStopField(): string {
    return this.props.payload.arch.gantt?.dateStop
      || this.props.payload.arch.calendar?.dateStop
      || this.dateStartField();
  }

  private openRecord(row: Record<string, unknown>): void {
    const id = Number(row.id ?? 0);
    if (id <= 0) return;
    const payload = this.props.payload;
    this.env.services.action.openRecord({
      actionId: payload.actionId,
      menuId: payload.menuId,
      recordId: id,
      viewType: VIEW_FORM,
    });
  }

  private range(): { start: number; end: number } {
    const startField = this.dateStartField();
    const stopField = this.dateStopField();
    let min = Infinity;
    let max = -Infinity;
    for (const row of this.props.payload.records ?? []) {
      const start = parseDate(row[startField])?.getTime();
      const stop = parseDate(row[stopField])?.getTime() ?? start;
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

    return html`
      <div class="sum-collection-view sum-gantt-view">
        ${this.collectionBar.renderOrPatch()}
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
            const from = parseDate(row[startField])?.getTime();
            const to = parseDate(row[stopField])?.getTime() ?? from;
            if (from == null || to == null) {
              return html`<li class="sum-gantt-row">
                <span class="sum-gantt-label">${String(row.name ?? row.display_name ?? row.id)}</span>
              </li>`;
            }
            const left = ((from - start) / span) * 100;
            const width = Math.max(((to - from) / span) * 100, 0.8);
            return html`<li class="sum-gantt-row" @click=${() => this.openRecord(row)}>
              <span class="sum-gantt-label">${String(row.name ?? row.display_name ?? row.id)}</span>
              <div class="sum-gantt-track">
                <div class="sum-gantt-bar" style=${`left:${left}%;width:${width}%`}></div>
              </div>
            </li>`;
          })}
        </ul>
      </div>
    `;
  }
}
