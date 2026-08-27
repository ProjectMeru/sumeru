import { describe, expect, it, vi } from "vitest";
import { GanttView } from "../../src/views/gantt/GanttView.js";
import { MapView } from "../../src/views/map/MapView.js";
import { CohortView } from "../../src/views/cohort/CohortView.js";
import type { SwcEnv } from "../../src/runtime/env.js";
import type { SwcViewArch, SwcWorkspacePayload } from "../../src/types/workspace.js";

function testEnv(): SwcEnv {
  return {
    bootstrap: {} as never,
    services: {
      action: { openRecord: vi.fn() },
    },
  } as unknown as SwcEnv;
}

function payload(arch: Partial<SwcViewArch>, records: Record<string, unknown>[]): SwcWorkspacePayload {
  return {
    actionId: 1,
    menuId: "2",
    viewType: arch.type ?? "list",
    model: "demo.model",
    recordId: 0,
    formEdit: false,
    csrfToken: "",
    arch: {
      type: arch.type ?? "list",
      model: "demo.model",
      fields: arch.fields ?? [],
      ...arch,
    },
    records,
    viewTabs: [],
    breadcrumbs: [],
  };
}

describe("advanced views", () => {
  it("renders gantt bars from date_start and date_stop", () => {
    const view = new GanttView(
      {
        payload: payload(
          { type: "gantt", title: "Plan", gantt: { dateStart: "date_start", dateStop: "date_stop" } },
          [{ id: 4, name: "Task", date_start: "2026-01-01", date_stop: "2026-01-08" }],
        ),
      },
      testEnv(),
    );
    view.callSetup();
    const el = view.render();
    expect(el.querySelector(".sum-gantt-bar")).toBeTruthy();
    expect(el.textContent).toContain("Task");
    view.destroy();
  });

  it("renders map markers with an OpenStreetMap link", () => {
    const view = new MapView(
      {
        payload: payload(
          { type: "map", map: { latitude: "lat", longitude: "lng" } },
          [{ id: 8, name: "Depot", lat: 12.97, lng: 77.59 }],
        ),
      },
      testEnv(),
    );
    const el = view.render();
    const link = el.querySelector(".sum-map-link") as HTMLAnchorElement;
    expect(link.href).toContain("openstreetmap.org");
    expect(link.href).toContain("mlat=12.97");
    view.destroy();
  });

  it("buckets cohort rows by month", () => {
    const view = new CohortView(
      {
        payload: payload(
          { type: "cohort", cohort: { dateStart: "create_date", interval: "month" } },
          [
            { id: 1, create_date: "2026-01-10" },
            { id: 2, create_date: "2026-01-20" },
            { id: 3, create_date: "2026-02-02" },
          ],
        ),
      },
      testEnv(),
    );
    const el = view.render();
    expect(el.querySelector(".sum-cohort-table")?.textContent).toContain("2026-01");
    expect(el.querySelector(".sum-cohort-table")?.textContent).toContain("2026-02");
    view.destroy();
  });
});
