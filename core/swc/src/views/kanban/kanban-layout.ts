/** Default cards per row when kanban arch omits columns_per_row. */
export const KANBAN_COLUMNS_PER_ROW_DEFAULT = 4;

/** Max cards per row allowed from view arch. */
export const KANBAN_COLUMNS_PER_ROW_MAX = 12;

/** Clamp kanban columns_per_row from arch to a usable grid count. */
export function kanbanColumnsPerRow(raw: unknown): number {
  const n = Number(raw);
  if (!Number.isFinite(n) || n < 1) {
    return KANBAN_COLUMNS_PER_ROW_DEFAULT;
  }
  return Math.min(KANBAN_COLUMNS_PER_ROW_MAX, Math.max(1, Math.round(n)));
}

/** Inline style for `.sum-kanban-columns` grid repeat count. */
export function kanbanColumnsStyle(raw: unknown): string {
  return `--sum-kanban-columns:${kanbanColumnsPerRow(raw)}`;
}
