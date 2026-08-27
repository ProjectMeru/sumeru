import { describe, expect, it } from "vitest";
import {
  exportQuery,
  newRecordUrl,
  renderReportActions,
  renderSearchField,
  visibleFieldNames,
} from "../../src/views/shared/view-toolbar.js";
import type { SwcWorkspacePayload } from "../../src/types/workspace.js";

function basePayload(overrides: Partial<SwcWorkspacePayload> = {}): SwcWorkspacePayload {
  return {
    actionId: 12,
    menuId: "5",
    viewType: "list",
    model: "crm.lead",
    recordId: 0,
    formEdit: false,
    csrfToken: "tok",
    arch: { type: "list", model: "crm.lead", fields: [{ name: "name" }, { name: "email" }] },
    viewTabs: [],
    breadcrumbs: [],
    ...overrides,
  };
}

describe("view-toolbar", () => {
  it("newRecordUrl includes action, menu_id, and view_type=form", () => {
    const url = newRecordUrl(basePayload({ recordId: 99 }));
    expect(url).toContain("action=12");
    expect(url).toContain("menu_id=5");
    expect(url).toContain("view_type=form");
    expect(url).not.toContain("id=99");
  });

  it("exportQuery includes model, action, fields, and optional id", () => {
    const listParams = exportQuery(basePayload(), "name,email");
    expect(listParams.get("model")).toBe("crm.lead");
    expect(listParams.get("action")).toBe("12");
    expect(listParams.get("fields")).toBe("name,email");
    expect(listParams.get("id")).toBeNull();

    const formParams = exportQuery(basePayload({ recordId: 7 }), "name,email", 7);
    expect(formParams.get("id")).toBe("7");
  });

  it("visibleFieldNames joins arch field names", () => {
    expect(visibleFieldNames([{ name: "a" }, { name: "b" }])).toBe("a,b");
  });

  it("renderReportActions returns null when arch.report is absent", () => {
    expect(renderReportActions(basePayload(), "name")).toBeNull();
  });

  it("renderReportActions returns fragment when download enabled", () => {
    const result = renderReportActions(
      basePayload({ arch: { type: "list", model: "crm.lead", fields: [], report: { download: true, upload: false, pdfSizes: "", bulkModes: "" } } }),
      "name,email",
    );
    expect(result).not.toBeNull();
    expect(result!.textContent).toContain("Export CSV");
    expect(result!.textContent).toContain("Export PDF");
    expect(result!.querySelector('a[href*="/web/export/csv"]')).not.toBeNull();
  });

  it("renderSearchField renders search input", () => {
    const root = renderSearchField("acme", () => {}, () => {}).render();
    const input = root.querySelector(".sum-list-search") as HTMLInputElement;
    expect(input).toBeTruthy();
    expect(input.value).toBe("acme");
    expect(input.getAttribute("placeholder")).toBe("Search…");
  });
});
