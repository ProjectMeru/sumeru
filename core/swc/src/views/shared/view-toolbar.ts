import { html, type TemplateResult } from "../../template/html.js";
import type { SwcArchField, SwcWorkspacePayload } from "../../types/workspace.js";
import { inputValueFromEvent } from "../../widgets/field-events.js";
import {
  BULK_TEMPLATE_ROUTE,
  BULK_UPLOAD_ROUTE,
  EXPORT_CSV_ROUTE,
  EXPORT_PDF_ROUTE,
  EXPORT_XLSX_ROUTE,
  EXPORT_PIVOT_ROUTE,
  EXPORT_GRAPH_ROUTE,
  EXPORT_ASYNC_ROUTE,
  MAX_SYNC_EXPORT_ROWS,
  VIEW_FORM,
  VIEW_GRAPH,
  VIEW_PIVOT,
} from "../../constants/routes.js";
import { RouterService } from "../../services/router.js";
import { graphAxes, pivotFields } from "./arch-fields.js";

function bulkJobFormUrl(jobId: number): string {
  const params = new URLSearchParams({
    model: "sys.bulk.import",
    view_type: VIEW_FORM,
    id: String(jobId),
  });
  return `/web?${params.toString()}`;
}

function parseDownloadFilename(contentDisposition: string | null): string {
  if (!contentDisposition) return "export";
  const star = /filename\*=UTF-8''([^;\s]+)/i.exec(contentDisposition);
  if (star?.[1]) {
    try {
      return decodeURIComponent(star[1]);
    } catch {
      return star[1];
    }
  }
  const quoted = /filename="([^"]+)"/i.exec(contentDisposition);
  if (quoted?.[1]) return quoted[1];
  const plain = /filename=([^;\s]+)/i.exec(contentDisposition);
  if (plain?.[1]) return plain[1].replace(/^"|"$/g, "");
  return "export";
}

async function downloadProtectedGET(url: string, csrfToken: string, failureLabel: string): Promise<void> {
  const headers: Record<string, string> = {};
  if (csrfToken) headers["X-CSRF-Token"] = csrfToken;
  const res = await fetch(url, { method: "GET", credentials: "same-origin", headers });
  if (!res.ok) {
    window.alert(`${failureLabel} failed: ${await res.text()}`);
    return;
  }
  const blob = await res.blob();
  const filename = parseDownloadFilename(res.headers.get("Content-Disposition"));
  const objectUrl = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = objectUrl;
  anchor.download = filename;
  anchor.rel = "noopener";
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(objectUrl);
}

function protectedDownloadButton(
  url: string,
  label: string,
  csrfToken: string,
  className = "sum-popover-menu-link",
): HTMLElement {
  const btn = document.createElement("button");
  btn.type = "button";
  btn.className = className;
  btn.textContent = label;
  btn.addEventListener("click", () => void downloadProtectedGET(url, csrfToken, label));
  return btn;
}

async function queueAsyncExport(
  payload: SwcWorkspacePayload,
  fields: string,
  format: "csv" | "xlsx",
): Promise<void> {
  const query = exportQuery(payload, fields);
  const body = new URLSearchParams();
  body.set("csrf_token", payload.csrfToken);
  body.set("model", payload.model);
  body.set("fields", fields);
  body.set("format", format);
  const res = await fetch(`${EXPORT_ASYNC_ROUTE}?${query.toString()}`, {
    method: "POST",
    credentials: "same-origin",
    headers: { "X-CSRF-Token": payload.csrfToken },
    body,
  });
  if (!res.ok) {
    window.alert(`Export could not be queued: ${await res.text()}`);
    return;
  }
  const data = (await res.json()) as { jobId?: number };
  const jobId = data.jobId ?? 0;
  if (jobId <= 0) {
    window.alert("Export queued but no job id was returned.");
    return;
  }
  const open = window.confirm(
    `Export queued (job #${jobId}). Background jobs need the platform bulk cron. Open the job record to download when ready?`,
  );
  if (open) {
    window.location.href = bulkJobFormUrl(jobId);
  }
}

