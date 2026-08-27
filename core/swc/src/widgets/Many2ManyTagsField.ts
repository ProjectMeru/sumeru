import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { fieldInputId, renderFieldShell } from "./field-shell.js";
import { AsyncFieldController } from "./field-async.js";
import type { FieldWidgetProps } from "./field-props.js";
import type { SwcRecord } from "../model/record.js";
import { inputValueFromEvent } from "./field-events.js";
import { isFieldReadonly } from "../model/modifiers.js";

interface TagRow {
  id: number;
  name: string;
}

function tagIds(record: SwcRecord, fieldName: string): number[] {
  const rawValue = record.get(fieldName);
  if (!Array.isArray(rawValue)) return [];
  return rawValue.map((v) => Number(v)).filter((n) => !Number.isNaN(n));
}

function tagNamesFromRecord(record: SwcRecord, fieldName: string): Map<number, string> {
  const out = new Map<number, string>();
  const rawValue = record.get(`${fieldName}_names`);
  if (Array.isArray(rawValue)) {
    for (const item of rawValue) {
      if (item && typeof item === "object") {
        const row = item as Record<string, unknown>;
        const id = Number(row.id);
        if (!Number.isNaN(id)) out.set(id, String(row.name ?? id));
      }
    }
  }
  return out;
}

export class Many2ManyTagsField extends SwcComponent<FieldWidgetProps> {
  private catalog: TagRow[] = [];
  private loaded = false;
  private readonly asyncCtrl = new AsyncFieldController(this);

  override setup(): void {
    void this.loadCatalog();
  }

  override onWillUnmount(): void {
    this.asyncCtrl.cancel();
  }

  private async loadCatalog(): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const { field, record, readonly } = this.props;

    if (isFieldReadonly(field, record, readonly)) {
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
      return;
    }

    const comodel = field.relation ?? field.options?.relation ?? "";
    if (!comodel) {
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
      return;
    }

    const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name"], 500);
    this.catalog = rows.map((row) => ({ id: Number(row.id), name: String(row.name ?? row.id) }));
    this.loaded = true;
    this.asyncCtrl.commitIfCurrent(gen);
  }

  private selectedTags(): TagRow[] {
    const { field, record } = this.props;
    const ids = tagIds(record, field.name);
    const names = tagNamesFromRecord(record, field.name);
    return ids.map((id) => {
      const fromCatalog = this.catalog.find((tag) => tag.id === id);
      if (fromCatalog) return fromCatalog;
      const fromRecord = names.get(id);
      if (fromRecord) return { id, name: fromRecord };
      return { id, name: `#${id}` };
    });
  }

  private setIds(ids: number[]): void {
    this.props.record.set(this.props.field.name, ids);
    this.asyncCtrl.refresh();
  }

  override template() {
    const { field, record, readonly } = this.props;
    const selected = this.selectedTags();
    const selectedSet = new Set(selected.map((tag) => tag.id));

    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(
        field,
        html`<div class="sum-multi-select-tags sum-multi-select-tags--readonly sum-field-tags">
          ${selected.map((tag) => html`<span class="sum-multi-select-tag"><span class="sum-multi-select-tag-label">${tag.name}</span></span>`)}
        </div>`,
        { labelFor: false },
      );
    }

    const id = fieldInputId(field);

    return renderFieldShell(
      field,
      html`<div class="sum-multi-select-box">
        <div class="sum-multi-select-tags sum-field-tags">
          ${selected.map(
            (tag) =>
              html`<span class="sum-multi-select-tag">
                <span class="sum-multi-select-tag-label">${tag.name}</span>
                <button type="button" class="sum-multi-select-remove" aria-label="Remove" @click=${() => this.setIds(selected.filter((t) => t.id !== tag.id).map((t) => t.id))}>×</button>
              </span>`,
          )}
        </div>
        <select
          id=${id}
          class="sum-multi-select-add sum-field-select"
          @change=${(event: Event) => {
            const fieldValue = Number(inputValueFromEvent(event));
            const select = event.target as HTMLSelectElement;
            select.value = "";
            if (!fieldValue || selectedSet.has(fieldValue)) return;
            this.setIds([...selected.map((tag) => tag.id), fieldValue]);
          }}
        >
          <option value="">Add tag…</option>
          ${this.catalog
            .filter((tag) => !selectedSet.has(tag.id))
            .map((tag) => html`<option value=${String(tag.id)}>${tag.name}</option>`)}
        </select>
        ${!this.loaded ? html`<span class="sum-field-hint">Loading…</span>` : ""}
      </div>`,
      { labelFor: id },
    );
  }
}
