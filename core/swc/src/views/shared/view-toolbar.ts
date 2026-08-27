import { html, type TemplateResult, type TemplateValue } from "../../template/html.js";
import type { SwcArchField, SwcWorkspacePayload } from "../../types/workspace.js";
import {
  BULK_TEMPLATE_ROUTE,
  BULK_UPLOAD_ROUTE,
  EXPORT_CSV_ROUTE,
  EXPORT_PDF_ROUTE,
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

export function visibleFieldNames(fields: SwcArchField[]): string {
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
  return params;
}

export function renderSearchField(
  value: string,
  onSearch: () => void,
  onInput: (next: string) => void,
): TemplateResult {
  return html`
    <div class="sum-list-search-wrap">
      <span class="sum-list-search-icon" aria-hidden="true">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="7" />
          <path d="M20 20l-3-3" />
        </svg>
      </span>
      <input
        type="search"
        class="sum-list-search"
        placeholder="Search…"
        value=${value}
        @keydown=${(ev: Event) => (ev as KeyboardEvent).key === "Enter" && onSearch()}
        @input=${(ev: Event) => onInput((ev.target as HTMLInputElement).value)}
      />
    </div>
  `;
}

export function renderNewButton(payload: SwcWorkspacePayload): HTMLElement {
  return linkButton(newRecordUrl(payload), "New", "sum-btn sum-list-btn-new");
}

export function renderCollectionToolbar(opts: {
  payload: SwcWorkspacePayload;
  viewType: string;
  search: string;
  onSearch: () => void;
  onInput: (next: string) => void;
  extraPrimary?: TemplateValue;
}): TemplateResult {
  const fields = visibleFieldNames((opts.payload.arch.fields ?? []).filter((f) => !f.invisible));
  const reportActions = renderReportActions(opts.payload, fields);
  const toolbarClass = opts.viewType === VIEW_KANBAN ? "sum-kanban-report-bar" : "sum-list-toolbar";
  return html`
    <div class="sum-view-toolbar ${toolbarClass}">
      <div class="sum-view-toolbar-primary">
        ${renderNewButton(opts.payload)}
        ${renderSearchField(opts.search, opts.onSearch, opts.onInput)}
        ${opts.extraPrimary ?? ""}
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
  const btn = document.createElement("button");
  btn.type = "button";
  btn.className = className;
  btn.textContent = label;
  btn.disabled = disabled;
  btn.addEventListener("click", onClick);
  return btn;
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
    items.push(linkButton(`${EXPORT_CSV_ROUTE}?${exportParams.toString()}`, "Export CSV"));
    items.push(linkButton(`${EXPORT_PDF_ROUTE}?${exportParams.toString()}`, "Export PDF"));
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
          <input type="file" name="file" accept=".csv,text/csv" class="sum-list-upload-input" @change=${(ev: Event) => (ev.target as HTMLInputElement).form?.requestSubmit()} />
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
