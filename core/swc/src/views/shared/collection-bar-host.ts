import { SwcComponent } from "../../runtime/component.js";
import { html, type TemplateResult, type TemplateValue } from "../../template/html.js";
import type { SwcArchField, SwcWorkspacePayload } from "../../types/workspace.js";
import { SAVED_SEARCHES_ROUTE } from "../../constants/routes.js";
import { inputValueFromEvent } from "../../widgets/field-events.js";
import { renderNewButton, renderReportActions, exportFieldNamesCsv, renderSearchIcon } from "./view-toolbar.js";
import {
  activeFilterTags,
  appendDomainTriple,
  filterCount,
  groupByCount,
  navigateCollectionQuery,
  presetDomainFilters,
  presetGroupByFilters,
  removeFilterTag,
  syncCollectionQuery,
  toggleGroupByField,
  togglePresetFilter,
  type CollectionQuery,
  type DomainTriple,
  type FilterTag,
} from "./collection-query.js";

export interface CollectionBarHostProps {
  payload: SwcWorkspacePayload;
  viewType: string;
  extraPrimary?: TemplateValue;
}

type PanelKind = "filters" | "group" | "favorites";

const OPERATORS_BY_TYPE: Record<string, string[]> = {
  char: ["=", "!=", "ilike"],
  text: ["=", "!=", "ilike"],
  integer: ["=", "!=", ">", "<", ">=", "<="],
  float: ["=", "!=", ">", "<", ">=", "<="],
  boolean: ["=", "!="],
  date: ["=", "!=", ">", "<", ">=", "<="],
  datetime: ["=", "!=", ">", "<", ">=", "<="],
  selection: ["=", "!="],
  many2one: ["=", "!="],
};

export class CollectionBarHost extends SwcComponent<CollectionBarHostProps> {
  private query!: CollectionQuery;
  private panelOpen: PanelKind | null = null;
  private customField = "";
  private customOp = "=";
  private customValue = "";
  private saveName = "";
  private savingFavorite = false;

  override setup(): void {
    this.syncQuery();
  }

  override onPropsChanged(props: CollectionBarHostProps): void {
    this.syncQuery();
    void props;
  }

  private syncQuery(): void {
    this.query = syncCollectionQuery(this.props.payload);
  }

  private searchMeta() {
    return this.props.payload.arch.search;
  }

  private applyQuery(patch: Partial<CollectionQuery>): void {
    navigateCollectionQuery(this.env, this.props.payload, this.props.viewType, {
      search: patch.search ?? this.query.search,
      presetFilters: patch.presetFilters ?? this.query.presetFilters,
      customDomain: patch.customDomain ?? this.query.customDomain,
      groupBy: patch.groupBy ?? this.query.groupBy,
      listOffset: 0,
    });
  }

  private applySearch(): void {
    this.applyQuery({ search: this.query.search });
  }

  private togglePanel(kind: PanelKind): void {
    this.panelOpen = this.panelOpen === kind ? null : kind;
    this.rerender();
  }

  private closePanel(): void {
    if (!this.panelOpen) return;
    this.panelOpen = null;
    this.rerender();
  }

  private togglePreset(name: string): void {
    this.applyQuery({ presetFilters: togglePresetFilter(this.query.presetFilters, name) });
  }

  private toggleGroupBy(field: string): void {
    this.applyQuery({ groupBy: toggleGroupByField(this.query.groupBy, field) });
  }

  private applyCustomFilter(): void {
    const field = this.customField.trim();
    if (!field) return;
    const triple: DomainTriple = [field, this.customOp, this.coerceValue(field)];
    this.applyQuery({ customDomain: appendDomainTriple(this.query.customDomain, triple) });
    this.customValue = "";
    this.panelOpen = null;
    this.rerender();
  }

  private coerceValue(fieldName: string): unknown {
    const field = this.filterFields().find((f) => f.name === fieldName);
    const raw = this.customValue.trim();
    if (field?.type === "boolean") return raw === "true" || raw === "1";
    if (field?.type === "integer" || field?.type === "float") {
      const n = Number(raw);
      return Number.isNaN(n) ? raw : n;
    }
    return raw;
  }

