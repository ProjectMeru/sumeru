import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import {
  fieldInputId,
  fieldPlaceholder,
  fieldAutocomplete,
  fieldReadonlyInput,
  renderFieldShell,
} from "./field-shell.js";
import type { FieldWidgetProps } from "./field-props.js";
import { stringFromUnknown } from "./field-value.js";
import { inputValueFromEvent } from "./field-events.js";
import { isFieldReadonly } from "../model/modifiers.js";

function inputTypeForField(field: SwcArchField): string {
  if (field.widget === "email") return "email";
  if (field.type === "integer" || field.type === "float" || field.type === "numeric") return "number";
  if (field.type === "date") return "date";
  if (field.type === "datetime") return "datetime-local";
  return "text";
}

function stepForField(field: SwcArchField): string | undefined {
  if (field.type === "integer") return "1";
  if (field.type === "float" || field.type === "numeric") return "any";
  return undefined;
}

function parseNumericValue(field: SwcArchField, raw: string): unknown {
  if (raw === "") return null;
  if (field.type === "integer") return Number.parseInt(raw, 10);
  if (field.type === "float" || field.type === "numeric") return Number.parseFloat(raw);
  return raw;
}

export class DefaultField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field, record, readonly } = this.props;
    const fieldValue = stringFromUnknown(record.get(field.name));
    const placeholder = fieldPlaceholder(field);
    const inputType = inputTypeForField(field);
    const step = stepForField(field);
    const id = fieldInputId(field);

    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(
        field,
        field.type === "integer" || field.type === "float" || field.type === "numeric"
          ? fieldReadonlyInput(field, fieldValue, "text")
          : fieldReadonlyInput(field, fieldValue, inputType === "text" ? "text" : inputType),
        { labelFor: id },
      );
    }

    return renderFieldShell(
      field,
      html`<input
        id=${id}
        type=${inputType}
        class="sum-field-input"
        name=${field.name}
        placeholder=${placeholder}
        value=${fieldValue}
        autocomplete=${fieldAutocomplete(field)}
        ${step ? html`step=${step}` : ""}
        @input=${(event: Event) =>
          record.set(field.name, parseNumericValue(field, inputValueFromEvent(event)))}
        @change=${() => record.notifyFieldChange(field.name)}
      />`,
      { labelFor: id },
    );
  }
}
