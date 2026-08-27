import type { SwcArchField } from "../../types/workspace.js";

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
