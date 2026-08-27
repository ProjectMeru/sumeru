import { registry, type RegistryEntry } from "../runtime/registry.js";
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
  ReferenceField,
  ColorField,
  UrlField,
  ProgressField,
  HandleField,
} from "./extra-fields.js";

export function registerDefaultWidgets(): void {
  const fields = registry.category("fields");
  const add = (key: string, Ctor: unknown) => fields.add(key, Ctor as RegistryEntry);

  add("default", DefaultField);
  add("char", DefaultField);
  add("email", DefaultField);
  add("integer", DefaultField);
  add("float", DefaultField);
  add("numeric", DefaultField);
  add("date", DateField);
  add("datetime", DateTimeField);
  add("json", TextareaField);
  add("many2one", Many2OneField);
  add("one2many", One2ManyField);
  add("many2many", Many2ManyTagsField);
  add("selection", SelectionField);
  add("boolean", BooleanField);
  add("text", TextareaField);
  add("statusbar", StatusbarField);
  add("priority", PriorityField);
  add("phone", PhoneField);
  add("radio", BooleanRadioField);
  add("boolean_toggle", BooleanToggleField);
  add("many2many_tags", Many2ManyTagsField);
  add("image", ImageField);
  add("monetary", MonetaryField);
  add("html", HtmlField);
  add("binary", BinaryField);
  add("reference", ReferenceField);
  add("color", ColorField);
  add("url", UrlField);
  add("progress", ProgressField);
  add("handle", HandleField);
}

const WIDGET_MAP: Record<string, string> = {
  many2many_tags: "many2many_tags",
  boolean_toggle: "boolean_toggle",
  radio: "radio",
  phone: "phone",
  image: "image",
  selection: "selection",
  email: "email",
  statusbar: "statusbar",
  priority: "priority",
  monetary: "monetary",
  html: "html",
  binary: "binary",
  reference: "reference",
  color: "color",
  url: "url",
  progressbar: "progress",
  progress: "progress",
  handle: "handle",
};

const TYPE_MAP: Record<string, string> = {
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
};

export function resolveFieldWidget(field: SwcArchField): string {
  if (field.widget && WIDGET_MAP[field.widget]) return WIDGET_MAP[field.widget];
  if (field.type && TYPE_MAP[field.type]) return TYPE_MAP[field.type];
  return field.widget ?? field.type ?? "default";
}

export type FieldWidgetInstance = {
  render(): HTMLElement;
  destroy(): void;
  setup?(): void;
};

export function instantiateFieldWidget(
  env: SwcEnv,
  field: SwcArchField,
  record: SwcRecord,
  readonly: boolean,
): FieldWidgetInstance {
  const key = resolveFieldWidget(field);
  const Ctor = (registry.get("fields", key) ?? registry.get("fields", "default")) as unknown as typeof DefaultField;
  const comp = new Ctor({ field, record, readonly }, env);
  comp.setup?.();
  return comp;
}

export function renderField(
  env: SwcEnv,
  field: SwcArchField,
  record: SwcRecord,
  readonly: boolean,
): HTMLElement {
  return instantiateFieldWidget(env, field, record, readonly).render();
}
