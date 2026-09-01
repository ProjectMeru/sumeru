import { SwcComponent } from "../../runtime/component.js";
import { html, type TemplateResult, type TemplateValue } from "../../template/html.js";
import type { SwcArchField, SwcWorkspacePayload } from "../../types/workspace.js";
import { SAVED_SEARCHES_ROUTE } from "../../constants/routes.js";
import { inputValueFromEvent } from "../../widgets/field-events.js";
import { renderNewButton, buildReportActionEntries, exportFieldNamesCsv, createToolbarIcon } from "./view-toolbar.js";
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

type PanelKind = "filters" | "group" | "favorites" | "actions";

const SEGMENT_TOOLTIPS: Record<Exclude<PanelKind, "actions">, string> = {
  filters: "Filter records using presets or custom field rules",
  group: "Group records by field values",
  favorites: "Save and reopen favorite searches",
};

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
  private saveShared = false;
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
        isShared: this.saveShared,
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
    const input = this.rootElement?.querySelector(".sum-control-bar-search") as HTMLInputElement | null;
    if (input) input.value = "";
    if (this.rootElement?.isConnected) this.patch();
    this.applyQuery({ search: "" });
  }

  private renderSearchBar(tags: FilterTag[]): TemplateResult {
    const hasQuery = this.query.search.trim().length > 0;
    return html`
      <div class="sum-control-bar-search-wrap">
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
              ${createToolbarIcon("close")}
            </button>`
          : ""}
        <button
          type="button"
          class="sum-control-bar-search-submit"
          aria-label="Search"
          title="Search records"
          @click=${(e: Event) => { e.stopPropagation(); this.applySearch(); }}
        >
          ${createToolbarIcon("search", "sum-control-bar-search-icon-svg")}
        </button>
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
                <span class="sum-popover-item-label">${f.name}${f.isShared ? " (shared)" : ""}</span>
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
          <label class="sum-popover-check">
            <input
              type="checkbox"
              ?checked=${this.saveShared}
              @change=${(e: Event) => { this.saveShared = (e.target as HTMLInputElement).checked; }}
            />
            Share with all users
          </label>
          <button type="button" class="sum-btn sum-btn--secondary" disabled=${this.savingFavorite ? "disabled" : undefined} @click=${() => void this.saveFavorite()}>Save current search</button>
        </div>
      </div>
    `;
  }

  private renderSegmentButton(
    kind: Exclude<PanelKind, "actions">,
    label: string,
    iconName: "filter" | "group" | "favorite",
    badgeCount: number,
  ): TemplateResult {
    const open = this.panelOpen === kind;
    const btnClass = open
      ? "sum-control-bar-segment-btn sum-control-bar-segment-btn--active"
      : "sum-control-bar-segment-btn";
    const tooltip = SEGMENT_TOOLTIPS[kind];
    return html`
      <div class="sum-control-bar-popover-anchor">
        <button
          type="button"
          class=${btnClass}
          aria-label=${label}
          title=${tooltip}
          aria-expanded=${open ? "true" : "false"}
          @click=${(e: Event) => { e.stopPropagation(); this.togglePanel(kind); }}
        >
          ${createToolbarIcon(iconName, "sum-control-bar-segment-icon")}
          <span class="sum-control-bar-segment-label">${label}</span>
          ${this.renderBadge(badgeCount)}
        </button>
        ${open
          ? (kind === "filters"
            ? this.renderFiltersPopover()
            : kind === "group"
              ? this.renderGroupPopover()
              : this.renderFavoritesPopover())
          : ""}
      </div>
    `;
  }

  private renderActionsPopover(): TemplateResult {
    const fields = exportFieldNamesCsv((this.props.payload.arch.fields ?? []).filter((f) => !f.invisible));
    const entries = buildReportActionEntries(this.props.payload, fields);

    return html`
      <div class="sum-popover sum-popover--actions" @click=${(e: Event) => e.stopPropagation()}>
        <h3 class="sum-popover-heading">Actions</h3>
        <ul class="sum-popover-menu">
          ${entries.map((entry) => html`<li class="sum-popover-menu-item">${entry.node}</li>`)}
        </ul>
      </div>
    `;
  }

  private renderActionsTrigger(entriesCount: number): TemplateResult | string {
    if (entriesCount <= 0) return "";
    const open = this.panelOpen === "actions";
    const btnClass = open
      ? "sum-control-bar-actions-btn sum-control-bar-actions-btn--active"
      : "sum-control-bar-actions-btn";
    const icon = createToolbarIcon("download", "sum-control-bar-actions-icon");
    const chevron = createToolbarIcon("chevron", "sum-control-bar-chip-chevron");
    return html`
      <div class="sum-control-bar-popover-anchor">
        <button
          type="button"
          class=${btnClass}
          aria-label="Actions"
          title="Export, import, and other record actions"
          aria-expanded=${open ? "true" : "false"}
          @click=${(e: Event) => { e.stopPropagation(); this.togglePanel("actions"); }}
        >
          ${icon}
          <span class="sum-control-bar-actions-label">Actions</span>
          ${chevron}
        </button>
        ${open ? this.renderActionsPopover() : ""}
      </div>
    `;
  }

  override template(): TemplateResult {
    const payload = this.props.payload;
    const fields = exportFieldNamesCsv((payload.arch.fields ?? []).filter((f) => !f.invisible));
    const actionEntries = buildReportActionEntries(payload, fields);
    const tags = activeFilterTags(this.query, this.searchMeta());
    const fCount = filterCount(this.query);
    const gCount = groupByCount(this.query);

    return html`
      <div class="sum-control-bar-wrap" @click=${() => this.closePanel()}>
        <div class="sum-control-bar sum-view-toolbar">
          <div class="sum-control-bar-row">
            <div class="sum-control-bar-leading">
              ${renderNewButton(payload)}
            </div>
            <div class="sum-control-bar-search-area">
              ${this.renderSearchBar(tags)}
            </div>
            <div class="sum-control-bar-tools">
              <div class="sum-control-bar-segment" role="group" aria-label="Search options">
                ${this.renderSegmentButton("filters", "Filters", "filter", fCount)}
                ${this.renderSegmentButton("group", "Group By", "group", gCount)}
                ${this.renderSegmentButton("favorites", "Favorites", "favorite", 0)}
              </div>
              ${this.renderActionsTrigger(actionEntries.length)}
              ${this.props.extraPrimary ?? ""}
            </div>
          </div>
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
