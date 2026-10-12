import { SwcComponent } from "../runtime/component.js";
import { html, type TemplateResult } from "../template/html.js";
import { AsyncFieldController } from "./field-async.js";
import { isFieldReadonly } from "../model/modifiers.js";
import { fieldInputId, renderFieldShell, shellOptions } from "./field-shell.js";
import type { FieldWidgetProps } from "./field-props.js";

interface PropertyDef {
  key: string;
  label: string;
  value_type: string;
  selection?: string;
}

function asObject(raw: unknown): Record<string, unknown> {
  if (raw && typeof raw === "object" && !Array.isArray(raw)) {
    return { ...(raw as Record<string, unknown>) };
  }
  if (typeof raw === "string" && raw.trim()) {
    try {
      const parsed = JSON.parse(raw) as unknown;
      if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
        return parsed as Record<string, unknown>;
      }
    } catch {
      return {};
    }
  }
  return {};
}

/** One input per property definition. An empty definition is an empty state. */
export class PropertiesField extends SwcComponent<FieldWidgetProps> {
  private defs: PropertyDef[] = [];
  private loaded = false;
  private readonly asyncCtrl = new AsyncFieldController(this);

  override setup(): void {
    void this.loadDefs();
  }

  override onWillUnmount(): void {
    this.asyncCtrl.cancel();
  }

  private async loadDefs(): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const schema = this.props.field.options?.schema;
    if (schema) {
      try {
        const parsed = JSON.parse(schema) as PropertyDef[];
        this.defs = Array.isArray(parsed) ? parsed.filter((row) => row.key) : [];
      } catch {
        this.defs = [];
      }
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
      return;
    }
    const model = this.props.modelName ?? "";
    if (!model) {
      this.defs = [];
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
      return;
    }
    const rows = await this.env.services.rpc.searchRead(
      "sys.property.definition",
      [
        ["model", "=", model],
        ["field_name", "=", this.props.field.name],
      ],
      ["key", "label", "value_type", "selection", "sequence"],
      50,
    );
    rows.sort((a, b) => Number(a.sequence ?? 0) - Number(b.sequence ?? 0));
    this.defs = rows.map((row) => ({
      key: String(row.key ?? ""),
      label: String(row.label ?? row.key ?? ""),
      value_type: String(row.value_type ?? "char"),
      selection: row.selection ? String(row.selection) : "",
    }));
    this.loaded = true;
    this.asyncCtrl.commitIfCurrent(gen);
  }

  private setKey(key: string, value: unknown): void {
    const next = asObject(this.props.record.get(this.props.field.name));
    next[key] = value;
    this.props.record.set(this.props.field.name, next);
  }

  private input(def: PropertyDef, value: unknown, readonly: boolean): TemplateResult {
    const id = `${fieldInputId(this.props.field)}-${def.key}`;
    if (readonly) {
      return html`<span>${value == null ? "" : String(value)}</span>`;
    }
    if (def.value_type === "boolean") {
      return html`<input id=${id} type="checkbox" checked=${value === true ? "checked" : ""} @change=${(event: Event) => {
        const target = event.target as HTMLInputElement;
        this.setKey(def.key, target.checked);
      }} />`;
    }
    if (def.value_type === "float") {
      return html`<input id=${id} type="number" step="any" class="sum-field-input" .value=${value == null ? "" : String(value)} @change=${(event: Event) => {
        const target = event.target as HTMLInputElement;
        this.setKey(def.key, target.value === "" ? null : Number(target.value));
      }} />`;
    }
    if (def.value_type === "integer" || def.value_type === "many2one") {
      return html`<input id=${id} type="number" class="sum-field-input" .value=${value == null ? "" : String(value)} @change=${(event: Event) => {
        const target = event.target as HTMLInputElement;
        this.setKey(def.key, target.value === "" ? null : Number(target.value));
      }} />`;
    }
    if (def.value_type === "selection") {
      const options = (def.selection ?? "").split(",").map((part) => part.trim()).filter(Boolean);
      return html`<select id=${id} class="sum-field-input" @change=${(event: Event) => {
        this.setKey(def.key, (event.target as HTMLSelectElement).value);
      }}>
        <option value=""></option>
        ${options.map((part) => {
          const key = part.split(":")[0] ?? part;
          return html`<option value=${key} selected=${String(value ?? "") === key ? "selected" : ""}>${part}</option>`;
        })}
      </select>`;
    }
    return html`<input id=${id} class="sum-field-input" type=${def.value_type === "date" ? "date" : "text"} .value=${value == null ? "" : String(value)} @change=${(event: Event) => {
      this.setKey(def.key, (event.target as HTMLInputElement).value);
    }} />`;
  }

  override template() {
    const { field, record, readonly } = this.props;
    const locked = isFieldReadonly(field, record, readonly);
    const values = asObject(record.get(field.name));
    let body: TemplateResult;
    if (!this.loaded) {
      body = html`<p>Loading properties</p>`;
    } else if (this.defs.length === 0) {
      body = html`<p class="sum-properties-empty">No properties defined</p>`;
    } else {
      body = html`<div class="sum-properties">
        ${this.defs.map(
          (def) => html`<label class="sum-properties-row">
            <span>${def.label || def.key}</span>
            ${this.input(def, values[def.key], locked)}
          </label>`,
        )}
      </div>`;
    }
    return renderFieldShell(field, body, shellOptions(record, { labelFor: false }));
  }
}
