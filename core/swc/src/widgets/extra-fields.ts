import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { DefaultField } from "./DefaultField.js";
import { renderFieldShell, fieldReadonlyValue } from "./field-shell.js";
import type { FieldWidgetProps } from "./field-props.js";
import { stringFromUnknown } from "./field-value.js";
import { isFieldReadonly } from "../model/modifiers.js";

export class MonetaryField extends DefaultField {
  override template() {
    const { field, record, readonly } = this.props;
    const symbol = field.options?.currency_symbol ?? "¤";
    const fieldValue = stringFromUnknown(record.get(field.name));
    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(field, fieldReadonlyValue(fieldValue ? `${symbol} ${fieldValue}` : ""), {
        labelFor: false,
      });
    }
    return super.template();
  }
}

export class HtmlField extends DefaultField {
  override template() {
    const { field, record, readonly } = this.props;
    const raw = stringFromUnknown(record.get(field.name));
    if (isFieldReadonly(field, record, readonly)) {
      const text = raw.replace(/<[^>]+>/g, " ").trim();
      return renderFieldShell(field, fieldReadonlyValue(text), { labelFor: false });
    }
    return super.template();
  }
}

export class BinaryField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field, record } = this.props;
    const name = stringFromUnknown(record.get(`${field.name}_name`) ?? record.get(field.name) ?? "Download");
    return renderFieldShell(
      field,
      html`<a class="sum-field-link" href="/web/content/${field.name}/${record.id}" download>${name}</a>`,
      { labelFor: false },
    );
  }
}

export class ColorField extends DefaultField {
  override template() {
    const { field, record, readonly } = this.props;
    const fieldValue = Number(record.get(field.name) ?? 0);
    const swatch = `hsl(${(fieldValue * 47) % 360} 70% 45%)`;
    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(
        field,
        html`<span class="sum-color-swatch" style=${`background:${swatch}`}></span>`,
        { labelFor: false },
      );
    }
    return super.template();
  }
}

export class UrlField extends DefaultField {
  override template() {
    const { field, record, readonly } = this.props;
    const fieldValue = stringFromUnknown(record.get(field.name));
    if (isFieldReadonly(field, record, readonly) && fieldValue) {
      return renderFieldShell(
        field,
        html`<a class="sum-field-link" href=${fieldValue} target="_blank" rel="noopener">${fieldValue}</a>`,
        { labelFor: false },
      );
    }
    return super.template();
  }
}

export class ProgressField extends DefaultField {
  override template() {
    const { field, record, readonly } = this.props;
    const fieldValue = Math.min(100, Math.max(0, Number(record.get(field.name) ?? 0)));
    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(
        field,
        html`<div class="sum-progress">
          <div class="sum-progress-bar" style=${`width:${fieldValue}%`}></div>
          <span>${fieldValue}%</span>
        </div>`,
        { labelFor: false },
      );
    }
    return super.template();
  }
}

export class HandleField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field } = this.props;
    return renderFieldShell(
      field,
      html`<span class="sum-handle-grip" title="Reorder" aria-hidden="true">⋮⋮</span>`,
      { labelFor: false },
    );
  }
}
