import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import {
  fieldInputId,
  fieldPlaceholder,
  fieldReadonlyInput,
  renderFieldShell,
} from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

function isDateTime(field: SwcArchField): boolean {
  return field.type === "datetime" || field.widget === "datetime";
}

function toNativeValue(field: SwcArchField, raw: unknown): string {
  const text = String(raw ?? "").trim();
  if (!text) return "";
  if (isDateTime(field)) {
    const d = new Date(text);
    if (Number.isNaN(d.getTime())) return text.slice(0, 16);
    const pad = (n: number) => String(n).padStart(2, "0");
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }
  return text.slice(0, 10);
}

function formatDisplay(field: SwcArchField, raw: unknown): string {
  const native = toNativeValue(field, raw);
  if (!native) return "";
  if (isDateTime(field)) {
    const d = new Date(native);
    if (Number.isNaN(d.getTime())) return native;
    return d.toLocaleString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }
  const d = new Date(`${native}T00:00:00`);
  if (Number.isNaN(d.getTime())) return native;
  return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

function todayNative(field: SwcArchField): string {
  const d = new Date();
  const pad = (n: number) => String(n).padStart(2, "0");
  if (isDateTime(field)) {
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export class DateField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const raw = record.get(field.name);
    const native = toNativeValue(field, raw);
    const display = formatDisplay(field, raw);
    const placeholder = fieldPlaceholder(field);
    const id = fieldInputId(field);
    const inputType = isDateTime(field) ? "datetime-local" : "date";

    if (readonly || field.readonly) {
      return renderFieldShell(field, fieldReadonlyInput(field, display, "text"), { labelFor: id });
    }

    return renderFieldShell(
      field,
      html`<div class="sum-date-field-inline">
        <input
          id=${id}
          type=${inputType}
          class="sum-field-input sum-date-input"
          name=${field.name}
          value=${native}
          placeholder=${placeholder}
          autocomplete="off"
          @input=${(ev: Event) => {
            record.set(field.name, (ev.target as HTMLInputElement).value || null);
            this.patch();
          }}
          @change=${(ev: Event) => {
            record.set(field.name, (ev.target as HTMLInputElement).value || null);
            record.notifyFieldChange(field.name);
          }}
        />
        <div class="sum-date-field-actions">
          <button
            type="button"
            class="sum-date-action-btn"
            @click=${() => {
              record.set(field.name, todayNative(field));
              record.notifyFieldChange(field.name);
              this.patch();
            }}
          >
            Today
          </button>
          <button
            type="button"
            class="sum-date-action-btn"
            @click=${() => {
              record.set(field.name, null);
              record.notifyFieldChange(field.name);
              this.patch();
            }}
          >
            Clear
          </button>
        </div>
      </div>`,
      { labelFor: id },
    );
  }
}

export { DateField as DateTimeField };
