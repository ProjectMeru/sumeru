import { html } from "../../template/html.js";
import { VIEW_CALENDAR } from "../../constants/routes.js";
import { CollectionView } from "../shared/collection-view.js";
import { resolveArchDateField } from "../shared/arch-date.js";
import { openWorkspaceRecord } from "../shared/collection-navigation.js";
import { recordDisplayLabel } from "../shared/field-display.js";

const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

/** Calendar view — month grid with events keyed by arch date_start. */
export class CalendarView extends CollectionView {
  protected readonly collectionViewType = VIEW_CALENDAR;
  private dateField = "date_deadline";
  private year = 0;
  private month = 0;

  protected override onCollectionSetup(): void {
    const now = new Date();
    this.year = now.getFullYear();
    this.month = now.getMonth();
    this.dateField = resolveArchDateField(this.props.payload.arch, ["calendar"], "date_deadline");
  }

  protected override onCollectionPropsChanged(): void {
    this.dateField = resolveArchDateField(this.props.payload.arch, ["calendar"], "date_deadline");
  }

  private eventsByDay(): Map<string, Record<string, unknown>[]> {
    const map = new Map<string, Record<string, unknown>[]>();
    for (const row of this.props.payload.records ?? []) {
      const raw = String(row[this.dateField] ?? "").slice(0, 10);
      if (!raw) continue;
      if (!map.has(raw)) map.set(raw, []);
      map.get(raw)!.push(row);
    }
    return map;
  }

  private openRecord(row: Record<string, unknown>): void {
    openWorkspaceRecord(this.env, this.props.payload, row);
  }

  private shiftMonth(delta: number): void {
    const d = new Date(this.year, this.month + delta, 1);
    this.year = d.getFullYear();
    this.month = d.getMonth();
    this.rerender();
  }

  private cells(): Array<{ date: Date; inMonth: boolean }> {
    const first = new Date(this.year, this.month, 1);
    const start = new Date(first);
    start.setDate(1 - first.getDay());
    const out: Array<{ date: Date; inMonth: boolean }> = [];
    for (let index = 0; index < 42; index++) {
      const d = new Date(start);
      d.setDate(start.getDate() + index);
      out.push({ date: d, inMonth: d.getMonth() === this.month });
    }
    return out;
  }

  private iso(d: Date): string {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
  }

  override template() {
    const events = this.eventsByDay();
    const title = new Date(this.year, this.month, 1).toLocaleString(undefined, {
      month: "long",
      year: "numeric",
    });
    return this.renderShell(html`
      <div class="sum-calendar-toolbar">
        <button type="button" class="sum-btn sum-btn--ghost" @click=${() => this.shiftMonth(-1)}>Prev</button>
        <h2 class="sum-calendar-title">${this.props.payload.arch.title ?? title}</h2>
        <button type="button" class="sum-btn sum-btn--ghost" @click=${() => this.shiftMonth(1)}>Next</button>
      </div>
      <div class="sum-calendar-grid">
        ${WEEKDAYS.map((d) => html`<div class="sum-calendar-weekday">${d}</div>`)}
        ${this.cells().map((cell) => {
          const key = this.iso(cell.date);
          const rows = events.get(key) ?? [];
          return html`<section class=${cell.inMonth ? "sum-calendar-cell" : "sum-calendar-cell sum-calendar-cell--muted"}>
            <h3 class="sum-calendar-day-title">${String(cell.date.getDate())}</h3>
            <ul class="sum-calendar-events">
              ${rows.map(
                (row) => html`<li class="sum-calendar-event" @click=${() => this.openRecord(row)}>
                  ${recordDisplayLabel(row)}
                </li>`,
              )}
            </ul>
          </section>`;
        })}
      </div>
    `);
  }
}
