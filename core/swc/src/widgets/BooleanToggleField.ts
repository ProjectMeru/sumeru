import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { fieldInputId, renderFieldShell } from "./field-shell.js";
import type { FieldWidgetProps } from "./field-props.js";
import { booleanFromUnknown } from "./field-value.js";
import { checkboxCheckedFromEvent } from "./field-events.js";
import { isFieldReadonly } from "../model/modifiers.js";

export class BooleanToggleField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field, record, readonly } = this.props;
    const checked = booleanFromUnknown(record.get(field.name));
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
          disabled=${isFieldReadonly(field, record, readonly) ? "disabled" : undefined}
          @change=${(event: Event) => record.set(field.name, checkboxCheckedFromEvent(event))}
        />
        <span>${checked ? "On" : "Off"}</span>
      </label>`,
      { showLabel: false },
    );
  }
}
