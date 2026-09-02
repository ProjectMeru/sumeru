import type { SwcArchGroup } from "../../types/workspace.js";

export type GroupLayoutContext = "sheet" | "row" | "inner";

export interface GroupRowItem {
  group: SwcArchGroup;
  gridSpan: number;
}

/** Max logical columns for an outer group row (group `col` attribute, else nested child count). */
export function outerGroupMaxCols(group: SwcArchGroup, childCount: number): number {
  if (group.col && group.col > 0) return group.col;
  return Math.max(childCount, 1);
}

export function childGroupColspan(group: SwcArchGroup): number {
  return group.colspan && group.colspan > 0 ? group.colspan : 1;
}

/** Map colspan within a maxCols row onto a 12-column grid span. */
export function gridSpan12(maxCols: number, colspan: number): number {
  const cols = Math.max(maxCols, 1);
  const span = Math.max(colspan, 1);
  return Math.min(12, Math.max(1, Math.round((span * 12) / cols)));
}

export function packGroupRows(parent: SwcArchGroup, nested: SwcArchGroup[]): GroupRowItem[][] {
  const maxCols = outerGroupMaxCols(parent, nested.length);
  const rows: GroupRowItem[][] = [];
  let current: GroupRowItem[] = [];
  let used = 0;

  for (const child of nested) {
    const colspan = childGroupColspan(child);
    if (used + colspan > maxCols && current.length > 0) {
      rows.push(current);
      current = [];
      used = 0;
    }
    current.push({ group: child, gridSpan: gridSpan12(maxCols, colspan) });
    used += colspan;
  }
  if (current.length > 0) rows.push(current);
  return rows;
}

export function groupClassNames(group: SwcArchGroup, ctx: GroupLayoutContext, plain: boolean): string {
  const parts = ["sum-form-group"];
  if (plain || !group.string) {
    parts.push("sum-form-group--plain");
  } else if (ctx === "row" || ctx === "inner") {
    parts.push("sum-form-group--col");
  } else {
    parts.push("sum-form-group--full");
  }
  if ((group.fields ?? []).length > 0) {
    parts.push("sum-form-group--row-layout");
  }
  return parts.join(" ");
}