  private filterFields(): SwcArchField[] {
    return this.searchMeta()?.filterFields ?? [];
  }

  private groupByFields(): SwcArchField[] {
    const meta = this.searchMeta();
    const fromModel = meta?.groupByFields ?? [];
    const extras = (meta?.filters ?? [])
      .filter((f) => f.groupBy && !fromModel.some((m) => m.name === f.groupBy))
      .map((f) => ({ name: f.groupBy!, string: f.string, type: "char" }));
    return [...fromModel, ...extras];
  }

  private operatorsForField(fieldName: string): string[] {
    const field = this.filterFields().find((f) => f.name === fieldName);
    return OPERATORS_BY_TYPE[field?.type ?? "char"] ?? ["=", "!="];
  }

  private removeTag(tag: FilterTag): void {
    const next = removeFilterTag(this.query, tag);
    this.applyQuery(next);
  }

  private applyFavorite(fav: { search?: string; filter?: string; domain?: string; groupBy?: string }): void {
    this.applyQuery({
      search: fav.search ?? "",
      presetFilters: (fav.filter ?? "").split(",").map((s) => s.trim()).filter(Boolean),
      customDomain: fav.domain ?? "",
      groupBy: (fav.groupBy ?? "").split(",").map((s) => s.trim()).filter(Boolean),
    });
    this.panelOpen = null;
    this.rerender();
  }

  private async saveFavorite(): Promise<void> {
    const name = this.saveName.trim();
    if (!name || this.savingFavorite) return;
    this.savingFavorite = true;
    this.rerender();
    try {
      const payload = this.props.payload;
      await this.env.services.http.postJSON(SAVED_SEARCHES_ROUTE, {
        actionId: payload.actionId,
        model: payload.model,
        name,
        search: this.query.search,
        filter: this.query.presetFilters.join(","),
        domain: this.query.customDomain,
        groupBy: this.query.groupBy.join(","),
      });
      this.saveName = "";
      this.env.services.notification.success("Saved", "Search saved to favorites.");
      this.panelOpen = null;
    } catch (err) {
      this.env.services.notification.error("Save failed", String(err));
    } finally {
      this.savingFavorite = false;
      this.rerender();
    }
  }

  private async deleteFavorite(id: number): Promise<void> {
    try {
      await this.env.services.http.delete(`${SAVED_SEARCHES_ROUTE}?id=${id}`);
      this.env.services.notification.success("Removed", "Favorite deleted.");
      this.rerender();
    } catch (err) {
      this.env.services.notification.error("Delete failed", String(err));
    }
  }

  private renderBadge(count: number): TemplateResult | string {
    if (count <= 0) return "";
    return html`<span class="sum-control-bar-badge">${count}</span>`;
  }

  private renderCheckmark(active: boolean): TemplateResult {
    return html`<span class="sum-popover-check" aria-hidden="true">${active ? "✓" : ""}</span>`;
  }

  private clearSearch(): void {
    this.query.search = "";
    this.applyQuery({ search: "" });
  }

  private renderSearchBar(tags: FilterTag[]): TemplateResult {
    const hasQuery = this.query.search.trim().length > 0;
    return html`
      <div class="sum-control-bar-search-wrap">
        <span class="sum-control-bar-search-icon" aria-hidden="true">
          ${renderSearchIcon("sum-control-bar-search-icon-svg")}
        </span>
        <div class="sum-control-bar-search-body">
          ${tags.length > 0
            ? html`<div class="sum-control-bar-search-tags">${tags.map((tag) => this.renderTag(tag))}</div>`
            : ""}
          <input
            type="search"
            class="sum-control-bar-search"
            placeholder=${tags.length > 0 ? "Search within results…" : "Search records…"}
            aria-label="Search records"
            value=${this.query.search}
            @keydown=${(event: Event) => (event as KeyboardEvent).key === "Enter" && this.applySearch()}
            @input=${(event: Event) => { this.query.search = inputValueFromEvent(event); this.rerender(); }}
            @click=${(e: Event) => e.stopPropagation()}
          />
        </div>
        ${hasQuery
          ? html`<button
              type="button"
              class="sum-control-bar-search-clear"
              aria-label="Clear search"
              title="Clear search"
              @click=${(e: Event) => { e.stopPropagation(); this.clearSearch(); }}
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M18 6L6 18" />
                <path d="M6 6l12 12" />
              </svg>
            </button>`
          : ""}
      </div>
    `;
  }

