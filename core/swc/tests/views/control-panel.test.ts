import { describe, expect, it, vi } from "vitest";
import {
  parseFilterCSV,
  renderControlPanel,
  renderRowCheckbox,
  renderSearchFilters,
  renderSelectAllHeader,
  renderSortHeader,
  toggleFilterName,
} from "../../src/views/list/control-panel.js";
import type { SwcWorkspacePayload } from "../../src/types/workspace.js";

function payload(overrides: Partial<SwcWorkspacePayload> = {}): SwcWorkspacePayload {
  return {
    actionId: 1,
    menuId: "1",
    viewType: "list",
    model: "m",
    recordId: 0,
    formEdit: false,
    csrfToken: "t",
    arch: { type: "list", model: "m", fields: [] },
    viewTabs: [],
    breadcrumbs: [],
    listTotal: 100,
    records: [],
    ...overrides,
  };
}

describe("control-panel helpers", () => {
  it("parseFilterCSV splits and trims", () => {
    expect(parseFilterCSV("a, b ,c")).toEqual(["a", "b", "c"]);
    expect(parseFilterCSV()).toEqual([]);
  });

  it("toggleFilterName adds and removes names", () => {
    expect(toggleFilterName(["a"], "b")).toEqual(["a", "b"]);
    expect(toggleFilterName(["a", "b"], "a")).toEqual(["b"]);
  });

  it("renderControlPanel shows pager when multiple pages", () => {
    const onPage = vi.fn();
    const el = renderControlPanel({
        payload: payload({ listTotal: 50 }),
        state: { search: "", offset: 0, limit: 10, selectedIds: new Set(), filters: [] },
        onPage,
      }).render();
    expect(el.textContent).toContain("1 / 5");
    el.querySelectorAll("button")[1]?.dispatchEvent(new MouseEvent("click"));
    expect(onPage).toHaveBeenCalledWith(10);
  });

  it("renderControlPanel returns empty when single page", () => {
    const el = renderControlPanel({
        payload: payload({ listTotal: 5 }),
        state: { search: "", offset: 0, limit: 10, selectedIds: new Set(), filters: [] },
        onPage: vi.fn(),
      }).render();
    expect(el.textContent).toBe("");
  });

  it("renderSearchFilters renders domain and group chips", () => {
    const onToggle = vi.fn();
    const el = renderSearchFilters({
        filters: [
          { name: "draft", string: "Draft", domain: "[]" },
          { name: "by_state", string: "State", groupBy: "state" },
        ],
        active: ["draft"],
        onToggle,
      }).render();
    expect(el.textContent).toContain("Draft");
    expect(el.textContent).toContain("Group");
    el.querySelector(".sum-search-chip--active")?.dispatchEvent(new MouseEvent("click"));
    expect(onToggle).toHaveBeenCalledWith("draft");
  });

  it("renderSortHeader toggles sort marker", () => {
    const onSort = vi.fn();
    const el = renderSortHeader({ name: "name", string: "Name" }, "-name", onSort).render();
    expect(el.textContent).toContain("↓");
    el.dispatchEvent(new MouseEvent("click"));
    expect(onSort).toHaveBeenCalledWith("name");
  });

  it("renderRowCheckbox and select-all header fire callbacks", () => {
    const onToggle = vi.fn();
    const row = renderRowCheckbox(5, false, onToggle).render();
    const input = row.querySelector("input") as HTMLInputElement;
    input.checked = true;
    input.dispatchEvent(new Event("change", { bubbles: true }));
    expect(onToggle).toHaveBeenCalledWith(5, true);

    const onToggleAll = vi.fn();
    const head = renderSelectAllHeader(false, onToggleAll).render();
    const all = head.querySelector("input") as HTMLInputElement;
    all.checked = true;
    all.dispatchEvent(new Event("change", { bubbles: true }));
    expect(onToggleAll).toHaveBeenCalledWith(true);
  });
});
