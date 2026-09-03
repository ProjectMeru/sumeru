import { describe, expect, it, vi } from "vitest";
import {
  buildReportActionEntries,
  createToolbarIcon,
  exportQuery,
  graphExportUrl,
  newRecordUrl,
  pivotExportUrl,
  renderGraphExportLink,
  renderPivotExportLink,
  renderReportActions,
  renderSearchField,
  exportFieldNamesCsv,
  renderNewButton,
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

  it("exportFieldNamesCsv joins arch field names", () => {
    expect(exportFieldNamesCsv([{ name: "a" }, { name: "b" }])).toBe("a,b");
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

  it("buildReportActionEntries lists export formats for actions menu", () => {
    const entries = buildReportActionEntries(
      basePayload({
        arch: {
          type: "list",
          model: "crm.lead",
          fields: [],
          report: { download: true, upload: false, formats: "csv,xlsx", pdfSizes: "", bulkModes: "" },
        },
      }),
      "name,email",
    );
    expect(entries.map((e) => e.label)).toEqual(["Export CSV", "Export Excel"]);
  });

  it("renderSearchField renders search input with icon", () => {
    const root = renderSearchField("acme", () => {}, () => {}).render();
    const input = root.querySelector(".sum-list-search") as HTMLInputElement;
    expect(input).toBeTruthy();
    expect(input.value).toBe("acme");
    expect(input.getAttribute("placeholder")).toBe("Search…");
    expect(root.querySelector(".sum-list-search-submit svg")).toBeTruthy();
  });

  it("renderSearchField triggers search callbacks", () => {
    const onSearch = vi.fn();
    const onInput = vi.fn();
    const root = renderSearchField("", onSearch, onInput).render();
    const input = root.querySelector(".sum-list-search") as HTMLInputElement;
    input.value = "x";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    expect(onInput).toHaveBeenCalledWith("x");
    input.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
    root.querySelector(".sum-list-search-submit")?.dispatchEvent(new MouseEvent("click"));
    expect(onSearch).toHaveBeenCalledTimes(2);
  });

  it("renderNewButton renders New link", () => {
    const btn = renderNewButton(basePayload());
    expect(btn.textContent).toBe("New");
    expect(btn.className).toContain("sum-list-btn-new");
  });

  it("createToolbarIcon renders SVG paths for toolbar buttons", () => {
    const icon = createToolbarIcon("filter", "test-icon");
    expect(icon.querySelector("svg.test-icon")).toBeTruthy();
    expect(icon.querySelector("polygon")).toBeTruthy();
  });

  it("createToolbarIcon covers all icon names", () => {
    for (const name of ["search", "filter", "group", "favorite", "download", "chevron", "close"] as const) {
      expect(createToolbarIcon(name).querySelector("svg")).toBeTruthy();
    }
  });

  it("buildReportActionEntries includes upload import form", () => {
    const entries = buildReportActionEntries(
      basePayload({
        arch: {
          type: "list",
          model: "crm.lead",
          fields: [{ name: "name" }],
          report: { download: false, upload: true, formats: "", pdfSizes: "", bulkModes: "" },
        },
      }),
      "name",
    );
    expect(entries.some((e) => e.label === "Import CSV")).toBe(true);
    const importEntry = entries.find((e) => e.label === "Import CSV");
    expect(importEntry?.node.querySelector('input[type="file"]')).toBeTruthy();
  });

  it("renderReportActions wraps upload forms without restyling", () => {
    const result = renderReportActions(
      basePayload({
        arch: {
          type: "list",
          model: "crm.lead",
          fields: [],
          report: { download: true, upload: true, formats: "csv,pdf", pdfSizes: "", bulkModes: "" },
        },
      }),
      "name",
    );
    expect(result?.querySelector(".sum-list-upload-form")).toBeTruthy();
    expect(result?.querySelector('a[href*="/web/export/pdf"]')).toBeTruthy();
  });

  it("pivot and graph export helpers build URLs", () => {
    const payload = basePayload({
      arch: {
        type: "pivot",
        model: "crm.lead",
        fields: [
          { name: "state", pivotType: "row" },
          { name: "amount", pivotType: "measure" },
        ],
      },
    });
    expect(pivotExportUrl(payload, ["state"], ["amount"])).toContain("/web/export/pivot");
    expect(graphExportUrl(payload, "state", "amount")).toContain("/web/export/graph");
    expect(renderPivotExportLink(payload)?.getAttribute("href")).toContain("group_by=state");
    expect(renderGraphExportLink(payload, "state", "amount")?.textContent).toBe("Export CSV");
  });
});
