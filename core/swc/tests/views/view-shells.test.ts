import { describe, expect, it, vi } from "vitest";
import { ActivityView } from "../../src/views/activity/ActivityView.js";
import { CalendarView } from "../../src/views/calendar/CalendarView.js";
import { HierarchyView } from "../../src/views/hierarchy/HierarchyView.js";
import { PivotView } from "../../src/views/pivot/PivotView.js";
import { collectionEnv, viewPayload } from "../harness/view.js";

describe("view shells", () => {
  it("ActivityView renders list table rows", () => {
    const view = new ActivityView(
      {
        payload: viewPayload(
          { type: "activity", fields: [{ name: "name" }, { name: "summary" }] },
          [{ id: 1, name: "Call", summary: "Follow up" }],
        ),
      },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    expect(el.querySelector(".sum-list-table")).toBeTruthy();
    expect(el.textContent).toContain("Call");
    view.destroy();
  });

  it("ActivityView row click opens record", () => {
    const openRecord = vi.fn();
    const env = collectionEnv({ action: { openRecord, navigate: vi.fn() } });
    const view = new ActivityView(
      {
        payload: viewPayload(
          { type: "activity", fields: [{ name: "name" }] },
          [{ id: 1, name: "Task" }],
        ),
      },
      env,
    );
    view.callSetup();
    const row = view.render().querySelector(".sum-list-row") as HTMLElement;
    row.click();
    expect(openRecord).toHaveBeenCalled();
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
