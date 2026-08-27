import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { VIEW_FORM } from "../../constants/routes.js";
import { useState } from "../../runtime/hooks.js";

interface CalendarViewProps {
  payload: SwcWorkspacePayload;
}

const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

export class CalendarView extends SwcComponent<CalendarViewProps> {
  private dateField = "date_deadline";
  private year = 0;
  private month = 0;

  setup(): void {
    const now = new Date();
    this.year = now.getFullYear();
    this.month = now.getMonth();
    const archStart = this.props.payload.arch.calendar?.dateStart;
    if (archStart) {
      this.dateField = archStart;
    } else {
      const fields = this.props.payload.arch.fields;
      const dateField = fields.find((f) => f.type === "date" || f.type === "datetime");
      if (dateField) this.dateField = dateField.name;
    }
    const [, bump] = useState(0);
    this.bump = () => bump((n) => n + 1);
  }

  private bump: (() => void) | null = null;

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
    const id = Number(row.id ?? 0);
    if (id <= 0) return;
    const p = this.props.payload;
    this.env.services.action.openRecord({
      actionId: p.actionId,
      menuId: p.menuId,
      recordId: id,
      viewType: VIEW_FORM,
    });
  }

  private shiftMonth(delta: number): void {
    const d = new Date(this.year, this.month + delta, 1);
    this.year = d.getFullYear();
    this.month = d.getMonth();
    this.bump?.();
  }

  private cells(): Array<{ date: Date; inMonth: boolean }> {
    const first = new Date(this.year, this.month, 1);
    const start = new Date(first);
    start.setDate(1 - first.getDay());
    const out: Array<{ date: Date; inMonth: boolean }> = [];
    for (let i = 0; i < 42; i++) {
      const d = new Date(start);
      d.setDate(start.getDate() + i);
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

  template() {
    const events = this.eventsByDay();
    const title = new Date(this.year, this.month, 1).toLocaleString(undefined, {
      month: "long",
      year: "numeric",
    });
    return html`
      <div class="sum-calendar-view">
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
                    ${String(row.name ?? row.display_name ?? `#${row.id}`)}
                  </li>`,
                )}
              </ul>
            </section>`;
          })}
        </div>
      </div>
    `;
  }
}
