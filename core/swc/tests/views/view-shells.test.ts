import { describe, expect, it, vi } from "vitest";
import { ActivityView } from "../../src/views/activity/ActivityView.js";
import { CalendarView } from "../../src/views/calendar/CalendarView.js";
import { HierarchyView } from "../../src/views/hierarchy/HierarchyView.js";
import { IframeView } from "../../src/views/iframe/IframeView.js";
import { PivotView } from "../../src/views/pivot/PivotView.js";
import { collectionEnv, viewPayload } from "../harness/view.js";

describe("view shells", () => {
  it("IframeView embeds src from payload", () => {
    const payload = viewPayload({ type: "iframe" }, []);
    payload.viewType = "iframe";
    payload.iframeUrl = "/account/reports/view?type=profit_loss";
    const view = new IframeView({ payload }, collectionEnv());
    view.callSetup();
    const frame = view.render().querySelector("iframe") as HTMLIFrameElement;
    expect(frame?.getAttribute("src")).toBe("/account/reports/view?type=profit_loss");
    view.destroy();
  });

  it("IframeView shows error when iframeUrl missing", () => {
    const view = new IframeView({ payload: viewPayload({ type: "iframe" }, []) }, collectionEnv());
    view.callSetup();
    expect(view.render().textContent).toContain("Missing iframe URL");
    view.destroy();
  });

  it("IframeView updates src when payload changes", () => {
    const payload = viewPayload({ type: "iframe" }, []);
    payload.iframeUrl = "/a";
    const view = new IframeView({ payload }, collectionEnv());
    view.callSetup();
    expect(view.render().querySelector("iframe")?.getAttribute("src")).toBe("/a");
    const next = viewPayload({ type: "iframe" }, []);
    next.iframeUrl = "/b";
    view.updateProps({ payload: next });
    expect(view.render().querySelector("iframe")?.getAttribute("src")).toBe("/b");
    view.destroy();
  });

  it("ActivityView renders overdue, today, and later columns", () => {
    const today = new Date();
    const y = today.getFullYear();
    const m = String(today.getMonth() + 1).padStart(2, "0");
    const d = String(today.getDate()).padStart(2, "0");
    const due = `${y}-${m}-${d}`;
    const view = new ActivityView(
      {
        payload: viewPayload({ type: "activity", fields: [{ name: "summary" }] }, [
          { id: 1, summary: "Late", date_deadline: "2000-01-01", user_id: 2, user_id_name: "Ada", model: "crm.lead", res_id: 9 },
          { id: 2, summary: "Now", date_deadline: due, user_id: 2, user_id_name: "Ada", model: "crm.lead", res_id: 9 },
        ]),
      },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    expect(el.querySelector('[data-activity-column="overdue"]')?.textContent).toContain("Late");
    expect(el.querySelector('[data-activity-column="today"]')?.textContent).toContain("Now");
    expect(el.textContent).toContain("Ada");
    view.destroy();
  });

  it("ActivityView card opens the parent record", () => {
    const navigate = vi.fn();
    const env = collectionEnv({ action: { openRecord: vi.fn(), navigate } });
    const view = new ActivityView(
      {
        payload: viewPayload({ type: "activity", fields: [{ name: "summary" }] }, [
          { id: 1, summary: "Call", model: "crm.lead", res_id: 9, user_id_name: "Ada" },
        ]),
      },
      env,
    );
    view.callSetup();
    const card = view.render().querySelector(".sum-kanban-card") as HTMLElement;
    card.click();
    expect(navigate).toHaveBeenCalledWith("/web?model=crm.lead&view_type=form&id=9");
    view.destroy();
  });

  it("ActivityView empty board has no fake row", () => {
    const view = new ActivityView(
      { payload: viewPayload({ type: "activity", fields: [{ name: "summary" }] }, []) },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    expect(el.textContent).toContain("No activities");
    expect(el.querySelector(".sum-kanban-card")).toBeNull();
    view.destroy();
  });

  it("CalendarView uses arch title when provided", () => {
    const view = new CalendarView(
      {
        payload: viewPayload({ type: "calendar", title: "My Calendar" }, []),
      },
      collectionEnv(),
    );
    view.callSetup();
    expect(view.render().querySelector(".sum-calendar-title")?.textContent).toBe("My Calendar");
    view.destroy();
  });

  it("CalendarView places events on the month grid", () => {
    const now = new Date();
    const iso = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-15`;
    const view = new CalendarView(
      {
        payload: viewPayload(
          { type: "calendar", calendar: { dateStart: "date_deadline" } },
          [{ id: 2, name: "Deadline", date_deadline: iso }],
        ),
      },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    expect(el.querySelector(".sum-calendar-grid")).toBeTruthy();
    expect(el.querySelector(".sum-calendar-event")?.textContent).toContain("Deadline");
    view.destroy();
  });

  it("CalendarView shifts month on toolbar navigation", () => {
    const view = new CalendarView(
      { payload: viewPayload({ type: "calendar" }, []) },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    const titleBefore = el.querySelector(".sum-calendar-title")?.textContent ?? "";
    el.querySelectorAll(".sum-calendar-toolbar button")[1]?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    const el2 = view.render();
    const titleAfter = el2.querySelector(".sum-calendar-title")?.textContent ?? "";
    expect(titleAfter).not.toBe(titleBefore);
    view.destroy();
  });

  it("HierarchyView indents child rows by parent depth", () => {
    const view = new HierarchyView(
      {
        payload: viewPayload(
          { type: "hierarchy", hierarchy: { parentField: "parent_id" }, fields: [{ name: "name" }] },
          [
            { id: 1, name: "Root", parent_id: 0 },
            { id: 2, name: "Child", parent_id: 1 },
          ],
        ),
      },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    const cells = el.querySelectorAll(".sum-list-td");
    expect(cells.length).toBeGreaterThan(0);
    expect(el.textContent).toContain("Root");
    expect(el.textContent).toContain("Child");
    view.destroy();
  });

  it("PivotView toggles collapsed rows and shows empty state", () => {
    const empty = new PivotView({ payload: viewPayload({ type: "pivot" }, []) }, collectionEnv());
    empty.callSetup();
    expect(empty.render().querySelector(".sum-pivot-view--empty")).toBeTruthy();
    empty.destroy();

    const view = new PivotView(
      {
        payload: viewPayload({
          type: "pivot",
          pivot: {
            rowLabels: ["Sales"],
            colLabels: ["Q1"],
            values: { Sales: { Q1: 42 } },
            measureLabel: "Amount",
          },
        }),
      },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    expect(el.querySelector(".sum-pivot-table")?.textContent).toContain("42");
    const toggle = el.querySelector(".sum-pivot-row-toggle") as HTMLButtonElement;
    toggle.click();
    expect(view.render().querySelector(".sum-pivot-row-toggle")?.textContent?.trim()).toBe("+");
    view.destroy();
  });
});
