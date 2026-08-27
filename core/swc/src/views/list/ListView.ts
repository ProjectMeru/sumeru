import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcArchButton, SwcWorkspacePayload } from "../../types/workspace.js";
import { useState } from "../../runtime/hooks.js";
import { headerButton, renderCollectionToolbar } from "../shared/view-toolbar.js";
import { SwcError } from "../../runtime/error.js";
import {
  parseFilterCSV,
  renderControlPanel,
  renderRowCheckbox,
  renderSearchFilters,
  renderSelectAllHeader,
  renderSortHeader,
  toggleFilterName,
  type ControlPanelState,
} from "./control-panel.js";
import { forEach } from "../../template/helpers.js";
import { patchKeyedChildren } from "../../runtime/patch/keyed.js";
import { VIEW_FORM, VIEW_LIST } from "../../constants/routes.js";

interface ListViewProps {
  payload: SwcWorkspacePayload;
}

export class ListView extends SwcComponent<ListViewProps> {
  private panelState: ControlPanelState = {
    search: "",
    offset: 0,
    limit: 40,
    selectedIds: new Set(),
    filters: [],
  };
  private deleting = false;
  private acting = false;

  setup(): void {
    this.syncFromPayload(this.props.payload);
    const [, bump] = useState(0);
    this.bump = () => bump((n) => n + 1);
  }

  onPropsChanged(props: ListViewProps): void {
    this.syncFromPayload(props.payload);
    this.panelState.selectedIds = new Set();
  }

  private bump: (() => void) | null = null;

  private syncFromPayload(p: SwcWorkspacePayload): void {
    this.panelState.search = p.listSearch ?? "";
    this.panelState.offset = p.listOffset ?? 0;
    this.panelState.order = p.listSort ?? "";
    this.panelState.filters = parseFilterCSV(p.listFilter);
  }

  private columns() {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
  }

  private pageRows() {
    return [...(this.props.payload.records ?? [])];
  }

  private navigateList(patch: {
    listSearch?: string;
    listOffset?: number;
    listSort?: string;
    listFilter?: string;
  }): void {
    const p = this.props.payload;
    const url = this.env.services.router.workspaceUrl({
      actionId: p.actionId,
      menuId: p.menuId,
      viewType: VIEW_LIST,
      listSearch: patch.listSearch ?? this.panelState.search,
      listOffset: patch.listOffset ?? 0,
      listSort: patch.listSort ?? this.panelState.order ?? "",
      listFilter: patch.listFilter ?? this.panelState.filters.join(","),
      model: p.actionId ? "" : p.model,
    });
    this.env.services.action.navigate(url);
  }

  private applySearch(): void {
    this.navigateList({ listSearch: this.panelState.search, listOffset: 0 });
  }

  private applyPage(offset: number): void {
    this.navigateList({ listOffset: offset });
  }

  private applySort(fieldName: string): void {
    const current = this.panelState.order ?? "";
    let next = fieldName;
    if (current === fieldName) next = `-${fieldName}`;
    else if (current === `-${fieldName}`) next = "";
    this.navigateList({ listSort: next, listOffset: 0 });
  }

  private applyFilter(name: string): void {
    const next = toggleFilterName(this.panelState.filters, name);
    this.navigateList({ listFilter: next.join(","), listOffset: 0 });
  }

  private openRow(row: Record<string, unknown>): void {
    const id = Number(row.id ?? 0);
    if (id <= 0) return;
    this.env.services.action.openRecord({
      actionId: this.props.payload.actionId,
      menuId: this.props.payload.menuId,
      recordId: id,
      viewType: VIEW_FORM,
    });
  }

  private toggleRow(id: number, checked: boolean): void {
    if (checked) this.panelState.selectedIds.add(id);
    else this.panelState.selectedIds.delete(id);
    this.bump?.();
  }

  private toggleAll(checked: boolean, ids: number[]): void {
    this.panelState.selectedIds = checked ? new Set(ids) : new Set();
    this.bump?.();
  }

  private toolbarBusy(): boolean {
    return this.deleting || this.acting;
  }

  private headerObjectButtons(): SwcArchButton[] {
    return (this.props.payload.arch.header?.buttons ?? []).filter((btn) => btn.type === "object");
  }