  private renderTag(tag: FilterTag): TemplateResult {
    return html`<span class="sum-filter-tag sum-filter-tag--${tag.kind}">
      ${tag.label}
      <button type="button" class="sum-filter-tag-remove" aria-label="Remove" @click=${() => this.removeTag(tag)}>×</button>
    </span>`;
  }

  private renderPopoverItem(label: string, active: boolean, onClick: () => void): TemplateResult {
    return html`<li>
      <button
        type="button"
        class=${active ? "sum-popover-item sum-popover-item--active" : "sum-popover-item"}
        @click=${onClick}
      >
        ${this.renderCheckmark(active)}
        <span class="sum-popover-item-label">${label}</span>
      </button>
    </li>`;
  }

  private renderFiltersPopover(): TemplateResult {
    const meta = this.searchMeta();
    const domainPresets = presetDomainFilters(meta?.filters);
    const fields = this.filterFields();
    if (!this.customField && fields[0]) this.customField = fields[0].name;

    return html`
      <div class="sum-popover sum-popover--filters" @click=${(e: Event) => e.stopPropagation()}>
        <h3 class="sum-popover-heading">Filters</h3>
        <ul class="sum-popover-list">
          ${domainPresets.map((f) =>
            this.renderPopoverItem(
              f.string || f.name,
              this.query.presetFilters.includes(f.name),
              () => this.togglePreset(f.name),
            ),
          )}
        </ul>
        <div class="sum-popover-custom">
          <strong class="sum-popover-custom-title">Custom filter</strong>
          <select class="sum-popover-select" @change=${(e: Event) => { this.customField = (e.target as HTMLSelectElement).value; this.customOp = this.operatorsForField(this.customField)[0] ?? "="; this.rerender(); }}>
            ${fields.map((f) => html`<option value=${f.name} selected=${f.name === this.customField ? "selected" : undefined}>${f.string || f.name}</option>`)}
          </select>
          <select class="sum-popover-select" @change=${(e: Event) => { this.customOp = (e.target as HTMLSelectElement).value; }}>
            ${this.operatorsForField(this.customField).map((op) => html`<option value=${op} selected=${op === this.customOp ? "selected" : undefined}>${op}</option>`)}
          </select>
          <input
            type="text"
            class="sum-popover-input"
            placeholder="Value"
            value=${this.customValue}
            @input=${(e: Event) => { this.customValue = inputValueFromEvent(e); }}
          />
          <button type="button" class="sum-btn sum-btn--secondary" @click=${() => this.applyCustomFilter()}>Apply</button>
        </div>
      </div>
    `;
  }

  private renderGroupPopover(): TemplateResult {
    const meta = this.searchMeta();
    const groupPresets = presetGroupByFilters(meta?.filters);

    return html`
      <div class="sum-popover sum-popover--group" @click=${(e: Event) => e.stopPropagation()}>
        <h3 class="sum-popover-heading">Group By</h3>
        <ul class="sum-popover-list">
          ${groupPresets.map((f) =>
            this.renderPopoverItem(
              f.string || f.name,
              this.query.groupBy.includes(f.groupBy!),
              () => this.toggleGroupBy(f.groupBy!),
            ),
          )}
          ${this.groupByFields().map((f) =>
            this.renderPopoverItem(
              f.string || f.name,
              this.query.groupBy.includes(f.name),
              () => this.toggleGroupBy(f.name),
            ),
          )}
        </ul>
      </div>
    `;
  }

