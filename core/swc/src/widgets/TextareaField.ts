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

export class TextareaField extends SwcComponent<FieldProps> {
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
      html`<textarea
        id=${id}
        class="sum-field-textarea"
        name=${field.name}
        placeholder=${placeholder}
        autocomplete=${fieldAutocomplete(field)}
        rows="5"
        @input=${(ev: Event) => record.set(field.name, (ev.target as HTMLTextAreaElement).value)}
      >${val}</textarea>`,
      { labelFor: id },
    );
  }
}
