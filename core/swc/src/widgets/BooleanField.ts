import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import {
  fieldInputId,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";
import type { FieldWidgetProps } from "./field-props.js";
import { booleanFromUnknown } from "./field-value.js";
import { checkboxCheckedFromEvent } from "./field-events.js";
import { isFieldReadonly } from "../model/modifiers.js";

export class BooleanField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field, record, readonly } = this.props;
    const checked = booleanFromUnknown(record.get(field.name));
    const id = fieldInputId(field);

    if (isFieldReadonly(field, record, readonly)) {
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
        @change=${(event: Event) => record.set(field.name, checkboxCheckedFromEvent(event))}
      />`,
      { labelFor: id },
    );
  }
}
