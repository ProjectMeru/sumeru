import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcArchField, SwcWorkspacePayload } from "../../types/workspace.js";
import { renderCollectionToolbar } from "../shared/view-toolbar.js";
import { renderKanbanCardInner } from "./kanban-card.js";
import {
  parseFilterCSV,
  renderSearchFilters,
  toggleFilterName,
} from "../list/control-panel.js";
import { RECORD_UPDATED, VIEW_FORM, VIEW_KANBAN } from "../../constants/routes.js";
import { SwcError } from "../../runtime/error.js";

interface KanbanViewProps {
  payload: SwcWorkspacePayload;
}

export class KanbanView extends SwcComponent<KanbanViewProps> {
  private search = "";
  private filters: string[] = [];
  private drafts: Record<string, string> = {};

  setup(): void {
    this.syncFromPayload(this.props.payload);
  }

  onPropsChanged(props: KanbanViewProps): void {
    this.syncFromPayload(props.payload);
  }

  private syncFromPayload(p: SwcWorkspacePayload): void {
    this.search = p.listSearch ?? "";
    this.filters = parseFilterCSV(p.listFilter);
  }

  private cardFields(): SwcArchField[] {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
  }

  private navigateKanban(patch: { listSearch?: string; listFilter?: string }): void {
    const p = this.props.payload;
    this.env.services.action.navigate(
      this.env.services.router.workspaceUrl({
        actionId: p.actionId,
        menuId: p.menuId,
        viewType: VIEW_KANBAN,
        listSearch: patch.listSearch ?? this.search,
        listFilter: patch.listFilter ?? this.filters.join(","),
      }),
    );
  }

  private applySearch(): void {
    this.navigateKanban({ listSearch: this.search });
  }

  private applyFilter(name: string): void {
    this.navigateKanban({ listFilter: toggleFilterName(this.filters, name).join(",") });
  }

  private openCard(row: Record<string, unknown>): void {
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

  private async moveCard(recordId: number, columnValue: number): Promise<void> {
    const groupField = this.props.payload.arch.kanban?.groupField;
    if (!groupField) return;
    try {
      await this.env.services.rpc.write(this.props.payload.model, [recordId], {
        [groupField]: columnValue || false,
      });
      this.env.services.bus.emit(RECORD_UPDATED, {
        model: this.props.payload.model,
        id: recordId,
      });
    } catch (err) {
      this.env.services.notification.error(
        "Move failed",
        err instanceof SwcError ? err.message : String(err),
      );
    }
  }

  private async quickCreate(columnValue: number): Promise<void> {
    const key = String(columnValue);
    const name = (this.drafts[key] ?? "").trim();
    const groupField = this.props.payload.arch.kanban?.groupField;
    if (!name || !groupField) return;
    try {
      await this.env.services.rpc.create(this.props.payload.model, {
        ...(this.props.payload.defaults ?? {}),
        name,
        [groupField]: columnValue || false,
      });
      this.drafts[key] = "";
      this.env.services.notification.success("Created", "Record created.");
      this.env.services.bus.emit(RECORD_UPDATED, { model: this.props.payload.model });
    } catch (err) {
      this.env.services.notification.error(
        "Create failed",
        err instanceof SwcError ? err.message : String(err),
      );
    }
  }

  private toolbar() {
    return renderCollectionToolbar({
      payload: this.props.payload,
      viewType: VIEW_KANBAN,
      search: this.search,
      onSearch: () => this.applySearch(),
      onInput: (next) => {
        this.search = next;
      },
    });
  }

  private renderCard(
    row: Record<string, unknown>,
    fields: SwcArchField[],
    opts: { draggable?: boolean; dropValue?: number } = {},
  ) {
    const draggable = Boolean(opts.draggable);
    const dropValue = opts.dropValue;
    return html`<div
      class="sum-kanban-card"
      draggable=${draggable ? "true" : undefined}
      @click=${() => this.openCard(row)}
      @dragstart=${draggable
        ? (ev: Event) => (ev as DragEvent).dataTransfer?.setData("text/plain", String(row.id))
        : undefined}
      @dragover=${dropValue !== undefined ? (ev: Event) => ev.preventDefault() : undefined}
      @drop=${dropValue !== undefined
        ? (ev: Event) => {
            ev.preventDefault();
            const id = Number((ev as DragEvent).dataTransfer?.getData("text/plain"));
            if (id) void this.moveCard(id, dropValue);
          }
        : undefined}
    >
      ${renderKanbanCardInner(row, fields)}
    </div>`;
  }

  template() {
    const p = this.props.payload;
    const kanban = p.arch.kanban;
    const fields = this.cardFields();
    const filters = p.arch.search?.filters ?? [];
    if (!kanban?.columns?.length) {
      const rows = p.records ?? [];
      return html`
        <div class="sum-kanban-view">
          ${this.toolbar()}
          ${renderSearchFilters({
            filters,
            active: this.filters,
            onToggle: (name) => this.applyFilter(name),
          })}
          <div class="sum-kanban-columns">
            ${rows.length === 0
              ? html`<div class="sum-kanban-empty">No records</div>`
              : rows.map((row) => this.renderCard(row, fields))}
          </div>
        </div>
      `;
    }
    return html`
      <div class="sum-kanban-view">
        ${this.toolbar()}
        ${renderSearchFilters({
          filters,
          active: this.filters,
          onToggle: (name) => this.applyFilter(name),
        })}
        <div class="sum-kanban-board sum-kanban-board--grouped">
          <div class="sum-kanban-stage-columns">
            ${kanban.columns.map(
              (col) => html`<div class="sum-kanban-stage-column" data-column=${String(col.value)}>
                <div class="sum-kanban-stage-header">
                  <span>${col.label}</span>
                  <span class="sum-kanban-stage-count">${String(col.records.length)}</span>
                </div>
                <div class="sum-kanban-cards">
                  ${col.records.map((row) =>
                    this.renderCard(row, fields, { draggable: kanban.draggable, dropValue: col.value }),
                  )}
                </div>
                ${kanban.quickCreate
                  ? html`<form
                      class="sum-kanban-quick-create"
                      @submit=${(ev: Event) => {
                        ev.preventDefault();
                        void this.quickCreate(col.value);
                      }}
                    >
                      <input
                        type="text"
                        class="sum-kanban-quick-input"
                        placeholder="Add…"
                        value=${this.drafts[String(col.value)] ?? ""}
                        @input=${(ev: Event) => {
                          this.drafts[String(col.value)] = (ev.target as HTMLInputElement).value;
                        }}
                      />
                      <button type="submit" class="sum-btn sum-btn--ghost">Add</button>
                    </form>`
                  : ""}
              </div>`,
            )}
          </div>
        </div>
      </div>
    `;
  }
}
