import { describe, expect, it, vi } from "vitest";
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
  it("renders search field and segmented filter controls with icons and labels", () => {
    const bar = new CollectionBarHost({ payload: basePayload(), viewType: "list" }, testEnv());
    bar.callSetup();
    const el = bar.render();
    expect(el.querySelector(".sum-control-bar-search")).toBeTruthy();
    expect(el.querySelector(".sum-control-bar-search-submit svg")).toBeTruthy();
    expect(el.querySelector(".sum-control-bar-search-area")).toBeTruthy();
    expect(el.querySelector(".sum-control-bar-segment")).toBeTruthy();
    const filterBtn = el.querySelector('[aria-label="Filters"]') as HTMLButtonElement;
    const groupBtn = el.querySelector('[aria-label="Group By"]') as HTMLButtonElement;
    const favBtn = el.querySelector('[aria-label="Favorites"]') as HTMLButtonElement;
    expect(filterBtn).toBeTruthy();
    expect(groupBtn).toBeTruthy();
    expect(favBtn).toBeTruthy();
    expect(filterBtn.getAttribute("title")).toContain("Filter records");
    expect(filterBtn.classList.contains("sum-control-bar-segment-btn")).toBe(true);
    expect(filterBtn.querySelector(".sum-control-bar-segment-icon")).toBeTruthy();
    expect(filterBtn.querySelector(".sum-control-bar-segment-label")?.textContent).toBe("Filters");
    expect(groupBtn.querySelector(".sum-control-bar-segment-label")?.textContent).toBe("Group By");
    bar.destroy();
  });

  it("clear button empties search input and removes q from navigation", () => {
    let navigated = "";
    const env = testEnv();
    env.services.action.navigate = (url: string) => {
      navigated = url;
    };
    const bar = new CollectionBarHost(
      { payload: { ...basePayload(), listSearch: "acme" }, viewType: "list" },
      env,
    );
    bar.callSetup();
    document.body.append(bar.render());
    const input = bar.rootElement!.querySelector(".sum-control-bar-search") as HTMLInputElement;
    expect(input.value).toBe("acme");
    const clear = bar.rootElement!.querySelector(".sum-control-bar-search-clear") as HTMLButtonElement;
    expect(clear).toBeTruthy();
    clear.click();
    const cleared = bar.rootElement!.querySelector(".sum-control-bar-search") as HTMLInputElement;
    expect(cleared.value).toBe("");
    expect(navigated).not.toContain("q=acme");
    expect(bar.rootElement!.querySelector(".sum-control-bar-search-clear")).toBeFalsy();
    bar.destroy();
  });

  it("search submit button is present and clickable", () => {
    const bar = new CollectionBarHost({ payload: basePayload(), viewType: "list" }, testEnv());
    bar.callSetup();
    document.body.append(bar.render());
    const submit = bar.rootElement!.querySelector(".sum-control-bar-search-submit") as HTMLButtonElement;
    expect(submit).toBeTruthy();
    expect(submit.getAttribute("title")).toBe("Search records");
    submit.click();
    bar.destroy();
  });

  it("opens filters popover as single anchored panel", () => {
    const bar = new CollectionBarHost({ payload: basePayload(), viewType: "list" }, testEnv());
    bar.callSetup();
    document.body.append(bar.render());
    const buttons = [...bar.rootElement!.querySelectorAll(".sum-control-bar-segment-btn")] as HTMLButtonElement[];
    const filtersBtn = buttons.find((b) => b.getAttribute("aria-label") === "Filters")!;
    filtersBtn.click();
    bar.patch();
    expect(bar.rootElement!.querySelector(".sum-popover--filters")).toBeTruthy();
    expect(bar.rootElement!.querySelectorAll(".sum-popover").length).toBe(1);
    expect(bar.rootElement!.querySelector(".sum-filter-panel")).toBeFalsy();
    bar.destroy();
  });

  it("shows active filter and group tags inside the search bar", () => {
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
    const tagsRow = el.querySelector(".sum-control-bar-search-tags");
    expect(tagsRow).toBeTruthy();
    expect(el.querySelector(".sum-control-bar-tags")).toBeFalsy();
    expect(tagsRow!.textContent).toContain("Draft");
    expect(tagsRow!.textContent).toContain("Group:");
    bar.destroy();
  });

  it("opens group and favorites popovers and applies favorite", () => {
    let navigated = "";
    const env = testEnv();
    env.services.action.navigate = (url: string) => {
      navigated = url;
    };
    const bar = new CollectionBarHost(
      {
        payload: {
          ...basePayload(),
          favorites: [{ id: 1, name: "My search", domain: "[]", isShared: false }],
        },
        viewType: "list",
      },
      env,
    );
    bar.callSetup();
    document.body.append(bar.render());
    const buttons = [...bar.rootElement!.querySelectorAll(".sum-control-bar-segment-btn")] as HTMLButtonElement[];
    buttons.find((b) => b.getAttribute("aria-label") === "Group By")!.click();
    bar.patch();
    expect(bar.rootElement!.querySelector(".sum-popover--group")).toBeTruthy();
    buttons.find((b) => b.getAttribute("aria-label") === "Favorites")!.click();
    bar.patch();
    expect(bar.rootElement!.querySelector(".sum-popover--favorites")).toBeTruthy();
    bar.rootElement!.querySelector<HTMLButtonElement>(".sum-popover--favorites .sum-popover-item")?.click();
    expect(navigated).toBeTruthy();
    bar.destroy();
  });

  it("actions menu lists export entries when report enabled", () => {
    const bar = new CollectionBarHost(
      {
        payload: {
          ...basePayload(),
          arch: {
            ...basePayload().arch,
            report: { download: true, upload: false, formats: "csv", pdfSizes: "", bulkModes: "" },
          },
        },
        viewType: "list",
      },
      testEnv(),
    );
    bar.callSetup();
    document.body.append(bar.render());
    const actionsBtn = bar.rootElement!.querySelector('[aria-label="Actions"]') as HTMLButtonElement;
    expect(actionsBtn).toBeTruthy();
    actionsBtn.click();
    bar.patch();
    expect(bar.rootElement!.querySelector(".sum-popover--actions")).toBeTruthy();
    expect(bar.rootElement!.textContent).toContain("Export CSV");
    bar.destroy();
  });

  it("applies custom filter from filters popover", () => {
    let navigated = "";
    const env = testEnv();
    env.services.action.navigate = (url: string) => {
      navigated = url;
    };
    env.services.router.workspaceUrl = (opts: { listDomain?: string }) =>
      `/web?action=1&domain=${encodeURIComponent(opts.listDomain ?? "")}`;
    const bar = new CollectionBarHost({ payload: basePayload(), viewType: "list" }, env);
    bar.callSetup();
    document.body.append(bar.render());
    bar.rootElement!.querySelector<HTMLButtonElement>('[aria-label="Filters"]')?.click();
    bar.patch();
    const valueInput = bar.rootElement!.querySelector(".sum-popover-input") as HTMLInputElement;
    valueInput.value = "Acme";
    valueInput.dispatchEvent(new Event("input", { bubbles: true }));
    bar.rootElement!.querySelector<HTMLButtonElement>(".sum-popover-custom .sum-btn")?.click();
    expect(navigated).toContain("domain=");
    bar.destroy();
  });

  it("saves and deletes favorites", async () => {
    const postJSON = vi.fn().mockResolvedValue({ id: 2 });
    const del = vi.fn().mockResolvedValue(undefined);
    const env = testEnv();
    env.services.http = { postJSON, delete: del } as never;
    const bar = new CollectionBarHost(
      {
        payload: {
          ...basePayload(),
          favorites: [{ id: 9, name: "Old", domain: "[]", isShared: true }],
        },
        viewType: "list",
      },
      env,
    );
    bar.callSetup();
    document.body.append(bar.render());
    bar.rootElement!.querySelector<HTMLButtonElement>('[aria-label="Favorites"]')?.click();
    bar.patch();
    const nameInput = bar.rootElement!.querySelector(".sum-popover--favorites .sum-popover-input") as HTMLInputElement;
    nameInput.value = "My fav";
    nameInput.dispatchEvent(new Event("input", { bubbles: true }));
    bar.rootElement!.querySelector<HTMLButtonElement>(".sum-popover--favorites .sum-btn")?.click();
    await vi.waitFor(() => expect(postJSON).toHaveBeenCalled());
    bar.rootElement!.querySelector<HTMLButtonElement>(".sum-popover-fav-delete")?.click();
    await vi.waitFor(() => expect(del).toHaveBeenCalled());
    bar.destroy();
  });
});
