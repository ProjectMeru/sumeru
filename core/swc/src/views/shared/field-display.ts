import type { SwcArchField } from "../../types/workspace.js";

/** Plain-text display name for a record row (name, display_name, or fallback). */
export function recordDisplayLabel(
  row: Record<string, unknown>,
  fallback?: string | number,
): string {
  const name = row.name ?? row.display_name;
  if (name != null && name !== "") return String(name);
  if (fallback != null) return String(fallback);
  const id = Number(row.id ?? 0);
  return id > 0 ? `#${id}` : "";
}

/** Plain-text cell value for list/kanban surfaces (no field widgets). */
export function formatFieldValue(row: Record<string, unknown>, field: SwcArchField): string {
  if (field.type === "boolean") {
    const raw = row[field.name];
    if (raw === true) return "Yes";
    if (raw === false) return "No";
    return "";
  }

  if (field.type === "selection" && field.selection?.length) {
    const raw = row[field.name];
    if (raw == null || raw === "") return "";
    const key = String(raw);
    const match = field.selection.find(([value]) => value === key);
    return match?.[1] ?? key;
  }

  const raw = row[`${field.name}_name`] ?? row[field.name];
  if (raw == null || raw === false) return "";
  return String(raw);
}
