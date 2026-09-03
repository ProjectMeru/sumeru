import { describe, expect, it, vi } from "vitest";
import { html } from "../../src/template/html.js";
import { CollectionView } from "../../src/views/shared/collection-view.js";
import { renderCollectionShell } from "../../src/views/shared/collection-layout.js";
import type { SwcEnv } from "../../src/runtime/env.js";
import type { SwcWorkspacePayload } from "../../src/types/workspace.js";

class StubCollectionView extends CollectionView {
  protected readonly collectionViewType = "list";

  override template() {
    return html`<div class="sum-stub-collection">${this.collectionBar.renderOrPatch()}</div>`;
  }
}

class ShellCollectionView extends CollectionView {
  protected readonly collectionViewType = "kanban";

  override template() {
    return this.renderShell(html`<p class="sum-stub-body">content</p>`);
  }
}

function testEnv(): SwcEnv {
  return {
    bootstrap: {} as never,
    services: {
      action: { navigate: () => undefined },
      router: { workspaceUrl: () => "/web" },
      http: { postJSON: async () => ({}), delete: async () => undefined },
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
    arch: { type: "list", model: "demo.model", fields: [{ name: "name" }] },
    viewTabs: [],
    breadcrumbs: [],
    favorites: [],
  };
}

describe("renderCollectionShell", () => {
  it("emits sum-collection-view and sum-{type}-view wrapper classes", () => {
    const bar = {
      renderOrPatch: () => {
        const el = document.createElement("div");
        el.className = "sum-control-bar";
        return el;
      },
    };
    const root = renderCollectionShell("kanban", bar as never, html`<p class="sum-body">x</p>`).render();
    const shell = root.matches(".sum-collection-view")
      ? root
      : root.querySelector(".sum-collection-view");
    expect(shell).toBeTruthy();
    expect(shell!.classList.contains("sum-kanban-view")).toBe(true);
    expect(shell!.querySelector(".sum-control-bar")).toBeTruthy();
    expect(shell!.querySelector(".sum-body")?.textContent).toBe("x");
  });
});

describe("CollectionView", () => {
  it("mounts and destroys the collection bar on lifecycle hooks", () => {
    const view = new StubCollectionView({ payload: basePayload() }, testEnv());
    view.callSetup();

    expect(view.render().querySelector(".sum-control-bar")).toBeTruthy();

    const destroySpy = vi.spyOn(view["collectionBar"], "destroy");
    view.onWillUnmount();
    expect(destroySpy).toHaveBeenCalled();
  });

  it("renderShell nests control bar and body inside collection wrapper", () => {
    const view = new ShellCollectionView({ payload: basePayload() }, testEnv());
    view.callSetup();
    const root = view.render();
    const shell = root.matches(".sum-collection-view")
      ? root
      : root.querySelector(".sum-collection-view");
    expect(shell).toBeTruthy();
    expect(shell!.classList.contains("sum-kanban-view")).toBe(true);
    expect(shell!.querySelector(".sum-control-bar")).toBeTruthy();
    expect(shell!.querySelector(".sum-stub-body")?.textContent).toBe("content");
  });
});
