import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import {
  fieldInputId,
  fieldPlaceholder,
  fieldAutocomplete,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";
import type { FieldWidgetProps } from "./field-props.js";
import { stringFromUnknown } from "./field-value.js";
import { inputValueFromEvent } from "./field-events.js";
import { isFieldReadonly } from "../model/modifiers.js";

export class TextareaField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field, record, readonly } = this.props;
    const fieldValue = stringFromUnknown(record.get(field.name));
    const placeholder = fieldPlaceholder(field);
    const id = fieldInputId(field);

    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(field, fieldReadonlyValue(fieldValue, placeholder), { labelFor: false });
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
        @input=${(event: Event) => record.set(field.name, inputValueFromEvent(event))}
      >${fieldValue}</textarea>`,
      { labelFor: id },
    );
  }
}