function asyncExportButton(
  label: string,
  payload: SwcWorkspacePayload,
  fields: string,
  format: "csv" | "xlsx",
): HTMLElement {
  const btn = document.createElement("button");
  btn.type = "button";
  btn.className = "sum-popover-menu-link";
  btn.textContent = label;
  btn.addEventListener("click", () => void queueAsyncExport(payload, fields, format));
  return btn;
}

function linkButton(href: string, label: string, className = "sum-btn sum-btn--secondary"): HTMLElement {
  const a = document.createElement("a");
  a.className = className;
  a.href = href;
  a.textContent = label;
  return a;
}

export function exportFieldNamesCsv(fields: SwcArchField[]): string {
  return fields
    .map((f) => f.name)
    .filter(Boolean)
    .join(",");
}

export function newRecordUrl(payload: SwcWorkspacePayload): string {
  return RouterService.buildUrl({
    actionId: payload.actionId,
    menuId: payload.menuId,
    viewType: VIEW_FORM,
  });
}

export function exportQuery(
  payload: SwcWorkspacePayload,
  fields: string,
  recordId = 0,
): URLSearchParams {
  const params = new URLSearchParams();
  params.set("model", payload.model);
  if (payload.actionId > 0) params.set("action", String(payload.actionId));
  if (fields) params.set("fields", fields);
  if (recordId > 0) params.set("id", String(recordId));
  if (payload.listFilter) params.set("filter", payload.listFilter);
  if (payload.listDomain) params.set("domain", payload.listDomain);
  if (payload.listSearch) params.set("q", payload.listSearch);
  return params;
}

/** Build export URL for pivot read_group CSV snapshot. */
export function pivotExportUrl(payload: SwcWorkspacePayload, groupFields: string[], measures: string[]): string {
  const params = exportQuery(payload, "");
  if (groupFields.length) params.set("group_by", groupFields.join(","));
  if (measures.length) params.set("measures", measures.join(","));
  return `${EXPORT_PIVOT_ROUTE}?${params.toString()}`;
}

/** Build export URL for graph read_group CSV snapshot. */
export function graphExportUrl(payload: SwcWorkspacePayload, groupField: string, measureField: string): string {
  const params = exportQuery(payload, "");
  if (groupField) params.set("group_by", groupField);
  if (measureField) params.set("measure", measureField);
  return `${EXPORT_GRAPH_ROUTE}?${params.toString()}`;
}

function pivotExportFields(payload: SwcWorkspacePayload): { groups: string[]; measures: string[] } {
  const { rowFields, colFields, measureFields } = pivotFields(payload.arch);
  return { groups: [...rowFields, ...colFields], measures: measureFields };
}

export type ToolbarIconName = "search" | "filter" | "group" | "favorite" | "download" | "chevron" | "close";

const SVG_NS = "http://www.w3.org/2000/svg";

function svgEl(tag: string, attrs: Record<string, string>): SVGElement {
  const el = document.createElementNS(SVG_NS, tag);
  for (const [key, value] of Object.entries(attrs)) {
    el.setAttribute(key, value);
  }
  return el;
}

/** DOM-built icons — the html template engine does not pass through SVG attributes. */
export function createToolbarIcon(name: ToolbarIconName, className = ""): HTMLElement {
  const svg = svgEl("svg", {
    class: className,
    width: name === "search" ? "16" : "15",
    height: name === "search" ? "16" : "15",
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    "stroke-width": "2",
    "stroke-linecap": "round",
    "stroke-linejoin": "round",
    "aria-hidden": "true",
  }) as SVGSVGElement;

  switch (name) {
    case "search":
      svg.appendChild(svgEl("circle", { cx: "11", cy: "11", r: "7" }));
      svg.appendChild(svgEl("path", { d: "M21 21l-4.35-4.35" }));
      break;
    case "filter":
      svg.appendChild(svgEl("polygon", { points: "22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" }));
      break;
    case "group":
      svg.appendChild(svgEl("path", { d: "M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" }));
      svg.appendChild(svgEl("circle", { cx: "9", cy: "7", r: "4" }));
      svg.appendChild(svgEl("path", { d: "M23 21v-2a4 4 0 0 0-3-3.87" }));
      svg.appendChild(svgEl("path", { d: "M16 3.13a4 4 0 0 1 0 7.75" }));
      break;
    case "favorite":
      svg.appendChild(svgEl("path", { d: "M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" }));
      break;
    case "download":
      svg.appendChild(svgEl("path", { d: "M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" }));
      svg.appendChild(svgEl("polyline", { points: "7 10 12 15 17 10" }));
      svg.appendChild(svgEl("line", { x1: "12", y1: "15", x2: "12", y2: "3" }));
      break;
    case "chevron":
      svg.appendChild(svgEl("path", { d: "M6 9l6 6 6-6" }));
      break;
    case "close":
      svg.appendChild(svgEl("path", { d: "M18 6L6 18" }));
      svg.appendChild(svgEl("path", { d: "M6 6l12 12" }));
      break;
  }

  const wrap = document.createElement("span");
  wrap.className = "sum-toolbar-icon-wrap";
  wrap.setAttribute("aria-hidden", "true");
  wrap.appendChild(svg);
  return wrap;
}

