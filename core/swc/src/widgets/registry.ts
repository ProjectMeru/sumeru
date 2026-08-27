import { registry, type FieldWidgetConstructor } from "../runtime/registry.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcEnv } from "../runtime/env.js";
import type { SwcRecord } from "../model/record.js";
import { DefaultField } from "./DefaultField.js";
import { Many2OneField } from "./Many2OneField.js";
import { StatusbarField } from "./StatusbarField.js";
import { PriorityField } from "./PriorityField.js";
import { BooleanField } from "./BooleanField.js";
import { TextareaField } from "./TextareaField.js";
import { SelectionField } from "./SelectionField.js";
import { PhoneField } from "./PhoneField.js";
import { BooleanRadioField } from "./BooleanRadioField.js";
import { BooleanToggleField } from "./BooleanToggleField.js";
import { Many2ManyTagsField } from "./Many2ManyTagsField.js";
import { One2ManyField } from "./One2ManyField.js";
import { DateField, DateTimeField } from "./DateField.js";
import { ImageField } from "./ImageField.js";
import {
  MonetaryField,
  HtmlField,
  BinaryField,
  ColorField,
  UrlField,
  ProgressField,
  HandleField,
} from "./extra-fields.js";

const FIELD_CONSTRUCTORS = {
  default: DefaultField,
  char: DefaultField,
  email: DefaultField,
  integer: DefaultField,
  float: DefaultField,
  numeric: DefaultField,
  date: DateField,
  datetime: DateTimeField,
  json: TextareaField,
  many2one: Many2OneField,
  one2many: One2ManyField,
  many2many: Many2ManyTagsField,
  selection: SelectionField,
  boolean: BooleanField,
  text: TextareaField,
  statusbar: StatusbarField,
  priority: PriorityField,
  phone: PhoneField,
  radio: BooleanRadioField,
  boolean_toggle: BooleanToggleField,
  many2many_tags: Many2ManyTagsField,
  image: ImageField,
  monetary: MonetaryField,
  html: HtmlField,
  binary: BinaryField,
  reference: DefaultField,
  color: ColorField,
  url: UrlField,
  progress: ProgressField,
  handle: HandleField,
} satisfies Record<string, FieldWidgetConstructor>;

export function registerDefaultWidgets(): void {
  const fields = registry.category("fields");
  for (const [key, WidgetConstructor] of Object.entries(FIELD_CONSTRUCTORS)) {
    fields.add(key, WidgetConstructor);
  }
}

/** Widget names that alias a different registered constructor. */
const WIDGET_ALIASES = {
  progressbar: "progress",
} satisfies Record<string, string>;

const TYPE_MAP = {
  boolean: "boolean",
  text: "text",
  many2one: "many2one",
  one2many: "one2many",
  many2many: "many2many_tags",
  selection: "selection",
  date: "date",
  datetime: "datetime",
  integer: "integer",
  float: "float",
  numeric: "numeric",
} satisfies Record<string, string>;

export function resolveFieldWidget(field: SwcArchField): string {
  const widget = field.widget ?? "";
  if (widget && widget in WIDGET_ALIASES) {
    return WIDGET_ALIASES[widget as keyof typeof WIDGET_ALIASES];
  }
  if (widget) return widget;
  if (field.type && field.type in TYPE_MAP) {
    return TYPE_MAP[field.type as keyof typeof TYPE_MAP];
  }
  return field.type ?? "default";
}

export type FieldWidgetInstance = {
  render(): HTMLElement;
  renderOrPatch(): HTMLElement;
  destroy(): void;
  callSetup(): void;
};

export function instantiateFieldWidget(
  env: SwcEnv,
  field: SwcArchField,
  record: SwcRecord,
  readonly: boolean,
): FieldWidgetInstance {
  const key = resolveFieldWidget(field);
  const WidgetConstructor =
    registry.get("fields", key) ?? registry.get("fields", "default") ?? DefaultField;
  const widget = new WidgetConstructor({ field, record, readonly }, env);
  widget.callSetup();
  return widget;
}

export function renderField(
  env: SwcEnv,
  field: SwcArchField,
  record: SwcRecord,
  readonly: boolean,
): HTMLElement {
  return instantiateFieldWidget(env, field, record, readonly).render();
}
