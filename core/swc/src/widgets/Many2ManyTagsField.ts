import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import { fieldInputId, renderFieldShell } from "./field-shell.js";
import { AsyncFieldController } from "./field-async.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

interface TagRow {
  id: number;
  name: string;
}

function tagIds(record: SwcRecord, fieldName: string): number[] {
  const raw = record.get(fieldName);
  if (!Array.isArray(raw)) return [];
  return raw.map((v) => Number(v)).filter((n) => !Number.isNaN(n));
}

function tagNamesFromRecord(record: SwcRecord, fieldName: string): Map<number, string> {
  const out = new Map<number, string>();
  const raw = record.get(`${fieldName}_names`);
  if (Array.isArray(raw)) {
    for (const item of raw) {
      if (item && typeof item === "object") {
        const row = item as Record<string, unknown>;
        const id = Number(row.id);
        if (!Number.isNaN(id)) out.set(id, String(row.name ?? id));
      }
    }
  }
  return out;
}

export class Many2ManyTagsField extends SwcComponent<FieldProps> {
  private catalog: TagRow[] = [];
  private loaded = false;
  private readonly asyncCtrl = new AsyncFieldController(this);

  setup(): void {
    void this.loadCatalog();
  }

  onWillUnmount(): void {
    this.asyncCtrl.cancel();
  }

  private async loadCatalog(): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const { field, readonly } = this.props;

    if (readonly || field.readonly) {
      this.loaded = true;
      this.asyncCtrl.finish(gen);
      return;
    }

    const comodel = field.relation ?? field.options?.relation ?? "";
    if (!comodel) {
      this.loaded = true;
      this.asyncCtrl.finish(gen);
      return;
    }

    const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name"], 500);
    this.catalog = rows.map((row) => ({ id: Number(row.id), name: String(row.name ?? row.id) }));
    this.loaded = true;
    this.asyncCtrl.finish(gen);
  }

  private selectedTags(): TagRow[] {
    const { field, record } = this.props;
    const ids = tagIds(record, field.name);
    const names = tagNamesFromRecord(record, field.name);
    return ids.map((id) => {
      const fromCatalog = this.catalog.find((t) => t.id === id);
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

  template() {
    const { field, readonly } = this.props;
    const selected = this.selectedTags();
    const selectedSet = new Set(selected.map((t) => t.id));

    if (readonly || field.readonly) {
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
          @change=${(ev: Event) => {
            const val = Number((ev.target as HTMLSelectElement).value);
            (ev.target as HTMLSelectElement).value = "";
            if (!val || selectedSet.has(val)) return;
            this.setIds([...selected.map((t) => t.id), val]);
          }}
        >
          <option value="">Add tag…</option>
          ${this.catalog
            .filter((t) => !selectedSet.has(t.id))
            .map((t) => html`<option value=${String(t.id)}>${t.name}</option>`)}
        </select>
        ${!this.loaded ? html`<span class="sum-field-hint">Loading…</span>` : ""}
      </div>`,
      { labelFor: id },
    );
  }
}
