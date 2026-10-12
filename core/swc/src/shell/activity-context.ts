import type { SwcWorkspacePayload } from "../types/workspace.js";

export const ACTIVITY_CONTEXT = "activity.context";

export interface ActivityContextPayload {
  model: string;
  recordId: number;
  recordMessagesEligible: boolean;
  recordLogEligible: boolean;
}

export function activityContextFromWorkspace(payload: SwcWorkspacePayload): ActivityContextPayload {
  return {
    model: payload.model,
    recordId: payload.recordId,
    recordMessagesEligible: payload.recordMessagesEligible === true,
    recordLogEligible: payload.recordLogEligible === true,
  };
}

export function emptyActivityContext(): ActivityContextPayload {
  return {
    model: "",
    recordId: 0,
    recordMessagesEligible: false,
    recordLogEligible: false,
  };
}

/** Open the activity board, filtered to the current record when one is open. */
export function activityBoardHref(ctx: ActivityContextPayload): string {
  const params = new URLSearchParams({
    action: "mail.action_mail_activity",
    view_type: "activity",
  });
  if (ctx.model && ctx.recordId > 0) {
    params.set(
      "domain",
      JSON.stringify([
        ["model", "=", ctx.model],
        ["res_id", "=", ctx.recordId],
      ]),
    );
  }
  return `/web?${params.toString()}`;
}