export function renderSearchIcon(className = "sum-search-icon"): HTMLElement {
  return createToolbarIcon("search", className);
}

export function renderSearchField(
  value: string,
  onSearch: () => void,
  onInput: (next: string) => void,
): TemplateResult {
  return html`
    <div class="sum-list-search-wrap">
      <input
        type="search"
        class="sum-list-search"
        placeholder="Search…"
        value=${value}
        @keydown=${(event: Event) => (event as KeyboardEvent).key === "Enter" && onSearch()}
        @input=${(event: Event) => onInput(inputValueFromEvent(event))}
      />
      <button
        type="button"
        class="sum-list-search-submit"
        aria-label="Search"
        title="Search"
        @click=${() => onSearch()}
      >
        ${renderSearchIcon()}
      </button>
    </div>
  `;
}

export function renderNewButton(payload: SwcWorkspacePayload): HTMLElement {
  return linkButton(newRecordUrl(payload), "New", "sum-btn sum-list-btn-new");
}

export function toolbarButton(
  label: string,
  className: string,
  onClick: () => void,
  disabled = false,
): HTMLElement {
  const button = document.createElement("button");
  button.type = "button";
  button.className = className;
  button.textContent = label;
  button.disabled = disabled;
  button.addEventListener("click", onClick);
  return button;
}

/** Maps arch button class (e.g. sum_highlight) to pre-SWC header button modifiers. */
export function resolveHeaderButtonClass(archClass?: string): string {
  const base = "sum-header-btn";
  if (archClass?.includes("sum_highlight")) {
    return `${base} sum-header-btn--primary`;
  }
  return `${base} sum-header-btn--secondary`;
}

export function headerButton(
  label: string,
  archClass: string | undefined,
  onClick: () => void,
  disabled = false,
): HTMLElement {
  const className = disabled
    ? `${resolveHeaderButtonClass(archClass)} sum-header-btn--disabled`
    : resolveHeaderButtonClass(archClass);
  return toolbarButton(label, className, onClick, disabled);
}

export type ReportActionEntry = {
  label: string;
  node: HTMLElement;
};

const EXPORT_FORMAT_LINKS = [
  { fmt: "csv", route: EXPORT_CSV_ROUTE, label: "Export CSV" },
  { fmt: "pdf", route: EXPORT_PDF_ROUTE, label: "Export PDF" },
  { fmt: "xlsx", route: EXPORT_XLSX_ROUTE, label: "Export Excel" },
] as const;

