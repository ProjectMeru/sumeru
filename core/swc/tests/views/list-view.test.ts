import { beforeEach, describe, expect, it, vi } from "vitest";
import { ListView } from "../../src/views/list/ListView.js";
import { collectionEnv, viewPayload } from "../harness/view.js";
import * as collectionQuery from "../../src/views/shared/collection-query.js";

describe("ListView", () => {
  beforeEach(() => {
    vi.spyOn(collectionQuery, "navigateCollectionQuery").mockImplementation(() => undefined);
  });

  it("renders rows and pagination when list spans pages", () => {
    const view = new ListView(
      {
        payload: {
          ...viewPayload(
            { type: "list", fields: [{ name: "name" }] },
            [
              { id: 1, name: "Alpha" },
              { id: 2, name: "Beta" },
            ],
          ),
          listTotal: 100,
        },
      },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    expect(el.querySelector(".sum-list-table")).toBeTruthy();
    expect(el.textContent).toContain("Alpha");
    expect(el.textContent).toContain("1 /");
    view.destroy();
  });

  it("shows bulk delete when rows are selected", () => {
    const env = collectionEnv();
    const view = new ListView(
      {
        payload: viewPayload({ type: "list", fields: [{ name: "name" }] }, [{ id: 1, name: "A" }]),
      },
      env,
    );
    view.callSetup();
    const el = view.render();
    const checkbox = el.querySelector(".sum-list-select-cell input") as HTMLInputElement;
    checkbox.checked = true;
    checkbox.dispatchEvent(new Event("change", { bubbles: true }));
    expect(view.render().textContent).toContain("Delete (1)");
    view.destroy();
  });

  it("bulk delete calls rpc.unlink and reloads", async () => {
    const env = collectionEnv();
    const unlink = env.services.rpc.unlink as ReturnType<typeof vi.fn>;
    const view = new ListView(
      {
        payload: viewPayload({ type: "list", fields: [{ name: "name" }] }, [{ id: 5, name: "X" }]),
      },
      env,
    );
    view.callSetup();
    const el = view.render();
    const checkbox = el.querySelector(".sum-list-select-cell input") as HTMLInputElement;
    checkbox.checked = true;
    checkbox.dispatchEvent(new Event("change", { bubbles: true }));
    const deleteBtn = view.render().querySelector(".sum-btn--danger") as HTMLButtonElement;
    await deleteBtn.click();
    await vi.waitFor(() => expect(unlink).toHaveBeenCalledWith("demo.model", [5]));
    view.destroy();
  });

  it("renders grouped sections and toggles fold state", () => {
    const view = new ListView(
      {
        payload: {
          ...viewPayload({ type: "list", fields: [{ name: "name" }] }),
          listSections: [
            { value: "open", label: "Open", count: 1, records: [{ id: 3, name: "Todo" }] },
          ],
        },
      },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    expect(el.textContent).toContain("Open");
    const head = el.querySelector(".sum-list-section-head") as HTMLElement;
    head.click();
    expect(view.render().querySelector(".sum-list-section-toggle")?.textContent).toBe("▸");
    view.destroy();
  });

  it("sort header triggers navigateCollectionQuery", () => {
    const view = new ListView(
      {
        payload: viewPayload({ type: "list", fields: [{ name: "name", string: "Name" }] }, [
          { id: 1, name: "A" },
        ]),
      },
      collectionEnv(),
    );
    view.callSetup();
    const sortHeader = view.render().querySelector(".sum-list-th") as HTMLElement;
    sortHeader.click();
    expect(collectionQuery.navigateCollectionQuery).toHaveBeenCalled();
    view.destroy();
  });

  it("select-all toggles every row selection", () => {
    const view = new ListView(
      {
        payload: viewPayload({ type: "list", fields: [{ name: "name" }] }, [
          { id: 1, name: "A" },
          { id: 2, name: "B" },
        ]),
      },
      collectionEnv(),
    );
    view.callSetup();
    const selectAll = view.render().querySelector(".sum-list-select-head input") as HTMLInputElement;
    selectAll.checked = true;
    selectAll.dispatchEvent(new Event("change", { bubbles: true }));
    expect(view.render().textContent).toContain("Delete (2)");
    view.destroy();
  });
});
