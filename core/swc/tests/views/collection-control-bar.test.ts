import { describe, expect, it } from "vitest";
import { CollectionBarHost } from "../../src/views/shared/collection-bar-host.js";
import type { SwcEnv } from "../../src/runtime/env.js";
import type { SwcWorkspacePayload } from "../../src/types/workspace.js";

function testEnv(): SwcEnv {
  return {
    bootstrap: {} as never,
    services: {
      action: { navigate: () => undefined },
      router: {
        workspaceUrl: () => "/web?action=1",
        parse: () => ({
          actionId: 1,
          menuId: "2",
          viewType: "list",
          recordId: 0,
          formEdit: false,
          listSearch: "",
          listFilter: "",
          listGroupBy: "",
          listDomain: "",
        }),
      },
      http: {
        postJSON: async () => ({ id: 1 }),
        delete: async () => undefined,
      },
      notification: { success: () => undefined, error: () => undefined },
    },
  } as unknown as SwcEnv;
}

function basePayload(): SwcWorkspacePayload {
  return {
    actionId: 1,
    menuId: "2",
    viewType: "list",
    model: "demo.model",
    recordId: 0,
    formEdit: false,
    csrfToken: "tok",
    arch: {
      type: "list",
      model: "demo.model",
      fields: [{ name: "name" }],
      search: {
        filters: [{ name: "draft", string: "Draft", domain: "[]" }],
        filterFields: [{ name: "name", string: "Name", type: "char" }],
        groupByFields: [{ name: "state", string: "Status", type: "selection" }],
      },
    },
    viewTabs: [],
    breadcrumbs: [],
    favorites: [],
  };
}

describe("CollectionBarHost", () => {
  it("renders search field and separate filter/group buttons beside it", () => {
    const bar = new CollectionBarHost({ payload: basePayload(), viewType: "list" }, testEnv());
    bar.callSetup();
    const el = bar.render();
    expect(el.querySelector(".sum-control-bar-search")).toBeTruthy();
    expect(el.querySelector(".sum-control-bar-search-group")).toBeTruthy();
    expect(el.querySelector(".sum-control-bar-search-wrap .sum-control-bar-action-btn")).toBeFalsy();
    const labels = [...el.querySelectorAll(".sum-control-bar-action-label")].map((n) => n.textContent);
    expect(labels).toContain("Filters");
    expect(labels).toContain("Group By");
    bar.destroy();
  });

  it("opens filters popover as single anchored panel", () => {
    const bar = new CollectionBarHost({ payload: basePayload(), viewType: "list" }, testEnv());
    bar.callSetup();
    document.body.append(bar.render());
    const buttons = [...bar.rootElement!.querySelectorAll(".sum-control-bar-action-btn")] as HTMLButtonElement[];
    const filtersBtn = buttons.find((b) => b.textContent?.includes("Filters"))!;
    filtersBtn.click();
    bar.patch();
    expect(bar.rootElement!.querySelector(".sum-popover--filters")).toBeTruthy();
    expect(bar.rootElement!.querySelectorAll(".sum-popover").length).toBe(1);
    expect(bar.rootElement!.querySelector(".sum-filter-panel")).toBeFalsy();
    bar.destroy();
  });

  it("shows active tags in dedicated tags row", () => {
    const bar = new CollectionBarHost(
      {
        payload: {
          ...basePayload(),
          listFilter: "draft",
          listGroupBy: "state",
        },
        viewType: "list",
      },
      testEnv(),
    );
    bar.callSetup();
    const el = bar.render();
    expect(el.querySelector(".sum-control-bar-tags")).toBeTruthy();
    bar.destroy();
  });
});
