import type { SwcEnv } from "../../runtime/env.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { VIEW_FORM } from "../../constants/routes.js";

export function recordIdFromRow(row: Record<string, unknown> | number): number {
  if (typeof row === "number") return row;
  return Number(row.id ?? 0);
}

/** Open a workspace record in form view (shared by list, kanban, calendar, etc.). */
export function openWorkspaceRecord(
  env: SwcEnv,
  payload: SwcWorkspacePayload,
  row: Record<string, unknown> | number,
): void {
  const recordId = recordIdFromRow(row);
  if (recordId <= 0) return;
  env.services.action.openRecord({
    actionId: payload.actionId,
    menuId: payload.menuId,
    recordId,
    viewType: VIEW_FORM,
  });
}