  private async bulkDelete(): Promise<void> {
    const ids = [...this.panelState.selectedIds];
    if (ids.length === 0 || this.toolbarBusy()) return;
    const ok = await this.env.services.dialog.confirm(
      "Delete records",
      `Delete ${ids.length} selected record(s)?`,
    );
    if (!ok) return;
    this.deleting = true;
    this.bump?.();
    try {
      await this.env.services.rpc.unlink(this.props.payload.model, ids);
      this.panelState.selectedIds = new Set();
      this.env.services.notification.success("Deleted", `${ids.length} record(s) removed.`);
      this.applySearch();
    } catch (err) {
      this.env.services.notification.error(
        "Delete failed",
        err instanceof SwcError ? err.message : String(err),
      );
    } finally {
      this.deleting = false;
      this.bump?.();
    }
  }

  private async runHeaderObject(btn: SwcArchButton): Promise<void> {
    const ids = [...this.panelState.selectedIds];
    if (ids.length === 0 || this.toolbarBusy()) return;
    this.acting = true;
    this.bump?.();
    try {
      const result = await this.env.services.rpc.callMethod(this.props.payload.model, btn.name, ids[0], {
        active_ids: ids.join(","),
      });
      if (await this.env.services.action.applyCallResult(result)) {
        return;
      }
      this.env.services.notification.success(btn.string || btn.name, "Action completed.");
      this.applySearch();
    } catch (err) {
      this.env.services.notification.error(
        btn.string || btn.name,
        err instanceof SwcError ? err.message : String(err),
      );
    } finally {
      this.acting = false;
      this.bump?.();
    }
  }

  private renderRow(row: Record<string, unknown>) {
    const id = Number(row.id ?? 0);
    const cols = this.columns();
    return html`<tr class="sum-list-row sum-list-row--click" @click=${() => this.openRow(row)}>
      ${renderRowCheckbox(id, this.panelState.selectedIds.has(id), (rid, checked) =>
        this.toggleRow(rid, checked),
      )}
      ${cols.map((c) => {
        const display = row[`${c.name}_name`] ?? row[c.name];
        return html`<td class="sum-list-td">${String(display ?? "")}</td>`;
      })}
    </tr>`;
  }

  patch(): void {
    const tbody = this.el?.querySelector("tbody");
    if (tbody) {
      const rows = this.pageRows();
      patchKeyedChildren(
        tbody,
        rows.map((row) => ({
          key: String(row.id ?? 0),
          render: () => this.renderRow(row).render(),
        })),
      );
      return;
    }
    super.patch();
  }

  template() {
    const p = this.props.payload;
    const cols = this.columns();
    const rows = this.pageRows();
    const ids = rows.map((r) => Number(r.id ?? 0)).filter((id) => id > 0);
    const allSelected = ids.length > 0 && ids.every((id) => this.panelState.selectedIds.has(id));
    const filters = p.arch.search?.filters ?? [];

    return html`
      <div class="sum-list-view">
        ${renderCollectionToolbar({
          payload: p,
          viewType: VIEW_LIST,
          search: this.panelState.search,
          onSearch: () => this.applySearch(),
          onInput: (next) => {
            this.panelState.search = next;
          },
          extraPrimary: [
            this.panelState.selectedIds.size > 0
              ? html`<button
                  type="button"
                  class="sum-btn sum-btn--danger"
                  disabled=${this.toolbarBusy() ? "disabled" : undefined}
                  @click=${() => void this.bulkDelete()}
                >
                  Delete (${this.panelState.selectedIds.size})
                </button>`
              : "",
            this.panelState.selectedIds.size >= 2
              ? this.headerObjectButtons().map((btn) =>
                  headerButton(
                    btn.string || btn.name,
                    btn.class,
                    () => void this.runHeaderObject(btn),
                    this.toolbarBusy(),
                  ),
                )
              : "",
          ],
        })}
        ${renderSearchFilters({
          filters,
          active: this.panelState.filters,
          onToggle: (name) => this.applyFilter(name),
        })}
        ${renderControlPanel({
          payload: p,
          state: this.panelState,
          onPage: (o) => this.applyPage(o),
        })}
        <div class="sum-list-table-wrap">
          <table class="sum-list-table">
            <thead>
              <tr>
                ${renderSelectAllHeader(allSelected, (checked) => this.toggleAll(checked, ids))}
                ${cols.map((c) =>
                  renderSortHeader(c, this.panelState.order ?? "", (name) => this.applySort(name)),
                )}
              </tr>
            </thead>
            <tbody>
              ${forEach(rows, (row) => Number(row.id ?? 0), (row) => this.renderRow(row))}
            </tbody>
          </table>
        </div>
      </div>
    `;
  }
}
