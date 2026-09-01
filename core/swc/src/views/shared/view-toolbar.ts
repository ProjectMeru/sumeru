import { html, type TemplateResult, type TemplateValue } from "../../template/html.js";
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
  VIEW_FORM,
  VIEW_KANBAN,
} from "../../constants/routes.js";
import { RouterService } from "../../services/router.js";

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
  const groups: string[] = [];
  const measures: string[] = [];
  for (const f of payload.arch.fields ?? []) {
    const kind = (f.pivotType ?? "").toLowerCase();
    if (kind === "row" || kind === "col" || kind === "column") groups.push(f.name);
    if (kind === "measure") measures.push(f.name);
  }
  return { groups, measures };
}

export function renderPivotExportLink(payload: SwcWorkspacePayload): HTMLElement | null {
  const { groups, measures } = pivotExportFields(payload);
  if (!groups.length || !measures.length) return null;
  return linkButton(pivotExportUrl(payload, groups, measures), "Export CSV");
}

export function renderGraphExportLink(payload: SwcWorkspacePayload, groupField: string, measureField: string): HTMLElement | null {
  if (!groupField || !measureField) return null;
  return linkButton(graphExportUrl(payload, groupField, measureField), "Export CSV");
}

export function renderSearchIcon(className = "sum-search-icon"): TemplateResult {
  return html`<svg class=${className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <circle cx="11" cy="11" r="7" />
    <path d="M21 21l-4.35-4.35" />
  </svg>`;
}

export function renderSearchField(
  value: string,
  onSearch: () => void,
  onInput: (next: string) => void,
): TemplateResult {
  return html`
    <div class="sum-list-search-wrap">
      <span class="sum-list-search-icon" aria-hidden="true">
        ${renderSearchIcon()}
      </span>
      <input
        type="search"
        class="sum-list-search"
        placeholder="Search…"
        value=${value}
        @keydown=${(event: Event) => (event as KeyboardEvent).key === "Enter" && onSearch()}
        @input=${(event: Event) => onInput(inputValueFromEvent(event))}
      />
    </div>
  `;
}

export function renderNewButton(payload: SwcWorkspacePayload): HTMLElement {
  return linkButton(newRecordUrl(payload), "New", "sum-btn sum-list-btn-new");
}

export function renderCollectionToolbar(options: {
  payload: SwcWorkspacePayload;
  viewType: string;
  search: string;
  onSearch: () => void;
  onInput: (next: string) => void;
  extraPrimary?: TemplateValue;
}): TemplateResult {
  const fields = exportFieldNamesCsv((options.payload.arch.fields ?? []).filter((f) => !f.invisible));
  const reportActions = renderReportActions(options.payload, fields);
  const toolbarClass = options.viewType === VIEW_KANBAN ? "sum-kanban-report-bar" : "sum-list-toolbar";
  return html`
    <div class="sum-view-toolbar ${toolbarClass}">
      <div class="sum-view-toolbar-primary">
        ${renderNewButton(options.payload)}
        ${renderSearchField(options.search, options.onSearch, options.onInput)}
        ${options.extraPrimary ?? ""}
      </div>
      ${reportActions ?? ""}
    </div>
  `;
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

export function renderReportActions(
  payload: SwcWorkspacePayload,
  fields: string,
  recordId = 0,
): HTMLElement | null {
  const report = payload.arch.report;
  if (!report?.download && !report?.upload) return null;

  const exportParams = exportQuery(payload, fields, recordId);
  const items: Array<HTMLElement | TemplateResult> = [];

  if (report.download) {
    const formats = (report.formats || "csv,pdf")
      .split(",")
      .map((f) => f.trim().toLowerCase())
      .filter(Boolean);
    const showAll = formats.length === 0;
    if (showAll || formats.includes("csv")) {
      items.push(linkButton(`${EXPORT_CSV_ROUTE}?${exportParams.toString()}`, "Export CSV"));
    }
    if (showAll || formats.includes("pdf")) {
      items.push(linkButton(`${EXPORT_PDF_ROUTE}?${exportParams.toString()}`, "Export PDF"));
    }
    if (formats.includes("xlsx")) {
      items.push(linkButton(`${EXPORT_XLSX_ROUTE}?${exportParams.toString()}`, "Export Excel"));
    }
  }
  if (report.upload && fields) {
    const templateParams = new URLSearchParams(exportParams);
    items.push(linkButton(`${BULK_TEMPLATE_ROUTE}?${templateParams.toString()}`, "Download template"));
    items.push(
      html`<form class="sum-list-upload-form" method="post" enctype="multipart/form-data" action=${BULK_UPLOAD_ROUTE}>
        <input type="hidden" name="csrf_token" value=${payload.csrfToken} />
        <input type="hidden" name="model" value=${payload.model} />
        ${payload.actionId > 0 ? html`<input type="hidden" name="action" value=${String(payload.actionId)} />` : ""}
        <input type="hidden" name="fields" value=${fields} />
        <label class="sum-btn sum-btn--secondary sum-list-upload-label">
          Import CSV
          <input type="file" name="file" accept=".csv,text/csv" class="sum-list-upload-input" @change=${(event: Event) => (event.target as HTMLInputElement).form?.requestSubmit()} />
        </label>
      </form>`,
    );
  }
  if (items.length === 0) return null;

  const wrap = document.createElement("div");
  wrap.className = "sum-view-toolbar-actions";
  for (const item of items) {
    wrap.appendChild(item instanceof HTMLElement ? item : item.render());
  }
  return wrap;
}
