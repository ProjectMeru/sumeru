import { html, type TemplateResult } from "../../template/html.js";
import type { SwcEnv } from "../../runtime/env.js";
import type {
  SwcArchField,
  SwcArchGroup,
  SwcArchNotebook,
  SwcArchSheet,
  SwcArchDiv,
  SwcArchSeparator,
  SwcArchLabel,
} from "../../types/workspace.js";
import type { SwcRecord } from "../../model/record.js";
import { renderField as defaultRenderField } from "../../widgets/registry.js";
import { fieldInputId, fieldPlaceholder, fieldAutocomplete } from "../../widgets/field-shell.js";

export type RenderFieldFn = (
  field: SwcArchField,
  record: SwcRecord,
  readonly: boolean,
) => HTMLElement;

function renderFields(
  rf: RenderFieldFn,
  fields: SwcArchField[],
  record: SwcRecord,
  readonly: boolean,
): Array<TemplateResult | HTMLElement> {
  return visibleFields(fields).map((f) => rf(f, record, readonly));
}

function collectDivFields(div: SwcArchDiv): SwcArchField[] {
  const out = [...(div.fields ?? []), ...(div.h1Fields ?? [])];
  for (const nested of div.divs ?? []) {
    out.push(...collectDivFields(nested));
  }
  return out;
}

export function collectFormFields(sheet?: SwcArchSheet, headerFields: SwcArchField[] = []): SwcArchField[] {
  const out = [...headerFields];
  if (!sheet) return out.filter((f) => !f.invisible);

  out.push(...(sheet.fields ?? []));
  for (const div of sheet.divs ?? []) {
    out.push(...collectDivFields(div));
  }
  for (const g of sheet.groups ?? []) {
    out.push(...collectGroupFields(g));
  }
  for (const nb of sheet.notebook ?? []) {
    for (const pg of nb.pages ?? []) {
      out.push(...(pg.fields ?? []));
      for (const g of pg.groups ?? []) {
        out.push(...collectGroupFields(g));
      }
    }
  }
  return out.filter((f) => !f.invisible);
}

function collectGroupFields(group: SwcArchGroup): SwcArchField[] {
  const out = [...(group.fields ?? [])];
  for (const nested of group.groups ?? []) {
    out.push(...collectGroupFields(nested));
  }
  return out;
}

function visibleFields(fields: SwcArchField[]): SwcArchField[] {
  return fields.filter((f) => !f.invisible);
}

function renderSeparators(separators: SwcArchSeparator[] = []): TemplateResult {
  if (separators.length === 0) return html``;
  return html`${separators.map((sep) =>
    sep.string
      ? html`<div class="sum-separator--title">${sep.string}</div>`
      : html`<hr class="sum-separator--rule" />`,
  )}`;
}

function renderLabels(labels: SwcArchLabel[] = []): TemplateResult {
  if (labels.length === 0) return html``;
  return html`${labels.map((lab) => {
    const text = lab.string ?? "";
    if (lab.for) {
      return html`<label class="sum-form-label sum-form-label--hint" for=${`f-${lab.for}`}>${text}</label>`;
    }
    return html`<div class="sum-form-label sum-form-label--hint">${text}</div>`;
  })}`;
}