  private renderFavoritesPopover(): TemplateResult {
    const favorites = this.props.payload.favorites ?? [];

    return html`
      <div class="sum-popover sum-popover--favorites" @click=${(e: Event) => e.stopPropagation()}>
        <h3 class="sum-popover-heading">Favorites</h3>
        <ul class="sum-popover-list">
          ${favorites.map(
            (f) => html`<li class="sum-popover-fav">
              <button type="button" class="sum-popover-item" @click=${() => this.applyFavorite(f)}>
                <span class="sum-popover-item-label">${f.name}</span>
              </button>
              <button type="button" class="sum-popover-fav-delete" @click=${() => void this.deleteFavorite(f.id)} title="Delete">×</button>
            </li>`,
          )}
        </ul>
        <div class="sum-popover-custom">
          <input
            type="text"
            class="sum-popover-input"
            placeholder="Favorite name"
            value=${this.saveName}
            @input=${(e: Event) => { this.saveName = inputValueFromEvent(e); }}
          />
          <button type="button" class="sum-btn sum-btn--secondary" disabled=${this.savingFavorite ? "disabled" : undefined} @click=${() => void this.saveFavorite()}>Save current search</button>
        </div>
      </div>
    `;
  }

  private renderActionButton(
    kind: PanelKind,
    label: string,
    icon: TemplateResult,
    badgeCount: number,
  ): TemplateResult {
    const open = this.panelOpen === kind;
    const btnClass = open
      ? "sum-control-bar-chip-btn sum-control-bar-chip-btn--active"
      : "sum-control-bar-chip-btn";
    const chevron = html`<svg class="sum-control-bar-chip-chevron" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M6 9l6 6 6-6" />
    </svg>`;
    return html`
      <div class="sum-control-bar-popover-anchor">
        <button
          type="button"
          class=${btnClass}
          aria-label=${label}
          aria-expanded=${open ? "true" : "false"}
          @click=${(e: Event) => { e.stopPropagation(); this.togglePanel(kind); }}
        >
          ${icon}
          <span class="sum-control-bar-chip-label">${label}</span>
          ${this.renderBadge(badgeCount)}
          ${chevron}
        </button>
        ${open ? (kind === "filters" ? this.renderFiltersPopover() : kind === "group" ? this.renderGroupPopover() : this.renderFavoritesPopover()) : ""}
      </div>
    `;
  }

  override template(): TemplateResult {
    const payload = this.props.payload;
    const fields = exportFieldNamesCsv((payload.arch.fields ?? []).filter((f) => !f.invisible));
    const reportActions = renderReportActions(payload, fields);
    const tags = activeFilterTags(this.query, this.searchMeta());
    const fCount = filterCount(this.query);
    const gCount = groupByCount(this.query);

    const filterIcon = html`<svg class="sum-control-bar-chip-icon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
    </svg>`;

    const groupIcon = html`<svg class="sum-control-bar-chip-icon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
      <path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>`;

    const favIcon = html`<svg class="sum-control-bar-chip-icon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
    </svg>`;

    return html`
      <div class="sum-control-bar-wrap" @click=${() => this.closePanel()}>
        <div class="sum-control-bar sum-view-toolbar">
          <div class="sum-view-toolbar-primary">
            ${renderNewButton(payload)}
            <div class="sum-control-bar-search-group">
              ${this.renderSearchBar(tags)}
              ${this.renderActionButton("filters", "Filters", filterIcon, fCount)}
              ${this.renderActionButton("group", "Group By", groupIcon, gCount)}
              ${this.renderActionButton("favorites", "Favorites", favIcon, 0)}
            </div>
            ${this.props.extraPrimary ?? ""}
          </div>
          ${reportActions ?? ""}
        </div>
      </div>
    `;
  }
}

export function mountCollectionBar(
  payload: SwcWorkspacePayload,
  viewType: string,
  env: CollectionBarHost["env"],
  extraPrimary?: TemplateValue,
): CollectionBarHost {
  const host = new CollectionBarHost({ payload, viewType, extraPrimary }, env);
  host.callSetup();
  return host;
}
