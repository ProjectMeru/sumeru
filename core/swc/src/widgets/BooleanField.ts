import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import {
  fieldInputId,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

function isChecked(val: unknown): boolean {
  return val === true || val === 1 || val === "1" || val === "true";
}

export class BooleanField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const checked = isChecked(record.get(field.name));
    const id = fieldInputId(field);

    if (readonly || field.readonly) {
      return renderFieldShell(field, fieldReadonlyValue(checked ? "Yes" : "No"), { labelFor: false });
    }

    return renderFieldShell(
      field,
      html`<input
        id=${id}
        type="checkbox"
        class="sum-field-input"
        name=${field.name}
        autocomplete="off"
        checked=${checked ? "checked" : ""}
        @change=${(ev: Event) => record.set(field.name, (ev.target as HTMLInputElement).checked)}
      />`,
      { labelFor: id },
    );
  }
}
