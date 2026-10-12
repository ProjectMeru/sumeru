import { describe, expect, it, vi } from "vitest";
import { ActivityView } from "../../src/views/activity/ActivityView.js";
import { activityBoardHref } from "../../src/shell/activity-context.js";
import { collectionEnv, viewPayload } from "../harness/view.js";

describe("activity board actions", () => {
  it("marks an activity done through the object action", async () => {
    const callMethod = vi.fn().mockResolvedValue({});
    const env = collectionEnv();
    env.services.rpc.callMethod = callMethod;
    env.services.action.applyCallResult = vi.fn().mockResolvedValue(false);
    const view = new ActivityView(
      {
        payload: viewPayload({ type: "activity", fields: [{ name: "summary" }] }, [
          { id: 4, summary: "Call", date_deadline: "", user_id_name: "Ada", model: "crm.lead", res_id: 3 },
        ]),
      },
      env,
    );
    view.callSetup();
    const el = view.render();
    expect(el.querySelector('[data-activity-column="later"]')?.textContent).toContain("Undated");
    (el.querySelector("[data-activity-done]") as HTMLButtonElement).click();
    await Promise.resolve();
    expect(callMethod).toHaveBeenCalledWith("mail.activity", "action_done", 4, undefined);
    view.destroy();
  });

  it("snoozes through the object action and keeps the activity id", async () => {
    const callMethod = vi.fn().mockResolvedValue({});
    const env = collectionEnv();
    env.services.rpc.callMethod = callMethod;
    env.services.action.applyCallResult = vi.fn().mockResolvedValue(false);
    const view = new ActivityView(
      {
        payload: viewPayload({ type: "activity", fields: [{ name: "summary" }] }, [
          { id: 8, summary: "Later", date_deadline: "2026-01-31", user_id_name: "Ada", model: "crm.lead", res_id: 3 },
        ]),
      },
      env,
    );
    view.callSetup();
    (view.render().querySelector("[data-activity-snooze]") as HTMLButtonElement).click();
    await Promise.resolve();
    expect(callMethod).toHaveBeenCalledWith("mail.activity", "action_snooze", 8, undefined);
    view.destroy();
  });

  it("uses the calendar day key when the arch sets date_start", () => {
    const view = new ActivityView(
      {
        payload: viewPayload(
          { type: "activity", calendar: { dateStart: "date_deadline" }, fields: [{ name: "summary" }] },
          [{ id: 1, summary: "Meet", date_deadline: "2026-10-12", model: "crm.lead", res_id: 2 }],
        ),
      },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    expect(el.querySelector(".sum-calendar-title")).toBeTruthy();
    expect(el.querySelector('[data-day="2026-10-12"]')?.textContent).toContain("Meet");
    view.destroy();
  });

  it("opens the unfiltered board when no record is selected", () => {
    expect(activityBoardHref({ model: "", recordId: 0, recordMessagesEligible: false, recordLogEligible: false })).toBe(
      "/web?action=mail.action_mail_activity&view_type=activity",
    );
  });
});
