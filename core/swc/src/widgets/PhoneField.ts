import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import {
  fieldInputId,
  fieldPlaceholder,
  fieldAutocomplete,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

export class PhoneField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const val = String(record.get(field.name) ?? "");
    const placeholder = fieldPlaceholder(field);
    const id = fieldInputId(field);

    if (readonly || field.readonly) {
      return renderFieldShell(field, fieldReadonlyValue(val, placeholder), { labelFor: false });
    }

    return renderFieldShell(
      field,
      html`<input
        id=${id}
        type="tel"
        class="sum-field-input sum-field-phone"
        name=${field.name}
        placeholder=${placeholder}
        value=${val}
        autocomplete=${fieldAutocomplete(field)}
        @input=${(ev: Event) => record.set(field.name, (ev.target as HTMLInputElement).value)}
      />`,
      { labelFor: id },
    );
  }
}