export function buildReportActionEntries(
  payload: SwcWorkspacePayload,
  fields: string,
  recordId = 0,
): ReportActionEntry[] {
  const report = payload.arch.report;
  if (!report?.download && !report?.upload) return [];

  const exportParams = exportQuery(payload, fields, recordId);
  const entries: ReportActionEntry[] = [];
  const listTotal = payload.listTotal ?? 0;
  const largeList = listTotal > MAX_SYNC_EXPORT_ROWS;

  if (report.download) {
    const formats = (report.formats || "csv,pdf")
      .split(",")
      .map((f) => f.trim().toLowerCase())
      .filter(Boolean);
    const showAll = formats.length === 0;
    for (const { fmt, route, label } of EXPORT_FORMAT_LINKS) {
      if (!showAll && !formats.includes(fmt)) continue;
      if (largeList && (fmt === "csv" || fmt === "xlsx")) {
        continue;
      }
      const syncLabel =
        largeList && fmt === "pdf"
          ? `${label} (max ${MAX_SYNC_EXPORT_ROWS} rows)`
          : listTotal > 0 && !largeList && (fmt === "csv" || fmt === "xlsx")
            ? label
            : label;
      const node = protectedDownloadButton(
        `${route}?${exportParams.toString()}`,
        syncLabel,
        payload.csrfToken,
        "sum-popover-menu-link",
      );
      if (largeList && fmt === "pdf") {
        node.title = `PDF export includes at most ${MAX_SYNC_EXPORT_ROWS} rows. Use background CSV export for full data.`;
      }
      entries.push({ label: syncLabel, node });
    }
    if (largeList && fields) {
      entries.push({
        label: "Export all CSV (background)",
        node: asyncExportButton("Export all CSV (background)", payload, fields, "csv"),
      });
      entries.push({
        label: "Export all Excel (background)",
        node: asyncExportButton("Export all Excel (background)", payload, fields, "xlsx"),
      });
    }
  }
  if (report.upload && fields) {
    const templateParams = new URLSearchParams(exportParams);
    entries.push({
      label: "Download template",
      node: linkButton(`${BULK_TEMPLATE_ROUTE}?${templateParams.toString()}`, "Download template", "sum-popover-menu-link"),
    });
    const form = document.createElement("form");
    form.className = "sum-list-upload-form sum-popover-menu-form";
    form.method = "post";
    form.enctype = "multipart/form-data";
    form.action = BULK_UPLOAD_ROUTE;

    const addHidden = (name: string, value: string): void => {
      const input = document.createElement("input");
      input.type = "hidden";
      input.name = name;
      input.value = value;
      form.appendChild(input);
    };
    addHidden("csrf_token", payload.csrfToken);
    addHidden("model", payload.model);
    if (payload.actionId > 0) addHidden("action", String(payload.actionId));
    addHidden("fields", fields);

    const label = document.createElement("label");
    label.className = "sum-popover-menu-link sum-popover-menu-link--upload";
    label.textContent = "Import CSV";
    const fileInput = document.createElement("input");
    fileInput.type = "file";
    fileInput.name = "file";
    fileInput.accept = ".csv,text/csv";
    fileInput.className = "sum-list-upload-input";
    fileInput.addEventListener("change", () => form.requestSubmit());
    label.appendChild(fileInput);
    form.appendChild(label);
    entries.push({ label: "Import CSV", node: form });
  }
  return entries;
}

/** Pivot/graph CSV exports for collection views (in addition to arch.report entries). */
function appendReadGroupExportEntries(
  payload: SwcWorkspacePayload,
  viewType: string | undefined,
  entries: ReportActionEntry[],
): void {
  if (viewType === VIEW_PIVOT) {
    const { groups, measures } = pivotExportFields(payload);
    if (groups.length > 0 && measures.length > 0) {
      entries.push({
        label: "Export pivot CSV",
        node: protectedDownloadButton(
          pivotExportUrl(payload, groups, measures),
          "Export pivot CSV",
          payload.csrfToken,
          "sum-popover-menu-link",
        ),
      });
    }
    return;
  }
  if (viewType === VIEW_GRAPH) {
    const { groupField, measureField } = graphAxes(payload.arch);
    if (groupField && measureField) {
      entries.push({
        label: "Export graph CSV",
        node: protectedDownloadButton(
          graphExportUrl(payload, groupField, measureField),
          "Export graph CSV",
          payload.csrfToken,
          "sum-popover-menu-link",
        ),
      });
    }
  }
}

/** Report menu entries for toolbar (arch.report + view-specific read_group exports). */
export function buildAllReportEntries(
  payload: SwcWorkspacePayload,
  fields: string,
  recordId = 0,
  viewType?: string,
): ReportActionEntry[] {
  const entries = buildReportActionEntries(payload, fields, recordId);
  appendReadGroupExportEntries(payload, viewType, entries);
  return entries;
}
