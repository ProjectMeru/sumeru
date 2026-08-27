import type { SwcEnv } from "../runtime/env.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import { instantiateFieldWidget, resolveFieldWidget, type FieldWidgetInstance } from "./registry.js";

interface FieldEntry {
  widget: FieldWidgetInstance;
  readonly: boolean;
  widgetName: string;
}

/** Reuses field widget instances across FormView patches (same record + mode). */
export class FieldHost {
  private readonly env: SwcEnv;
  private readonly entries = new Map<string, FieldEntry>();

  constructor(env: SwcEnv) {
    this.env = env;
  }

  render(field: SwcArchField, record: SwcRecord, readonly: boolean): HTMLElement {
    const widgetName = resolveFieldWidget(field);
    const key = field.name;
    const prev = this.entries.get(key);

    if (prev && prev.readonly === readonly && prev.widgetName === widgetName) {
      return prev.widget.render();
    }

    prev?.widget.destroy();
    const widget = instantiateFieldWidget(this.env, field, record, readonly);
    this.entries.set(key, { widget, readonly, widgetName });
    return widget.render();
  }

  /** Drop one field widget after onchange, or all widgets when `fieldName` is omitted. */
  invalidate(fieldName?: string): void {
    if (!fieldName) {
      this.clear();
      return;
    }
    const prev = this.entries.get(fieldName);
    if (!prev) return;
    prev.widget.destroy();
    this.entries.delete(fieldName);
  }

  clear(): void {
    for (const { widget } of this.entries.values()) {
      widget.destroy();
    }
    this.entries.clear();
  }
}
