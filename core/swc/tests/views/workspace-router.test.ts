import { describe, expect, it, vi, afterEach } from "vitest";
import { WorkspaceRouter } from "../../src/views/workspace/WorkspaceRouter.js";
import { ShellLayout } from "../../src/shell/ShellLayout.js";
import { ACTION_CLOSED, RECORD_UPDATED } from "../../src/constants/routes.js";
import { RouterService } from "../../src/services/router.js";
import { BusService } from "../../src/services/bus.js";
import type { SwcEnv } from "../../src/runtime/env.js";
import type { SwcWorkspacePayload } from "../../src/types/workspace.js";

function workspacePayload(): SwcWorkspacePayload {
  return {
    actionId: 1,
    menuId: "2",
    viewType: "list",
    model: "demo.model",
    recordId: 0,
    formEdit: false,
    csrfToken: "",
    arch: {
      type: "list",
      model: "demo.model",
      fields: [{ name: "name", type: "char", label: "Name" }],
    },
    records: [{ id: 1, name: "Alpha" }],
    viewTabs: [],
    breadcrumbs: [],
  };
}

function routerEnv(getJSON = vi.fn().mockResolvedValue(workspacePayload())): SwcEnv {
  const bus = new BusService();
  const router = new RouterService();
  return {
    bootstrap: { swcApiBase: "/web/swc", busEnabled: false },
    services: {
      http: { getJSON },
      bus,
      router: {
        parse: () => router.parse({ search: "?action=1&menu_id=2&view_type=list" }),
      },
      action: { openRecord: vi.fn() },
    },
  } as unknown as SwcEnv;
}

describe("WorkspaceRouter", () => {
  afterEach(() => {
    document.body.innerHTML = "";
  });

  it("callSetup and render without requiring hook host during setup", async () => {
    const router = new WorkspaceRouter({}, routerEnv());
    expect(() => router.callSetup()).not.toThrow();
    document.body.append(router.render());
    await vi.waitFor(() => {
      expect(document.body.querySelector(".sum-workspace-view")).toBeTruthy();
    });
    router.destroy();
  });

  it("registers popstate and bus listeners on mount", () => {
    const addSpy = vi.spyOn(window, "addEventListener");
    const removeSpy = vi.spyOn(window, "removeEventListener");
    const bus = new BusService();
    const subscribeSpy = vi.spyOn(bus, "subscribe");
    const env = routerEnv();
    env.services.bus = bus;

    const router = new WorkspaceRouter({}, env);
    router.callSetup();
    router.render();

    expect(addSpy).toHaveBeenCalledWith("popstate", expect.any(Function));
    expect(subscribeSpy).toHaveBeenCalledWith(ACTION_CLOSED, expect.any(Function));
    expect(subscribeSpy).toHaveBeenCalledWith(RECORD_UPDATED, expect.any(Function));

    router.destroy();
    expect(removeSpy).toHaveBeenCalledWith("popstate", expect.any(Function));

    addSpy.mockRestore();
    removeSpy.mockRestore();
  });

  it("reload refreshes payload after bus event", async () => {
    const getJSON = vi.fn().mockResolvedValue(workspacePayload());
    const env = routerEnv(getJSON);
    const router = new WorkspaceRouter({}, env);
    router.callSetup();
    router.render();
    await vi.waitFor(() => expect(getJSON).toHaveBeenCalled());
    getJSON.mockClear();
    router.reload();
    await vi.waitFor(() => expect(getJSON).toHaveBeenCalled());
    router.destroy();
  });

  it("shows error flash when workspace fetch fails", async () => {
    const router = new WorkspaceRouter(
      {},
      routerEnv(vi.fn().mockRejectedValue(new Error("network down"))),
    );
    router.callSetup();
    document.body.append(router.render());
    await vi.waitFor(() => {
      expect(document.body.querySelector(".sum-flash--error")?.textContent).toContain("network down");
    });
    router.destroy();
  });

  it("renders shell page for home route", () => {
    const env = routerEnv();
    env.bootstrap = { swcApiBase: "/web/swc", user: { name: "Admin" }, apps: [] } as never;
    env.services.router = {
      parse: () => ({ shell: "home" as const, search: "" }),
    } as never;
    const router = new WorkspaceRouter({}, env);
    router.callSetup();
    document.body.append(router.render());
    expect(document.body.querySelector(".sum-workspace-root--shell")).toBeTruthy();
    router.destroy();
  });

  it("ShellLayout mounts WorkspaceRouter via callSetup", async () => {
    const layout = new ShellLayout({}, routerEnv());
    expect(() => layout.callSetup()).not.toThrow();
    document.body.append(layout.render());
    await vi.waitFor(() => {
      expect(document.body.querySelector(".sum-workspace-view .sum-list-view")).toBeTruthy();
    });
    layout.destroy();
  });
});
