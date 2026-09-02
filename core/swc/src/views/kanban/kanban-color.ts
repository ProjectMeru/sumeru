import type { SwcArchField } from "../../types/workspace.js";

export const KANBAN_COLOR_COUNT = 12;

/** Palette aligned with stage header colors in sumeru-workspace.css */
export const KANBAN_COLOR_HEX: readonly string[] = [
  "#875a7b",
  "#9c9c9c",
  "#67917a",
  "#a06666",
  "#8899aa",
  "#d4a017",
  "#6a5acd",
  "#4682b4",
  "#c0392b",
  "#e67e22",
  "#27ae60",
  "#3498db",
];

export const KANBAN_COLOR_LABELS: readonly string[] = [
  "Plum",
  "Gray",
  "Sage",
  "Rose",
  "Slate",
  "Gold",
  "Violet",
  "Steel",
  "Crimson",
  "Orange",
  "Green",
  "Blue",
];

export function isKanbanColorField(field: SwcArchField): boolean {
  return field.name === "color" || field.widget === "color";
}

function colorIndexFromRow(row: Record<string, unknown>): number | null {
  const raw = row.color;
  if (raw == null || raw === false || raw === "") {
    return null;
  }
  const n = Number(raw);
  if (!Number.isFinite(n) || n < 0 || n >= KANBAN_COLOR_COUNT) {
    return null;
  }
  return n;
}

/** Record color wins; else stage column color; else null (no stripe). */
export function resolveCardColor(row: Record<string, unknown>, stageColor?: number): number | null {
  const recordColor = colorIndexFromRow(row);
  if (recordColor != null) {
    return recordColor;
  }
  if (stageColor != null && stageColor >= 0 && stageColor < KANBAN_COLOR_COUNT) {
    return stageColor;
  }
  return null;
}

export function kanbanStripeClass(colorIndex: number | null): string {
  if (colorIndex == null || colorIndex < 0 || colorIndex >= KANBAN_COLOR_COUNT) {
    return "sum-kanban-card-stripe sum-kanban-card-stripe--none";
  }
  return `sum-kanban-card-stripe sum-kanban-card-stripe--color-${colorIndex}`;
}

export function kanbanColorHasField(fields: SwcArchField[]): boolean {
  return fields.some((f) => isKanbanColorField(f) || f.name === "color");
}
