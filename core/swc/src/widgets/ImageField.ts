import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import { fieldInputId, renderFieldShell } from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

export class ImageField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const image = String(record.get(field.name) ?? "");
    const hasImage = image.length > 0;

    const id = fieldInputId(field);

    return renderFieldShell(
      field,
      html`<div data-sum-avatar>
        ${hasImage
          ? html`<div class="sum-image-thumb"><img class="sum-image-thumb-img" src=${image} alt="" /></div>`
          : html`<div class="sum-image-thumb sum-image-thumb--empty">No image</div>`}
        ${readonly || field.readonly
          ? html`<input type="hidden" data-sum-image-value name=${field.name} value=${image} />`
          : html`<label class="sum-form-avatar-upload">
              Upload
              <input id=${id} type="file" accept="image/*" />
              <input
                type="hidden"
                data-sum-image-value
                name=${field.name}
                value=${image}
                @input=${(ev: Event) => record.set(field.name, (ev.target as HTMLInputElement).value)}
              />
            </label>`}
      </div>`,
      { modifiers: ["sum-field-widget--image"], labelFor: readonly || field.readonly ? false : id },
    );
  }
}
