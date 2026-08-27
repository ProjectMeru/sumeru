import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { fieldInputId, renderFieldShell } from "./field-shell.js";
import type { FieldWidgetProps } from "./field-props.js";
import { stringFromUnknown } from "./field-value.js";
import { inputValueFromEvent } from "./field-events.js";
import { isFieldReadonly } from "../model/modifiers.js";

export class ImageField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field, record, readonly } = this.props;
    const image = stringFromUnknown(record.get(field.name));
    const hasImage = image.length > 0;

    const id = fieldInputId(field);
    const fieldReadonly = isFieldReadonly(field, record, readonly);

    return renderFieldShell(
      field,
      html`<div data-sum-avatar>
        ${hasImage
          ? html`<div class="sum-image-thumb"><img class="sum-image-thumb-img" src=${image} alt="" /></div>`
          : html`<div class="sum-image-thumb sum-image-thumb--empty">No image</div>`}
        ${fieldReadonly
          ? html`<input type="hidden" data-sum-image-value name=${field.name} value=${image} />`
          : html`<label class="sum-form-avatar-upload">
              Upload
              <input id=${id} type="file" accept="image/*" />
              <input
                type="hidden"
                data-sum-image-value
                name=${field.name}
                value=${image}
                @input=${(event: Event) => record.set(field.name, inputValueFromEvent(event))}
              />
            </label>`}
      </div>`,
      { modifiers: ["sum-field-widget--image"], labelFor: fieldReadonly ? false : id },
    );
  }
}
