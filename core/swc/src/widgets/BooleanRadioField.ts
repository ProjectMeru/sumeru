import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import { fieldLabelId, renderFieldShell } from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

function isChecked(val: unknown): boolean {
  return val === true || val === 1 || val === "1" || val === "true";
}

export class BooleanRadioField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const checked = isChecked(record.get(field.name));
    const name = field.name;

    return renderFieldShell(
      field,
      html`<div class="sum-field-radio-group" role="radiogroup" aria-labelledby=${fieldLabelId(field)}>
        <label class="sum-field-radio">
          <input
            type="radio"
            name=${name}
            value="1"
            checked=${checked ? "checked" : ""}
            disabled=${readonly || field.readonly ? "disabled" : undefined}
            @change=${() => !readonly && record.set(field.name, true)}
          />
          Yes
        </label>
        <label class="sum-field-radio">
          <input
            type="radio"
            name=${name}
            value="0"
            checked=${!checked ? "checked" : ""}
            disabled=${readonly || field.readonly ? "disabled" : undefined}
            @change=${() => !readonly && record.set(field.name, false)}
          />
          No
        </label>
      </div>`,
      { labelFor: false },
    );
  }
}
