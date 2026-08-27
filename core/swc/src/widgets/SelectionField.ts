import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import {
  fieldInputId,
  fieldPlaceholder,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";
import { AsyncFieldController, recordDisplayName } from "./field-async.js";
import type { FieldWidgetProps } from "./field-props.js";
import { inputValueFromEvent } from "./field-events.js";
import { isFieldReadonly } from "../model/modifiers.js";

interface SelectOption {
  value: string;
  label: string;
}

export class SelectionField extends SwcComponent<FieldWidgetProps> {
  private options: SelectOption[] = [];
  private loaded = false;
  private readonly asyncCtrl = new AsyncFieldController(this);

  override setup(): void {
    void this.loadOptions();
  }

  override onWillUnmount(): void {
    this.asyncCtrl.cancel();
  }

  private async loadOptions(): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const { field, record, readonly } = this.props;

    if (field.selection?.length) {
      this.options = field.selection.map(([value, label]) => ({ value, label }));
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
      return;
    }

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

    const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name"], 200);
    this.options = (rows ?? []).map((row) => ({
      value: String(row.id ?? ""),
      label: String(row.name ?? row.id ?? ""),
    }));
    this.loaded = true;
    this.asyncCtrl.commitIfCurrent(gen);
  }

  private displayValue(): string {
    const { field, record } = this.props;
    const rawValue = record.get(field.name);
    const id = rawValue == null || rawValue === "" ? "" : String(rawValue);
    if (!id) return "";
    const named = record.get(`${field.name}_name`);
    if (named) return String(named);
    const match = this.options.find((option) => option.value === id);
    return match?.label ?? recordDisplayName(record, field.name);
  }

  override template() {
    const { field, record, readonly } = this.props;
    const current = record.get(field.name);
    const currentVal = current == null || current === "" ? "" : String(current);
    const id = fieldInputId(field);

    const placeholder = fieldPlaceholder(field);

    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(field, fieldReadonlyValue(this.displayValue(), placeholder), { labelFor: false });
    }

    return renderFieldShell(
      field,
      html`<select
        id=${id}
        class="sum-field-input sum-field-select"
        name=${field.name}
        autocomplete="off"
        @change=${(event: Event) => {
          const fieldValue = inputValueFromEvent(event);
          const option = this.options.find((o) => o.value === fieldValue);
          record.set(field.name, fieldValue ? Number(fieldValue) || fieldValue : null);
          if (option) record.set(`${field.name}_name`, option.label);
          this.asyncCtrl.refresh();
        }}
      >
        <option value="" disabled=${currentVal !== "" ? "disabled" : false} selected=${currentVal === "" ? "selected" : false}>${placeholder}</option>
        ${this.options.map(
          (option) =>
            html`<option value=${option.value} selected=${option.value === currentVal ? "selected" : ""}>
              ${option.label}
            </option>`,
        )}
      </select>
      ${!this.loaded ? html`<span class="sum-field-hint">Loading…</span>` : ""}`,
      { labelFor: id },
    );
  }
}
