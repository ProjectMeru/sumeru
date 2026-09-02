import { describe, expect, it, vi } from "vitest";
import {
  renderActionsPopover,
  renderFavoritesPopover,
  renderFiltersPopover,
  renderGroupPopover,
  renderPopoverItem,
} from "../../src/views/shared/collection-bar-panels.js";
import type { SwcWorkspacePayload } from "../../src/types/workspace.js";

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
      report: { download: true, formats: "csv" },
    },
    viewTabs: [],
    breadcrumbs: [],
    favorites: [],
  };
}

describe("collection-bar-panels", () => {
  it("renders filter and group popover markup", () => {
    const filters = renderFiltersPopover(
      {
        query: { search: "", presetFilters: ["draft"], customDomain: "", groupBy: [] },
        customField: "name",
        customOp: "=",
        customValue: "",
        domainPresets: [{ name: "draft", string: "Draft", domain: "[]" }],
        filterFields: [{ name: "name", string: "Name", type: "char" }],
      },
      {
        onTogglePreset: vi.fn(),
        onCustomFieldChange: vi.fn(),
        onCustomOpChange: vi.fn(),
        onCustomValueInput: vi.fn(),
        onApplyCustom: vi.fn(),
      },
    ).render();
    expect(filters.querySelector(".sum-popover--filters")).toBeTruthy();
    expect(filters.textContent).toContain("Draft");

    const group = renderGroupPopover(
      {
        query: { search: "", presetFilters: [], customDomain: "", groupBy: ["state"] },
        groupPresets: [{ name: "by_state", string: "Status", groupBy: "state" }],
        groupByFields: [{ name: "user_id", string: "User", type: "many2one" }],
      },
      { onToggleGroupBy: vi.fn() },
    ).render();
    expect(group.querySelector(".sum-popover--group")).toBeTruthy();
    expect(group.textContent).toContain("Status");
  });

  it("renderPopoverItem marks active entries", () => {
    const el = renderPopoverItem("Active", true, () => undefined).render();
    expect(el.querySelector(".sum-popover-item--active")).toBeTruthy();
    expect(el.textContent).toContain("✓");
  });

  it("renders favorites and actions popovers", () => {
    const favorites = renderFavoritesPopover(
      {
        favorites: [{ id: 1, name: "Mine", isShared: false }],
        saveName: "New",
        saveShared: true,
        savingFavorite: false,
      },
      {
        onApplyFavorite: vi.fn(),
        onDeleteFavorite: vi.fn(),
        onSaveNameInput: vi.fn(),
        onSaveSharedChange: vi.fn(),
        onSaveFavorite: vi.fn(),
      },
    ).render();
    expect(favorites.querySelector(".sum-popover--favorites")).toBeTruthy();
    expect(favorites.textContent).toContain("Mine");

    const actions = renderActionsPopover(basePayload(), "name").render();
    expect(actions.querySelector(".sum-popover--actions")).toBeTruthy();
    expect(actions.textContent).toContain("Export CSV");
  });
});
