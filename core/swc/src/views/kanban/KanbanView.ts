import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcArchField, SwcWorkspacePayload } from "../../types/workspace.js";
import { renderKanbanCardInner } from "./kanban-card.js";
import { RECORD_UPDATED, VIEW_FORM, VIEW_KANBAN } from "../../constants/routes.js";
import { SwcError } from "../../runtime/error.js";
import { inputValueFromEvent } from "../../widgets/field-events.js";
import { CollectionBarHost, mountCollectionBar } from "../shared/collection-bar-host.js";

interface KanbanViewProps {
  payload: SwcWorkspacePayload;
}

export class KanbanView extends SwcComponent<KanbanViewProps> {
  private drafts: Record<string, string> = {};
  private collectionBar!: CollectionBarHost;

  override setup(): void {
    this.collectionBar = mountCollectionBar(this.props.payload, VIEW_KANBAN, this.env);
  }

  override onPropsChanged(props: KanbanViewProps): void {
    this.collectionBar.updateProps({ payload: props.payload, viewType: VIEW_KANBAN });
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
  }

  private cardFields(): SwcArchField[] {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
  }

  private openCard(row: Record<string, unknown>): void {
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

  private renderCard(
    row: Record<string, unknown>,
    fields: SwcArchField[],
    options: { draggable?: boolean; dropValue?: number } = {},
  ) {
    const draggable = Boolean(options.draggable);
    const dropValue = options.dropValue;
    return html`<div
      class="sum-kanban-card"
      draggable=${draggable ? "true" : undefined}
      @click=${() => this.openCard(row)}
      @dragstart=${draggable
        ? (event: Event) => (event as DragEvent).dataTransfer?.setData("text/plain", String(row.id))
        : undefined}
      @dragover=${dropValue !== undefined ? (event: Event) => event.preventDefault() : undefined}
      @drop=${dropValue !== undefined
        ? (event: Event) => {
            event.preventDefault();
            const id = Number((event as DragEvent).dataTransfer?.getData("text/plain"));
            if (id) void this.moveCard(id, dropValue);
          }
        : undefined}
    >
      ${renderKanbanCardInner(row, fields)}
    </div>`;
  }

  override template() {
    const payload = this.props.payload;
    const kanban = payload.arch.kanban;
    const fields = this.cardFields();
    if (!kanban?.columns?.length) {
      const rows = payload.records ?? [];
      return html`
        <div class="sum-collection-view sum-kanban-view">
          ${this.collectionBar.renderOrPatch()}
          <div class="sum-kanban-columns">
            ${rows.length === 0
              ? html`<div class="sum-kanban-empty">No records</div>`
              : rows.map((row) => this.renderCard(row, fields))}
          </div>
        </div>
      `;
    }
    return html`
      <div class="sum-collection-view sum-kanban-view">
        ${this.collectionBar.renderOrPatch()}
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
                      @submit=${(event: Event) => {
                        event.preventDefault();
                        void this.quickCreate(col.value);
                      }}
                    >
                      <input
                        type="text"
                        class="sum-kanban-quick-input"
                        placeholder="Add…"
                        value=${this.drafts[String(col.value)] ?? ""}
                        @input=${(event: Event) => {
                          this.drafts[String(col.value)] = inputValueFromEvent(event);
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
