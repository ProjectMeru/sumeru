import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import {
  fieldInputId,
  fieldPlaceholder,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";
import { AsyncFieldController, recordDisplayName } from "./field-async.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

interface SelectOption {
  value: string;
  label: string;
}

export class SelectionField extends SwcComponent<FieldProps> {
  private options: SelectOption[] = [];
  private loaded = false;
  private readonly asyncCtrl = new AsyncFieldController(this);

  setup(): void {
    void this.loadOptions();
  }

  onWillUnmount(): void {
    this.asyncCtrl.cancel();
  }

  private async loadOptions(): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const { field, readonly } = this.props;

    if (field.selection?.length) {
      this.options = field.selection.map(([value, label]) => ({ value, label }));
      this.loaded = true;
      this.asyncCtrl.finish(gen);
      return;
    }

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

    const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name"], 200);
    this.options = (rows ?? []).map((row) => ({
      value: String(row.id ?? ""),
      label: String(row.name ?? row.id ?? ""),
    }));
    this.loaded = true;
    this.asyncCtrl.finish(gen);
  }

  private displayValue(): string {
    const { field, record } = this.props;
    const raw = record.get(field.name);
    const id = raw == null || raw === "" ? "" : String(raw);
    if (!id) return "";
    const named = record.get(`${field.name}_name`);
    if (named) return String(named);
    const match = this.options.find((o) => o.value === id);
    return match?.label ?? recordDisplayName(record, field.name);
  }

  template() {
    const { field, record, readonly } = this.props;
    const current = record.get(field.name);
    const currentVal = current == null || current === "" ? "" : String(current);
    const id = fieldInputId(field);

    const placeholder = fieldPlaceholder(field);

    if (readonly || field.readonly) {
      return renderFieldShell(field, fieldReadonlyValue(this.displayValue(), placeholder), { labelFor: false });
    }

    return renderFieldShell(
      field,
      html`<select
        id=${id}
        class="sum-field-input sum-field-select"
        name=${field.name}
        autocomplete="off"
        @change=${(ev: Event) => {
          const val = (ev.target as HTMLSelectElement).value;
          const opt = this.options.find((o) => o.value === val);
          record.set(field.name, val ? Number(val) || val : null);
          if (opt) record.set(`${field.name}_name`, opt.label);
          this.asyncCtrl.refresh();
        }}
      >
        <option value="" disabled=${currentVal !== "" ? "disabled" : false} selected=${currentVal === "" ? "selected" : false}>${placeholder}</option>
        ${this.options.map(
          (opt) =>
            html`<option value=${opt.value} selected=${opt.value === currentVal ? "selected" : ""}>
              ${opt.label}
            </option>`,
        )}
      </select>
      ${!this.loaded ? html`<span class="sum-field-hint">Loading…</span>` : ""}`,
      { labelFor: id },
    );
  }
}
