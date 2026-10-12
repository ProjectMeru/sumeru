import { html, type TemplateResult } from "../../template/html.js";
import { VIEW_ACTIVITY, VIEW_FORM } from "../../constants/routes.js";
import { CollectionView } from "../shared/collection-view.js";
import { calendarDayKey } from "../calendar/CalendarView.js";
import { runObjectAction } from "../shared/object-action.js";

type ColumnName = "overdue" | "today" | "later";

const COLUMNS: Array<{ key: ColumnName; label: string }> = [
  { key: "overdue", label: "Overdue" },
  { key: "today", label: "Today" },
  { key: "later", label: "Later" },
];

/** Activity board for mail.activity, with an optional month grid when date_start is set. */
export class ActivityView extends CollectionView {
  protected readonly collectionViewType = VIEW_ACTIVITY;
  private year = 0;
  private month = 0;

  protected override onCollectionSetup(): void {
    const now = new Date();
    this.year = now.getFullYear();
    this.month = now.getMonth();
  }

  private rows(): Record<string, unknown>[] {
    return [...(this.props.payload.records ?? [])];
  }

  private deadline(row: Record<string, unknown>): string {
    return String(row.date_deadline ?? "").slice(0, 10);
  }

  private column(row: Record<string, unknown>, today: string): ColumnName {
    const due = this.deadline(row);
    if (!due) return "later";
    if (due < today) return "overdue";
    if (due === today) return "today";
    return "later";
  }

  private userKey(row: Record<string, unknown>): string {
    const name = row.user_id_name;
    if (name) return String(name);
    const raw = row.user_id;
    if (Array.isArray(raw) && raw.length > 1) return String(raw[1]);
    if (raw == null || raw === false || raw === "") return "Unassigned";
    return `User ${String(raw)}`;
  }

  private typeName(row: Record<string, unknown>): string {
    const named = row.activity_type_id_name;
    if (named) return String(named);
    const raw = row.activity_type_id;
    if (Array.isArray(raw) && raw.length > 1) return String(raw[1]);
    return raw ? String(raw) : "Activity";
  }

  private openParent(row: Record<string, unknown>): void {
    const model = String(row.model ?? "").trim();
    const id = Number(row.res_id ?? 0);
    if (!model || id <= 0) return;
    const params = new URLSearchParams({
      model,
      view_type: VIEW_FORM,
      id: String(id),
    });
    this.env.services.action.navigate(`/web?${params.toString()}`);
  }

  private async run(row: Record<string, unknown>, method: "action_done" | "action_snooze", label: string): Promise<void> {
    const id = Number(row.id ?? 0);
    if (id <= 0) return;
    await runObjectAction(this.env, {
      model: "mail.activity",
      methodName: method,
      recordId: id,
      buttonLabel: label,
      onSuccess: () => {
        this.env.services.action.navigate(window.location.href);
      },
    });
  }

  private card(row: Record<string, unknown>): TemplateResult {
    const due = this.deadline(row);
    return html`<article class="sum-kanban-card" @click=${() => this.openParent(row)}>
      <div class="sum-kanban-card-title">${String(row.summary ?? this.typeName(row))}</div>
      <div class="sum-kanban-card-field">${this.typeName(row)}${due ? "" : " · Undated"}</div>
      <div class="sum-kanban-card-actions">
        <button type="button" class="sum-btn sum-btn--ghost" data-activity-done @click=${(event: Event) => {
          event.stopPropagation();
          void this.run(row, "action_done", "Done");
        }}>Done</button>
        <button type="button" class="sum-btn sum-btn--ghost" data-activity-snooze @click=${(event: Event) => {
          event.stopPropagation();
          void this.run(row, "action_snooze", "Snooze");
        }}>Snooze</button>
      </div>
    </article>`;
  }

  private board(): TemplateResult {
    const today = calendarDayKey(new Date());
    const grouped = new Map<ColumnName, Map<string, Record<string, unknown>[]>>();
    for (const column of COLUMNS) grouped.set(column.key, new Map());
    for (const row of this.rows()) {
      const column = this.column(row, today);
      const users = grouped.get(column)!;
      const user = this.userKey(row);
      const list = users.get(user) ?? [];
      list.push(row);
      users.set(user, list);
    }
    if (this.rows().length === 0) {
      return html`<p class="sum-activity-empty">No activities</p>`;
    }
    return html`<div class="sum-kanban-board">
      ${COLUMNS.map((column) => html`<section class="sum-kanban-column" data-activity-column=${column.key}>
        <h2 class="sum-kanban-column-title">${column.label}</h2>
        ${[...(grouped.get(column.key)?.entries() ?? [])].map(
          ([user, rows]) => html`<div class="sum-activity-user">
            <h3>${user}</h3>
            ${rows.map((row) => this.card(row))}
          </div>`,
        )}
      </section>`)}
    </div>`;
  }

  private shiftMonth(delta: number): void {
    const next = new Date(this.year, this.month + delta, 1);
    this.year = next.getFullYear();
    this.month = next.getMonth();
    this.rerender();
  }

  private monthGrid(): TemplateResult {
    const field = this.props.payload.arch.calendar?.dateStart || "date_deadline";
    const byDay = new Map<string, Record<string, unknown>[]>();
    for (const row of this.rows()) {
      const key = String(row[field] ?? "").slice(0, 10);
      if (!key) continue;
      const list = byDay.get(key) ?? [];
      list.push(row);
      byDay.set(key, list);
    }
    const first = new Date(this.year, this.month, 1);
    const start = new Date(first);
    start.setDate(1 - first.getDay());
    const cells: Date[] = [];
    for (let index = 0; index < 42; index++) {
      const day = new Date(start);
      day.setDate(start.getDate() + index);
      cells.push(day);
    }
    const title = first.toLocaleString(undefined, { month: "long", year: "numeric" });
    return html`<div class="sum-calendar">
      <div class="sum-calendar-toolbar">
        <button type="button" class="sum-btn sum-btn--ghost" @click=${() => this.shiftMonth(-1)}>Prev</button>
        <h2 class="sum-calendar-title">${title}</h2>
        <button type="button" class="sum-btn sum-btn--ghost" @click=${() => this.shiftMonth(1)}>Next</button>
      </div>
      <div class="sum-calendar-grid">
        ${cells.map((day) => {
          const key = calendarDayKey(day);
          const items = byDay.get(key) ?? [];
          return html`<div class="sum-calendar-cell" data-day=${key}>
            <span>${day.getDate()}</span>
            ${items.map((row) => html`<button type="button" class="sum-calendar-event" @click=${() => this.openParent(row)}>${String(row.summary ?? "")}</button>`)}
          </div>`;
        })}
      </div>
    </div>`;
  }

  override template() {
    const body = this.props.payload.arch.calendar?.dateStart ? this.monthGrid() : this.board();
    return this.renderShell(body);
  }
}
