import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { DefaultField } from "./DefaultField.js";
import { renderFieldShell, fieldReadonlyValue, shellOptions } from "./field-shell.js";
import type { FieldWidgetProps } from "./field-props.js";
import { stringFromUnknown } from "./field-value.js";
import { isFieldReadonly } from "../model/modifiers.js";

function monetarySymbol(field: FieldWidgetProps["field"], record: FieldWidgetProps["record"]): string {
  const staticSymbol = field.options?.currency_symbol?.trim();
  if (staticSymbol) return staticSymbol;
  const currencyField = field.options?.currency_field?.trim();
  if (currencyField) {
    const named = record.get(`${currencyField}_name`);
    if (named != null && named !== "") return String(named);
    const sym = record.get(`${currencyField}_symbol`);
    if (sym != null && sym !== "") return String(sym);
  }
  return "¤";
}

export class MonetaryField extends DefaultField {
  override template() {
    const { field, record, readonly } = this.props;
    const symbol = monetarySymbol(field, record);
    const fieldValue = stringFromUnknown(record.get(field.name));
    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(
        field,
        fieldReadonlyValue(fieldValue ? `${symbol} ${fieldValue}` : ""),
        shellOptions(record, { labelFor: false }),
      );
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
      return renderFieldShell(field, fieldReadonlyValue(text), shellOptions(record, { labelFor: false }));
    }
    return super.template();
  }
}

export class BinaryField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field, record, modelName } = this.props;
    const name = stringFromUnknown(record.get(`${field.name}_name`) ?? record.get(field.name) ?? "Download");
    const model = modelName ?? record.model ?? "";
    const href =
      model && record.id > 0
        ? `/web/content/${encodeURIComponent(model)}/${encodeURIComponent(field.name)}/${record.id}`
        : "#";
    return renderFieldShell(
      field,
      html`<a class="sum-field-link" href=${href} download>${name}</a>`,
      shellOptions(record, { labelFor: false }),
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
        shellOptions(record, { labelFor: false }),
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
        shellOptions(record, { labelFor: false }),
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
        shellOptions(record, { labelFor: false }),
      );
    }
    return super.template();
  }
}

export class HandleField extends SwcComponent<FieldWidgetProps> {
  override template() {
    const { field, record } = this.props;
    return renderFieldShell(
      field,
      html`<span class="sum-handle-grip" title="Reorder" aria-hidden="true">⋮⋮</span>`,
      shellOptions(record, { labelFor: false }),
    );
  }
}
