import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import { DefaultField } from "./DefaultField.js";
import { renderFieldShell, fieldReadonlyValue } from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

export class MonetaryField extends DefaultField {
  template() {
    const { field, record, readonly } = this.props;
    const symbol = field.options?.currency_symbol ?? "¤";
    const val = String(record.get(field.name) ?? "");
    if (readonly || field.readonly) {
      return renderFieldShell(field, fieldReadonlyValue(val ? `${symbol} ${val}` : ""), { labelFor: false });
    }
    return super.template();
  }
}

export class HtmlField extends DefaultField {
  template() {
    const { field, record, readonly } = this.props;
    const raw = String(record.get(field.name) ?? "");
    if (readonly || field.readonly) {
      const text = raw.replace(/<[^>]+>/g, " ").trim();
      return renderFieldShell(field, fieldReadonlyValue(text), { labelFor: false });
    }
    return super.template();
  }
}

export class BinaryField extends SwcComponent<FieldProps> {
  template() {
    const { field, record } = this.props;
    const name = String(record.get(`${field.name}_name`) ?? record.get(field.name) ?? "Download");
    return renderFieldShell(
      field,
      html`<a class="sum-field-link" href="/web/content/${field.name}/${record.id}" download>${name}</a>`,
      { labelFor: false },
    );
  }
}

export class ReferenceField extends DefaultField {}

export class ColorField extends DefaultField {
  template() {
    const { field, record, readonly } = this.props;
    const val = Number(record.get(field.name) ?? 0);
    const swatch = `hsl(${(val * 47) % 360} 70% 45%)`;
    if (readonly || field.readonly) {
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
  template() {
    const { field, record, readonly } = this.props;
    const val = String(record.get(field.name) ?? "");
    if ((readonly || field.readonly) && val) {
      return renderFieldShell(
        field,
        html`<a class="sum-field-link" href=${val} target="_blank" rel="noopener">${val}</a>`,
        { labelFor: false },
      );
    }
    return super.template();
  }
}

export class ProgressField extends DefaultField {
  template() {
    const { field, record, readonly } = this.props;
    const val = Math.min(100, Math.max(0, Number(record.get(field.name) ?? 0)));
    if (readonly || field.readonly) {
      return renderFieldShell(
        field,
        html`<div class="sum-progress">
          <div class="sum-progress-bar" style=${`width:${val}%`}></div>
          <span>${val}%</span>
        </div>`,
        { labelFor: false },
      );
    }
    return super.template();
  }
}

export class HandleField extends SwcComponent<FieldProps> {
  template() {
    const { field } = this.props;
    return renderFieldShell(
      field,
      html`<span class="sum-handle-grip" title="Reorder" aria-hidden="true">⋮⋮</span>`,
      { labelFor: false },
    );
  }
}
