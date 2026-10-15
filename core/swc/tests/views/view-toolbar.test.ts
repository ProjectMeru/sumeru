import { describe, expect, it, vi } from "vitest";
import {
  buildAllReportEntries,
  buildReportActionEntries,
  createToolbarIcon,
  exportQuery,
  graphExportUrl,
  newRecordUrl,
  pivotExportUrl,
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
    expect(listParams.get("csrf_token")).toBeNull();

    const formParams = exportQuery(basePayload({ recordId: 7 }), "name,email", 7);
    expect(formParams.get("id")).toBe("7");
  });

  it("exportFieldNamesCsv joins arch field names", () => {
    expect(exportFieldNamesCsv([{ name: "a" }, { name: "b" }])).toBe("a,b");
  });

  it("buildAllReportEntries adds pivot export for pivot view type", () => {
    const payload = basePayload({
      viewType: "pivot",
      arch: {
        type: "pivot",
        model: "crm.lead",
        fields: [
          { name: "state", pivotType: "row" },
          { name: "amount", pivotType: "measure" },
        ],
      },
    });
    const entries = buildAllReportEntries(payload, "", 0, "pivot");
    expect(entries.some((e) => e.label === "Export pivot CSV")).toBe(true);
  });

  it("buildReportActionEntries offers background export when list exceeds sync cap", () => {
    const entries = buildReportActionEntries(
      basePayload({
        listTotal: 600,
        arch: {
          type: "list",
          model: "crm.lead",
          fields: [],
          report: { download: true, upload: false, formats: "csv,xlsx,pdf", pdfSizes: "", bulkModes: "" },
        },
      }),
      "name,email",
    );
    const labels = entries.map((e) => e.label);
    expect(labels).toContain("Export all CSV (background)");
    expect(labels).toContain("Export all Excel (background)");
    expect(labels).not.toContain("Export CSV");
    expect(labels.some((l) => l.includes("PDF"))).toBe(true);
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
    const pivotEntries = buildAllReportEntries(payload, "", 0, "pivot");
    expect(pivotEntries.some((e) => e.label === "Export pivot CSV")).toBe(true);
  });
});
