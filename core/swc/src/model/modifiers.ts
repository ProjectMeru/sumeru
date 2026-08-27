import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "./record.js";

type ModifierTriplet = { invisible: boolean; readonly: boolean; required: boolean };

/** Evaluate a dynamic modifier expression against record values. */
export function evalModifierExpr(expr: string | undefined, record?: SwcRecord): boolean | undefined {
  if (!expr || !record) return undefined;
  const trimmed = expr.trim();
  if (!trimmed) return undefined;

  try {
    const ctx: Record<string, unknown> = { ...record.data, record: record.data };
    const fn = new Function("ctx", `with (ctx) { return !!(${trimmed}); }`);
    return Boolean(fn(ctx));
  } catch {
    return undefined;
  }
}

/** Static arch modifiers plus dynamic overrides from onchange and modifier expressions. */
export function fieldModifiers(
  field: SwcArchField,
  record?: SwcRecord,
): ModifierTriplet {
  const override = record?.modifierOverrides.get(field.name);
  const dynamicInvisible = evalModifierExpr(field.invisible_expr, record);
  const dynamicReadonly = evalModifierExpr(field.readonly_expr, record);
  const dynamicRequired = evalModifierExpr(field.required_expr, record);

  return {
    invisible: override?.invisible ?? dynamicInvisible ?? field.invisible ?? false,
    readonly: override?.readonly ?? dynamicReadonly ?? field.readonly ?? false,
    required: override?.required ?? dynamicRequired ?? field.required ?? false,
  };
}

export function isFieldVisible(field: SwcArchField, record?: SwcRecord): boolean {
  return !fieldModifiers(field, record).invisible;
}

export function fieldDomain(field: SwcArchField, record?: SwcRecord): unknown[] | undefined {
  const fromRecord = record?.fieldDomains.get(field.name);
  if (fromRecord) return fromRecord;
  const raw = field.options?.domain;
  if (!raw) return undefined;
  try {
    const parsed = JSON.parse(raw) as unknown[];
    if (!record) return parsed;
    return evalDomainPlaceholders(parsed, record);
  } catch {
    return undefined;
  }
}

function evalDomainPlaceholders(domain: unknown[], record: SwcRecord): unknown[] {
  return domain.map((clause) => {
    if (!Array.isArray(clause)) return clause;
    return clause.map((part) => {
      if (typeof part !== "string") return part;
      if (part.startsWith("$") && part.endsWith("$")) {
        const key = part.slice(1, -1);
        return record.get(key);
      }
      return part;
    });
  });
}

/** Default values for create mode from arch field definitions. */
export function createDefaults(fields: SwcArchField[]): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const f of fields) {
    if (f.default !== undefined) out[f.name] = f.default;
  }
  return out;
}
