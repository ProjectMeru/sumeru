import type { SwcViewArch } from "../../types/workspace.js";

/** Parse arch date/datetime field values into a Date, or null when invalid. */
export function parseArchDate(raw: unknown): Date | null {
  const text = String(raw ?? "").trim();
  if (!text) return null;
  const date = new Date(text);
  return Number.isNaN(date.getTime()) ? null : date;
}

type ArchDateSource = "cohort" | "gantt" | "calendar";

/** Resolve the primary date field from arch metadata with view-specific fallback order. */
export function resolveArchDateField(
  arch: SwcViewArch,
  order: ArchDateSource[],
  fallback: string,
): string {
  for (const source of order) {
    const value =
      source === "cohort"
        ? arch.cohort?.dateStart
        : source === "gantt"
          ? arch.gantt?.dateStart
          : arch.calendar?.dateStart;
    if (value) return value;
  }
  const dateField = arch.fields.find((f) => f.type === "date" || f.type === "datetime");
  return dateField?.name ?? fallback;
}
