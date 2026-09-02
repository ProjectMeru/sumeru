import type { SwcArchField } from "../types/workspace.js";
import { booleanFromUnknown } from "./field-value.js";

export function inverseFieldName(parentModel: string): string {
  const part = parentModel.split(".").pop() ?? "parent";
  return `${part}_id`;
}

export function columnsForField(field: SwcArchField): SwcArchField[] {
  return field.subview?.fields ?? [];
}

export function parseCellValue(col: SwcArchField, raw: string): unknown {
  if (raw === "") return null;
  if (col.type === "integer") return Number.parseInt(raw, 10);
  if (col.type === "float" || col.type === "float64" || col.type === "numeric") {
    return Number.parseFloat(raw);
  }
  if (col.type === "boolean") return raw === "true" || raw === "1";
  return raw;
}

export function isNumericType(col: SwcArchField): boolean {
  return (
    col.type === "integer" ||
    col.type === "float" ||
    col.type === "float64" ||
    col.type === "numeric"
  );
}

/** Formats a number with thousand separators (e.g. 12000 → "12,000"). */
export function formatNumericValue(raw: unknown): string {
  const num = Number(raw);
  if (!Number.isFinite(num)) return raw == null ? "" : String(raw);
  const [intPart, decPart] = String(num).split(".");
  const withSep = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  return decPart !== undefined ? `${withSep}.${decPart}` : withSep;
}

export function displayCellValue(col: SwcArchField, line: Record<string, unknown>): string {
  const raw = line[col.name];
  if (raw == null) return "";
  const named = line[`${col.name}_name`];
  if (named != null && String(named) !== "") return String(named);
  if (col.type === "boolean") {
    return booleanFromUnknown(raw) ? "Yes" : "No";
  }
  if (isNumericType(col)) return formatNumericValue(raw);
  return String(raw);
}

/** Server-safe line values: drop id, `*_name`/`*_names` display fields, and empty values. */
export function serverLineValues(data: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const [k, v] of Object.entries(data)) {
    if (k === "id") continue;
    if (k.endsWith("_name") || k.endsWith("_names")) continue;
    if (v == null || v === "") continue;
    out[k] = v;
  }
  return out;
}
