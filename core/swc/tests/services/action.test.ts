import { describe, expect, it, vi } from "vitest";
import { ActionService } from "../../src/services/action.js";
import { RouterService } from "../../src/services/router.js";
import { ACTION_CLOSED } from "../../src/constants/routes.js";

describe("ActionService", () => {
  it("navigate delegates workspace URLs to router", () => {
    const router = new RouterService();
    const parseSpy = vi.spyOn(router, "parse").mockReturnValue({
      actionId: 1,
      menuId: "2",
      viewType: "list",
      recordId: 0,
      formEdit: false,
      listSearch: "",
      listFilter: "",
      listGroupBy: "",
      listDomain: "",
    });
    const assignRoute = vi.spyOn(router, "assign").mockImplementation(() => undefined);
    const action = new ActionService(router);
    action.navigate("/web?action=1&menu_id=2");
    expect(parseSpy).toHaveBeenCalled();
    expect(assignRoute).toHaveBeenCalled();
  });

  it("openWindowAction builds query params", () => {
    const action = new ActionService();
    const nav = vi.spyOn(action, "navigate").mockImplementation(() => undefined);
    action.openWindowAction(5, "9", { q: "x" });
    expect(nav).toHaveBeenCalledWith(expect.stringContaining("action=5"));
    expect(nav).toHaveBeenCalledWith(expect.stringContaining("menu_id=9"));
    expect(nav).toHaveBeenCalledWith(expect.stringContaining("q=x"));
  });

  it("openRecord pushes route when router present", () => {
    const router = new RouterService();
    const push = vi.spyOn(router, "push").mockImplementation(() => undefined);
    const action = new ActionService(router);
    action.openRecord({ actionId: 1, menuId: "2", recordId: 7 });
    expect(push).toHaveBeenCalledWith(
      expect.objectContaining({ actionId: 1, menuId: "2", recordId: 7, viewType: "form" }),
    );
  });

  it("applyCallResult handles close and emits bus event", async () => {
    const emitted: unknown[] = [];
    const action = new ActionService();
    action.setEnv({
      services: {
        bus: { emit: (_c: string, p: unknown) => emitted.push(p) },
        dialog: { close: vi.fn() },
      },
    } as never);
    const handled = await action.applyCallResult({ close: true });
    expect(handled).toBe(true);
    expect(emitted).toEqual([{}]);
  });

  it("applyCallResult navigates on redirect", async () => {
    const action = new ActionService();
    const nav = vi.spyOn(action, "navigate").mockImplementation(() => undefined);
    await action.applyCallResult({ redirect: "https://example.com/done" });
    expect(nav).toHaveBeenCalledWith("https://example.com/done");
  });

  it("applyCallResult opens current window for non-dialog target", async () => {
    const action = new ActionService();
    const nav = vi.spyOn(action, "navigate").mockImplementation(() => undefined);
    await action.applyCallResult({
      open: { model: "crm.lead", recordId: 3, viewType: "form", target: "current" },
    });
    expect(nav).toHaveBeenCalled();
  });

  it("parseRedirectOpen extracts wizard dialog params", () => {
    const action = new ActionService();
    const parsed = (
      action as unknown as { parseRedirectOpen: (url: string) => { model: string; recordId: number } | null }
    ).parseRedirectOpen("/web?model=crm.lead.lost&id=4&view_type=form");
    expect(parsed?.model).toBe("crm.lead.lost");
    expect(parsed?.recordId).toBe(4);
  });

  it("openFormDialog loads workspace payload and opens host", async () => {
    const openHost = vi.fn().mockResolvedValue(undefined);
    const getJSON = vi.fn().mockResolvedValue({
      arch: { type: "form", model: "wiz.model", title: "Wizard", fields: [] },
      model: "wiz.model",
      actionId: 1,
      menuId: "",
      viewType: "form",
      recordId: 2,
      formEdit: true,
      csrfToken: "t",
      viewTabs: [],
      breadcrumbs: [],
    });
    const action = new ActionService();
    action.setEnv({
      bootstrap: { swcApiBase: "/web/swc" },
      services: {
        http: { getJSON },
        dialog: { openHost, close: vi.fn() },
      },
    } as never);
    await (
      action as unknown as { openFormDialog: (open: Record<string, unknown>) => Promise<void> }
    ).openFormDialog({ model: "wiz.model", recordId: 2, viewType: "form" });
    expect(getJSON).toHaveBeenCalled();
    expect(openHost).toHaveBeenCalledWith("Wizard", expect.any(HTMLElement));
  });

  it("returns false for primitive call results", async () => {
    const action = new ActionService();
    expect(await action.applyCallResult(true)).toBe(false);
    expect(await action.applyCallResult(null)).toBe(false);
  });
});

describe("ACTION_CLOSED constant", () => {
  it("is exported for bus subscribers", () => {
    expect(ACTION_CLOSED).toBeTruthy();
  });
});
