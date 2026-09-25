import type { SwcArchButton, SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "./record.js";
import { isDebugMode } from "../devtools/debug.js";

type ModifierTriplet = { invisible: boolean; readonly: boolean; required: boolean };

export interface ModifierViewContext {
  userId?: number;
  companyId?: number;
  context?: Record<string, unknown>;
}

const UNSAFE_EXPR = /[`\\[\];]|=>|\bfunction\b|\bclass\b|\bimport\b|\beval\b|\bnew\b/i;

/** Evaluate a dynamic modifier expression against allowlisted ctx keys only. */
export function evalModifierExpr(
  expr: string | undefined,
  record?: SwcRecord,
  viewCtx?: ModifierViewContext,
): boolean | undefined {
  if (!expr || !record) return undefined;
  const trimmed = expr.trim();
  if (!trimmed) return undefined;
  if (UNSAFE_EXPR.test(trimmed)) {
    if (isDebugMode()) {
      console.warn("[SWC modifiers] rejected expression", trimmed);
    }
    return undefined;
  }

  try {
    const ctx: Record<string, unknown> = {
      ...record.data,
      record: record.data,
      user_id: viewCtx?.userId,
      company_id: viewCtx?.companyId,
      context: viewCtx?.context ?? {},
    };
    const fn = new Function("ctx", `with (ctx) { return !!(${trimmed}); }`);
    return Boolean(fn(ctx));
  } catch {
    if (isDebugMode()) {
      console.warn("[SWC modifiers] failed to evaluate", trimmed);
    }
    return undefined;
  }
}

/** Static arch modifiers plus dynamic overrides from onchange and modifier expressions. */
export function resolveFieldModifiers(
  field: SwcArchField,
  record?: SwcRecord,
  viewCtx?: ModifierViewContext,
): ModifierTriplet {
  const override = record?.modifierOverrides.get(field.name);
  const dynamicInvisible = evalModifierExpr(field.invisible_expr, record, viewCtx);
  const dynamicReadonly = evalModifierExpr(field.readonly_expr, record, viewCtx);
  const dynamicRequired = evalModifierExpr(field.required_expr, record, viewCtx);

  return {
    invisible: override?.invisible ?? dynamicInvisible ?? field.invisible ?? false,
    readonly: override?.readonly ?? dynamicReadonly ?? field.readonly ?? false,
    required: override?.required ?? dynamicRequired ?? field.required ?? false,
  };
}

/** @deprecated Use resolveFieldModifiers */
export function fieldModifiers(field: SwcArchField, record?: SwcRecord): ModifierTriplet {
  return resolveFieldModifiers(field, record);
}

export function isFieldVisible(
  field: SwcArchField,
  record?: SwcRecord,
  viewCtx?: ModifierViewContext,
): boolean {
  return !resolveFieldModifiers(field, record, viewCtx).invisible;
}

export function isFieldReadonly(
  field: SwcArchField,
  record: SwcRecord | undefined,
  viewReadonly: boolean,
  viewCtx?: ModifierViewContext,
): boolean {
  return viewReadonly || resolveFieldModifiers(field, record, viewCtx).readonly;
}

export function fieldDomain(field: SwcArchField, record?: SwcRecord): unknown[] | undefined {
  const fromRecord = record?.fieldDomains.get(field.name);
  if (fromRecord) return fromRecord;
  const raw = field.options?.domain;
  if (!raw) return undefined;
  try {
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return undefined;
    if (!record) return parsed;
    return evalDomainPlaceholders(parsed, record);
  } catch (err) {
    console.warn("fieldDomain: invalid domain JSON", field.name, err);
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

export function isButtonVisible(
  button: SwcArchButton,
  record?: SwcRecord,
  viewCtx?: ModifierViewContext,
): boolean {
  const dynamic = evalModifierExpr(button.invisible_expr, record, viewCtx);
  if (dynamic !== undefined) return !dynamic;
  return !button.invisible;
}