function initialsFromName(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

function renderHeroField(
  field: SwcArchField,
  record: SwcRecord,
  readonly: boolean,
): TemplateResult {
  const val = String(record.get(field.name) ?? "");
  const placeholder = fieldPlaceholder(field);
  const hasValue = val.trim() !== "";
  if (readonly || field.readonly) {
    const text = hasValue ? val : placeholder;
    const cls = hasValue
      ? "sum-form-hero-input sum-form-hero-input--bold"
      : "sum-form-hero-input sum-form-hero-input--bold sum-form-hero-input--placeholder";
    return html`<h1><div class=${cls}>${text}</div></h1>`;
  }
  return html`<h1>
    <input
      id=${fieldInputId(field)}
      class="sum-form-hero-input sum-form-hero-input--bold"
      name=${field.name}
      placeholder=${placeholder}
      value=${val}
      autocomplete=${fieldAutocomplete(field)}
      aria-label=${placeholder}
      @input=${(ev: Event) => record.set(field.name, (ev.target as HTMLInputElement).value)}
    />
  </h1>`;
}

function renderContactItem(
  field: SwcArchField,
  record: SwcRecord,
  readonly: boolean,
): TemplateResult {
  const val = String(record.get(field.name) ?? "");
  const label = field.string ?? field.name;
  const placeholder = fieldPlaceholder(field);
  const inputType = field.widget === "email" ? "email" : "text";
  if (readonly || field.readonly) {
    const text = val.trim() !== "" ? val : placeholder;
    const cls = val.trim() !== "" ? "sum-form-inline-input" : "sum-form-inline-input sum-form-inline-input--placeholder";
    return html`<div class="sum-form-contact-item">
      <label class="sum-field-label">${label}</label>
      <div class=${cls}>${text}</div>
    </div>`;
  }
  return html`<div class="sum-form-contact-item">
    <label class="sum-field-label">${label}</label>
    <input
      type=${inputType}
      class="sum-form-inline-input"
      name=${field.name}
      placeholder=${placeholder}
      value=${val}
      @input=${(ev: Event) => record.set(field.name, (ev.target as HTMLInputElement).value)}
    />
  </div>`;
}

function renderAvatar(record: SwcRecord, readonly: boolean): TemplateResult {
  const image = String(record.get("image") ?? "");
  const name = String(record.get("name") ?? "");
  const hasImage = image.length > 0;
  const initials = initialsFromName(name);

  return html`<div class="sum-form-avatar sum-form-avatar--compact" data-sum-avatar>
    <div class="sum-form-avatar-box sum-form-avatar-box--circle">
      ${hasImage
        ? html`<img
            .sum-form-avatar-img
            .sum-form-avatar-img--visible
            class=${image.includes("data:") ? "sum-form-avatar-img--cropped" : ""}
            src=${image}
            alt=""
          />`
        : html`<span class="sum-form-avatar-initials">${initials}</span>`}
    </div>
    ${readonly
      ? ""
      : html`<div class="sum-form-avatar-actions">
          <input
            type="hidden"
            name="image"
            data-sum-avatar-value
            value=${image}
            @input=${(ev: Event) => record.set("image", (ev.target as HTMLInputElement).value)}
          />
          <label class="sum-form-avatar-upload">
            Upload
            <input type="file" accept="image/*" />
          </label>
        </div>`}
  </div>`;
}

function renderTitleBody(
  rf: RenderFieldFn,
  div: SwcArchDiv,
  record: SwcRecord,
  readonly: boolean,
): TemplateResult {
  const h1Fields = visibleFields(div.h1Fields ?? []);
  const contactDiv = (div.divs ?? []).find((d) => (d.class ?? "").includes("sum-title-contact-row"));
  const contactFields = visibleFields(contactDiv?.fields ?? []);

  return html`<div class="sum-form-title-body sum-form-title-body--main">
    ${h1Fields.length > 0 ? renderHeroField(h1Fields[0], record, readonly) : ""}
    ${contactFields.length > 0
      ? html`<div class="sum-title-contact-row">
          ${contactFields.map((f) => renderContactItem(f, record, readonly))}
        </div>`
      : ""}
    ${h1Fields.length === 0 && contactFields.length === 0
      ? renderFields(rf, div.fields ?? [], record, readonly)
      : ""}
  </div>`;
}

function renderTitleDiv(
  rf: RenderFieldFn,
  div: SwcArchDiv,
  record: SwcRecord,
  readonly: boolean,
  hasImageField: boolean,
  onStatButton?: (name: string) => void,
): TemplateResult {
  const cls = div.class ?? "";
  if (cls.includes("sum_button_box") || cls.includes("button_box")) {
    const buttons = div.buttons ?? [];
    return html`<div class="sum-form-button-box ${cls}">
      ${buttons.map(
        (btn) => html`<button type="button" class="sum-stat-button ${btn.class ?? ""}" data-action=${btn.name} @click=${() => onStatButton?.(btn.name)}>
          ${btn.string || btn.name}
        </button>`,
      )}
    </div>`;
  }

  const isTitle = cls.includes("sum_title");
  if (!isTitle) {
    return html`<div class=${cls}>${renderFields(rf, div.fields ?? [], record, readonly)}</div>`;
  }

  const h1Fields = visibleFields(div.h1Fields ?? []);
  const legacySingle = h1Fields.length === 0 && visibleFields(div.fields ?? []).length === 1;
  const titleField = h1Fields[0] ?? (legacySingle ? visibleFields(div.fields ?? [])[0] : undefined);

  if (hasImageField) {
    return html`<div class="sum-form-split-layout sum-form-split-layout--compact" data-sum-form-split>
      <aside class="sum-form-split-left sum-form-split-left--avatar">${renderAvatar(record, readonly)}</aside>
      <div class="sum-form-split-main">${renderTitleBody(rf, div, record, readonly)}</div>
    </div>`;
  }

  if (titleField) {
    return html`<div class="sum-form-title-row sum-form-title-row--sheet">
      ${renderTitleBody(rf, div, record, readonly)}
    </div>`;
  }

  return html`<div class="sum-form-title-row sum-form-title-row--sheet">
    ${renderTitleBody(rf, div, record, readonly)}
  </div>`;
}

type GroupLayoutContext = "sheet" | "row" | "inner";

/** Max logical columns for an outer group row (group `col` attribute, else nested child count). */
function outerGroupMaxCols(group: SwcArchGroup, childCount: number): number {
  if (group.col && group.col > 0) return group.col;
  return Math.max(childCount, 1);
}

function childGroupColspan(group: SwcArchGroup): number {
  return group.colspan && group.colspan > 0 ? group.colspan : 1;
}

/** Map colspan within a maxCols row onto a 12-column grid span. */
function gridSpan12(maxCols: number, colspan: number): number {
  const cols = Math.max(maxCols, 1);
  const span = Math.max(colspan, 1);
  return Math.min(12, Math.max(1, Math.round((span * 12) / cols)));
}

interface GroupRowItem {
  group: SwcArchGroup;
  gridSpan: number;
}

function packGroupRows(parent: SwcArchGroup, nested: SwcArchGroup[]): GroupRowItem[][] {
  const maxCols = outerGroupMaxCols(parent, nested.length);
  const rows: GroupRowItem[][] = [];
  let current: GroupRowItem[] = [];
  let used = 0;

  for (const child of nested) {
    const colspan = childGroupColspan(child);
    if (used + colspan > maxCols && current.length > 0) {
      rows.push(current);
      current = [];
      used = 0;
    }
    current.push({ group: child, gridSpan: gridSpan12(maxCols, colspan) });
    used += colspan;
  }
  if (current.length > 0) rows.push(current);
  return rows;
}

function groupClassNames(group: SwcArchGroup, ctx: GroupLayoutContext, plain: boolean): string {
  const parts = ["sum-form-group"];
  if (plain || !group.string) {
    parts.push("sum-form-group--plain");
  } else if (ctx === "row" || ctx === "inner") {
    parts.push("sum-form-group--col");
  } else {
    parts.push("sum-form-group--full");
  }
  if ((group.fields ?? []).length > 0) {
    parts.push("sum-form-group--row-layout");
  }
  return parts.join(" ");
}

function renderGroup(
  rf: RenderFieldFn,
  group: SwcArchGroup,
  record: SwcRecord,
  readonly: boolean,
  ctx: GroupLayoutContext = "sheet",
  plain = false,
): TemplateResult {
  const nested = group.groups ?? [];
  const fields = group.fields ?? [];
  const hasNested = nested.length > 0;

  if (hasNested && fields.length === 0) {
    const rows = packGroupRows(group, nested);
    return html`<div class="sum-form-group-outer sum-field-region--sheet">
      ${rows.map(
        (row) => html`<div class="sum-form-group-row">
          ${row.map(
            (item) => html`<div class="sum-form-group-span" style=${`--sum-group-span:${item.gridSpan}`}>
              ${renderGroup(rf, item.group, record, readonly, "row")}
            </div>`,
          )}
        </div>`,
      )}
    </div>`;
  }

  const innerCols = group.col && group.col > 0 ? group.col : 0;
  const innerColsClass = innerCols > 0 ? " sum-form-group--inner-cols" : "";

  return html`<div
    class=${groupClassNames(group, ctx, plain) + innerColsClass}
    style=${innerCols ? `--sum-inner-cols:${innerCols}` : false}
  >
    ${group.string ? html`<div class="sum-form-group-title">${group.string}</div>` : ""}
    <div class="sum-form-group-grid">
      ${renderFields(rf, fields, record, readonly)}
      ${renderSeparators(group.separators)}
      ${renderLabels(group.labels)}
      ${nested.map((g) => renderGroup(rf, g, record, readonly, "inner", true))}
    </div>
  </div>`;
}

function renderNotebook(
  rf: RenderFieldFn,
  notebook: SwcArchNotebook,
  record: SwcRecord,
  readonly: boolean,
  notebookIndex: number,
  activePage: number,
  onTab: (notebookIndex: number, pageIndex: number) => void,
): TemplateResult {
  const pages = notebook.pages ?? [];
  if (pages.length === 0) return html``;

  const idx = Math.min(Math.max(activePage, 0), pages.length - 1);
  const page = pages[idx];

  return html`<div class="sum-notebook sum-notebook--sheet">
    <div class="sum-notebook-tabs" role="tablist">
      ${pages.map((pg, i) => {
        const tabClass = i === idx ? "sum-notebook-tab sum-notebook-tab--active" : "sum-notebook-tab";
        return html`<button type="button" class=${tabClass} role="tab" aria-selected=${i === idx ? "true" : "false"} @click=${() => onTab(notebookIndex, i)}>${pg.title}</button>`;
      })}
    </div>
    <div class="sum-notebook-page sum-notebook-page--sheet" role="tabpanel">
      <div class="sum-form-sheet-stack sum-notebook-page-body">
        ${renderFields(rf, page.fields ?? [], record, readonly)}
        ${(page.groups ?? []).map((g) => renderGroup(rf, g, record, readonly))}
        ${renderSeparators(page.separators)}
        ${renderLabels(page.labels)}
      </div>
    </div>
  </div>`;
}

export interface RenderFormSheetOptions {
  env: SwcEnv;
  sheet?: SwcArchSheet;
  record: SwcRecord;
  readonly: boolean;
  hasImageField?: boolean;
  activeNotebookPages: Record<number, number>;
  onNotebookTab: (notebookIndex: number, pageIndex: number) => void;
  renderField?: RenderFieldFn;
  onStatButton?: (name: string) => void;
}

export function renderFormSheet(opts: RenderFormSheetOptions): TemplateResult {
  const {
    env,
    sheet,
    record,
    readonly,
    hasImageField = false,
    activeNotebookPages,
    onNotebookTab,
    renderField: renderFieldOpt,
    onStatButton,
  } = opts;
  const rf: RenderFieldFn =
    renderFieldOpt ?? ((f, r, ro) => defaultRenderField(env, f, r, ro));
  if (!sheet) {
    return html`<div class="sum-form-sheet"></div>`;
  }

  const parts: Array<TemplateResult | HTMLElement> = [];

  for (const div of sheet.divs ?? []) {
    parts.push(renderTitleDiv(rf, div, record, readonly, hasImageField, onStatButton));
  }

  const topFields = visibleFields(sheet.fields ?? []);
  const groups = sheet.groups ?? [];
  if (topFields.length > 0 || groups.length > 0) {
    parts.push(
      html`<div class="sum-form-sheet-stack sum-field-region--sheet">
        ${renderFields(rf, topFields, record, readonly)}
        ${groups.map((g) => renderGroup(rf, g, record, readonly))}
      </div>`,
    );
  }

  (sheet.notebook ?? []).forEach((nb, notebookIndex) => {
    const activePage = activeNotebookPages[notebookIndex] ?? 0;
    parts.push(renderNotebook(rf, nb, record, readonly, notebookIndex, activePage, onNotebookTab));
  });

  const sheetSeparators = sheet.separators ?? [];
  const sheetLabels = sheet.labels ?? [];
  if (sheetSeparators.length > 0 || sheetLabels.length > 0) {
    parts.push(
      html`<div class="sum-form-sheet-meta">${renderSeparators(sheetSeparators)}${renderLabels(sheetLabels)}</div>`,
    );
  }

  return html`<div class="sum-form-sheet">${parts}</div>`;
}
