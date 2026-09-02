import { describe, expect, it, vi } from "vitest";
import { html } from "../../src/template/html.js";
import { CollectionView } from "../../src/views/shared/collection-view.js";
import type { SwcEnv } from "../../src/runtime/env.js";
import type { SwcWorkspacePayload } from "../../src/types/workspace.js";

class StubCollectionView extends CollectionView {
  protected readonly collectionViewType = "list";

  override template() {
    return html`<div class="sum-stub-collection">${this.collectionBar.renderOrPatch()}</div>`;
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

describe("CollectionView", () => {
  it("mounts and destroys the collection bar on lifecycle hooks", () => {
    const view = new StubCollectionView({ payload: basePayload() }, testEnv());
    view.callSetup();

    expect(view.render().querySelector(".sum-control-bar")).toBeTruthy();

    const destroySpy = vi.spyOn(view["collectionBar"], "destroy");
    view.onWillUnmount();
    expect(destroySpy).toHaveBeenCalled();
  });
});
