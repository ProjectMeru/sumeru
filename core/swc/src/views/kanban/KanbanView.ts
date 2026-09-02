import { html } from "../../template/html.js";
import type { SwcArchField } from "../../types/workspace.js";
import { kanbanColumnsStyle } from "./kanban-layout.js";
import { renderKanbanCardInner, isKanbanCardRotting } from "./kanban-card.js";
import {
  KANBAN_COLOR_COUNT,
  KANBAN_COLOR_LABELS,
  kanbanColorHasField,
  kanbanStripeClass,
  resolveCardColor,
} from "./kanban-color.js";
import { RECORD_UPDATED, VIEW_FORM, VIEW_KANBAN } from "../../constants/routes.js";
import { SwcError } from "../../runtime/error.js";
import { inputValueFromEvent } from "../../widgets/field-events.js";
import { CollectionView } from "../shared/collection-view.js";

export class KanbanView extends CollectionView {
  protected readonly collectionViewType = VIEW_KANBAN;
  private drafts: Record<string, string> = {};
  private openColorPickerId: number | null = null;
  private draggingCardId: number | null = null;

  private cardFields(): SwcArchField[] {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
  }

  private kanbanGridStyle(): string {
    return kanbanColumnsStyle(this.props.payload.arch.kanban?.columnsPerRow);
  }

  private allCardFields(): SwcArchField[] {
    return this.props.payload.arch.fields;
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

  private async setCardColor(recordId: number, color: number | false): Promise<void> {
    try {
      await this.env.services.rpc.write(this.props.payload.model, [recordId], { color });
      this.openColorPickerId = null;
      this.env.services.bus.emit(RECORD_UPDATED, {
        model: this.props.payload.model,
        id: recordId,
      });
    } catch (err) {
      this.env.services.notification.error(
        "Color update failed",
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

  private renderColorPicker(row: Record<string, unknown>): ReturnType<typeof html> | "" {
    const id = Number(row.id ?? 0);
    if (id <= 0 || !kanbanColorHasField(this.allCardFields())) return "";
    const open = this.openColorPickerId === id;
    const swatches = Array.from({ length: KANBAN_COLOR_COUNT }, (_, i) =>
      html`<button
        type="button"
        class="sum-kanban-color-swatch sum-kanban-color-swatch--${i}"
        aria-label=${KANBAN_COLOR_LABELS[i] ?? `Color ${i}`}
        title=${KANBAN_COLOR_LABELS[i] ?? `Color ${i}`}
        @click=${(event: Event) => {
          event.stopPropagation();
          void this.setCardColor(id, i);
        }}
      ></button>`,
    );
    return html`<div class="sum-kanban-card-color">
      <button
        type="button"
        class="sum-kanban-card-color-btn"
        aria-label="Set card color"
        aria-expanded=${open ? "true" : "false"}
        title="Set card color"
        @click=${(event: Event) => {
          event.stopPropagation();
          this.openColorPickerId = open ? null : id;
          this.rerender();
        }}
      >
        ◑
      </button>
      ${open
        ? html`<div class="sum-kanban-color-popover" role="menu" @click=${(event: Event) => event.stopPropagation()}>
            ${swatches}
            <button
              type="button"
              class="sum-kanban-color-clear"
              @click=${(event: Event) => {
                event.stopPropagation();
                void this.setCardColor(id, false);
              }}
            >
              Clear
            </button>
          </div>`
        : ""}
    </div>`;
  }

  private renderCard(
    row: Record<string, unknown>,
    fields: SwcArchField[],
    options: { draggable?: boolean; dropValue?: number; rottingDays?: number; stageColor?: number } = {},
  ) {
    const draggable = Boolean(options.draggable);
    const dropValue = options.dropValue;
    const recordId = Number(row.id ?? 0);
    const rotting = isKanbanCardRotting(row, options.rottingDays ?? 7);
    const dragging = this.draggingCardId === recordId;
    const cardColor = resolveCardColor(row, options.stageColor);
    const stripeClass = kanbanStripeClass(cardColor);
    const cardClasses = [
      "sum-kanban-card",
      rotting ? "sum-kanban-card--rotting" : "",
      dragging ? "sum-kanban-card--dragging" : "",
      draggable ? "sum-kanban-card--draggable" : "",
    ]
      .filter(Boolean)
      .join(" ");

    return html`<div
      class=${cardClasses}
      draggable=${draggable ? "true" : undefined}
      @click=${() => this.openCard(row)}
      @dragstart=${draggable
        ? (event: Event) => {
            this.draggingCardId = recordId;
            this.rerender();
            (event as DragEvent).dataTransfer?.setData("text/plain", String(row.id));
          }
        : undefined}
      @dragend=${draggable
        ? () => {
            this.draggingCardId = null;
            this.rerender();
          }
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
      <div class=${stripeClass} aria-hidden="true"></div>
      <div class="sum-kanban-card-content">
        ${renderKanbanCardInner(row, fields, this.props.payload.model)}
        ${this.renderColorPicker(row)}
      </div>
    </div>`;
  }

  private stageHeaderClass(color: number): string {
    if (color >= 0 && color <= 11) {
      return `sum-kanban-stage-header sum-kanban-stage-header--color-${color}`;
    }
    return "sum-kanban-stage-header";
  }

  private renderStageSection(
    col: {
      value: number;
      label: string;
      color: number;
      records: Record<string, unknown>[];
      progressSum?: number;
      progressMax?: number;
      rottingDays?: number;
    },
    fields: SwcArchField[],
    kanban: NonNullable<(typeof this.props.payload.arch.kanban)>,
  ) {
    const max = col.progressMax && col.progressMax > 0 ? col.progressMax : col.progressSum ?? 0;
    const pct = max > 0 ? Math.min(100, Math.round(((col.progressSum ?? 0) / max) * 100)) : 0;
    return html`<div
      class="sum-kanban-stage-section"
      data-column=${String(col.value)}
      @dragover=${kanban.draggable ? (event: Event) => event.preventDefault() : undefined}
      @drop=${kanban.draggable
        ? (event: Event) => {
            event.preventDefault();
            const id = Number((event as DragEvent).dataTransfer?.getData("text/plain"));
            if (id) void this.moveCard(id, col.value);
          }
        : undefined}
    >
      <div class=${this.stageHeaderClass(col.color)}>
        <span>${col.label}</span>
        <span class="sum-kanban-stage-count">${String(col.records.length)}</span>
      </div>
      ${col.progressSum != null && col.progressSum > 0
        ? html`<div class="sum-kanban-column-progress" title=${`MRR ${col.progressSum}`}>
            <div class="sum-progress">
              <div class="sum-progress-bar" style=${`width:${pct}%`}></div>
            </div>
          </div>`
        : ""}
      <div class="sum-kanban-columns sum-kanban-columns--in-stage" style=${this.kanbanGridStyle()}>
        ${col.records.length === 0
          ? html`<div class="sum-kanban-empty sum-kanban-empty--stage">No records</div>`
          : col.records.map((row) =>
              this.renderCard(row, fields, {
                draggable: kanban.draggable,
                dropValue: col.value,
                rottingDays: col.rottingDays,
                stageColor: col.color,
              }),
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
          <div class="sum-kanban-columns" style=${this.kanbanGridStyle()}>
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
          ${kanban.columns.map((col) => this.renderStageSection(col, fields, kanban))}
        </div>
      </div>
    `;
  }
}
