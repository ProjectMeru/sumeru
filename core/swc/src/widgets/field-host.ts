import type { SwcEnv } from "../runtime/env.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import { instantiateFieldWidget, resolveFieldWidget, type FieldWidgetInstance } from "./registry.js";

interface FieldEntry {
  comp: FieldWidgetInstance;
  readonly: boolean;
  widget: string;
}

/** Reuses field widget instances across FormView patches (same record + mode). */
export class FieldHost {
  private readonly env: SwcEnv;
  private readonly entries = new Map<string, FieldEntry>();

  constructor(env: SwcEnv) {
    this.env = env;
  }

  render(field: SwcArchField, record: SwcRecord, readonly: boolean): HTMLElement {
    const widget = resolveFieldWidget(field);
    const key = field.name;
    const prev = this.entries.get(key);

    if (prev && prev.readonly === readonly && prev.widget === widget) {
      return prev.comp.render();
    }

    prev?.comp.destroy();
    const comp = instantiateFieldWidget(this.env, field, record, readonly);
    this.entries.set(key, { comp, readonly, widget });
    return comp.render();
  }

  clear(): void {
    for (const { comp } of this.entries.values()) {
      comp.destroy();
    }
    this.entries.clear();
  }
}
