import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { fieldLabelId, renderFieldShell } from "./field-shell.js";
import type { FieldWidgetProps } from "./field-props.js";
import { booleanFromUnknown } from "./field-value.js";
import { isFieldReadonly } from "../model/modifiers.js";

export class BooleanRadioField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field, record, readonly } = this.props;
    const checked = booleanFromUnknown(record.get(field.name));
    const name = field.name;
    const fieldReadonly = isFieldReadonly(field, record, readonly);

    return renderFieldShell(
      field,
      html`<div class="sum-field-radio-group" role="radiogroup" aria-labelledby=${fieldLabelId(field)}>
        <label class="sum-field-radio">
          <input
            type="radio"
            name=${name}
            value="1"
            checked=${checked ? "checked" : ""}
            disabled=${fieldReadonly ? "disabled" : undefined}
            @change=${() => !fieldReadonly && record.set(field.name, true)}
          />
          Yes
        </label>
        <label class="sum-field-radio">
          <input
            type="radio"
            name=${name}
            value="0"
            checked=${!checked ? "checked" : ""}
            disabled=${fieldReadonly ? "disabled" : undefined}
            @change=${() => !fieldReadonly && record.set(field.name, false)}
          />
          No
        </label>
      </div>`,
      { labelFor: false },
    );
  }
}
