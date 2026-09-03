import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { fieldInputId, renderFieldShell } from "./field-shell.js";
import type { FieldWidgetProps } from "./field-props.js";
import { stringFromUnknown } from "./field-value.js";
import { inputValueFromEvent } from "./field-events.js";
import { isFieldReadonly } from "../model/modifiers.js";
import { isUploadedImageSrc, resolveImageDisplaySrc } from "../views/shared/image-placeholder.js";

export class ImageField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field, record, readonly } = this.props;
    const rawImage = record.get(field.name);
    const ctx = {
      model: record.model,
      gender: record.get("gender"),
      isCompany: record.get("is_company"),
    };
    const displaySrc = resolveImageDisplaySrc(rawImage, ctx);
    const uploaded = typeof rawImage === "string" && isUploadedImageSrc(rawImage);
    const storedImage = uploaded ? stringFromUnknown(rawImage) : "";

    const id = fieldInputId(field);
    const fieldReadonly = isFieldReadonly(field, record, readonly);

    return renderFieldShell(
      field,
      html`<div
        data-sum-avatar
        data-sum-has-upload=${uploaded ? "1" : "0"}
        data-sum-readonly=${fieldReadonly ? "1" : "0"}
      >
        <div
          class="sum-image-thumb sum-image-thumb--clickable${uploaded ? "" : " sum-image-thumb--placeholder"}"
          role="button"
          tabindex="0"
          aria-label=${uploaded ? "View or change image" : "Upload image"}
        >
          <img
            class="sum-image-thumb-img"
            src=${displaySrc}
            alt=""
            ${uploaded ? "" : 'data-sum-image-placeholder="1"'}
          />
          ${fieldReadonly || uploaded
            ? ""
            : html`<span class="sum-image-upload-hint">Click to upload</span>`}
          <input id=${id} type="file" class="sum-image-file-input" accept="image/*" hidden />
        </div>
        <input
          type="hidden"
          data-sum-image-value
          name=${field.name}
          value=${storedImage}
          @input=${(event: Event) => record.set(field.name, inputValueFromEvent(event))}
        />
      </div>`,
      { modifiers: ["sum-field-widget--image"], labelFor: false },
    );
  }
}
