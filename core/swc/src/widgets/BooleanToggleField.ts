import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import { fieldInputId, renderFieldShell } from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

function isChecked(val: unknown): boolean {
  return val === true || val === 1 || val === "1" || val === "true";
}

export class BooleanToggleField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const checked = isChecked(record.get(field.name));
    const id = fieldInputId(field);

    return renderFieldShell(
      field,
      html`<label class="sum-field-toggle" for=${id}>
        <span class="sum-field-toggle-name">${field.string ?? field.name}</span>
        <input
          id=${id}
          type="checkbox"
          class="sum-field-input"
          name=${field.name}
          autocomplete="off"
          checked=${checked ? "checked" : ""}
          disabled=${readonly || field.readonly ? "disabled" : undefined}
          @change=${(ev: Event) => record.set(field.name, (ev.target as HTMLInputElement).checked)}
        />
        <span>${checked ? "On" : "Off"}</span>
      </label>`,
      { showLabel: false },
    );
  }
}
