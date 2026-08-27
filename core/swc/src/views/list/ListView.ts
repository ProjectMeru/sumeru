import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcArchButton, SwcWorkspacePayload } from "../../types/workspace.js";
import { headerButton } from "../shared/view-toolbar.js";
import { SwcError } from "../../runtime/error.js";
import {
  parseFilterCSV,
  renderControlPanel,
  renderRowCheckbox,
  renderSelectAllHeader,
  renderSortHeader,
  type ControlPanelState,
} from "./control-panel.js";
import { forEach } from "../../template/helpers.js";
import { VIEW_FORM, VIEW_LIST } from "../../constants/routes.js";
import { runObjectAction } from "../shared/object-action.js";
import { formatFieldValue } from "../shared/field-display.js";
import { CollectionBarHost, mountCollectionBar } from "../shared/collection-bar-host.js";
import { navigateCollectionQuery } from "../shared/collection-query.js";

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
  private collectionBar!: CollectionBarHost;

  override setup(): void {
    this.syncFromPayload(this.props.payload);
    this.collectionBar = mountCollectionBar(this.props.payload, VIEW_LIST, this.env, this.listExtras());
  }

  override onPropsChanged(props: ListViewProps): void {
    this.syncFromPayload(props.payload);
    this.panelState.selectedIds = new Set();
    this.collectionBar.updateProps({ payload: props.payload, viewType: VIEW_LIST, extraPrimary: this.listExtras() });
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
  }

  private syncFromPayload(payload: SwcWorkspacePayload): void {
    this.panelState.search = payload.listSearch ?? "";
    this.panelState.offset = payload.listOffset ?? 0;
    this.panelState.order = payload.listSort ?? "";
    this.panelState.filters = parseFilterCSV(payload.listFilter);
  }

  private listExtras() {
    return [
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
        ? this.headerObjectButtons().map((archButton) =>
            headerButton(
              archButton.string || archButton.name,
              archButton.class,
              () => void this.runHeaderObject(archButton),
              this.toolbarBusy(),
            ),
          )
        : "",
    ];
  }

  private refreshBarExtras(): void {
    this.collectionBar.updateProps({
      payload: this.props.payload,
      viewType: VIEW_LIST,
      extraPrimary: this.listExtras(),
    });
  }

  private columns() {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
  }

  private pageRows() {
    return [...(this.props.payload.records ?? [])];
  }

  private reloadCollection(): void {
    navigateCollectionQuery(this.env, this.props.payload, VIEW_LIST, { listOffset: 0 });
  }

  private applyPage(offset: number): void {
    navigateCollectionQuery(this.env, this.props.payload, VIEW_LIST, { listOffset: offset });
  }

  private applySort(fieldName: string): void {
    const current = this.panelState.order ?? "";
    let next = fieldName;
    if (current === fieldName) next = `-${fieldName}`;
    else if (current === `-${fieldName}`) next = "";
    navigateCollectionQuery(this.env, this.props.payload, VIEW_LIST, { listSort: next, listOffset: 0 });
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
    this.refreshBarExtras();
    this.rerender();
  }

  private toggleAll(checked: boolean, ids: number[]): void {
    this.panelState.selectedIds = checked ? new Set(ids) : new Set();
    this.refreshBarExtras();
    this.rerender();
  }

  private toolbarBusy(): boolean {
    return this.deleting || this.acting;
  }

  private headerObjectButtons(): SwcArchButton[] {
    return (this.props.payload.arch.header?.buttons ?? []).filter(
      (archButton) => archButton.type === "object",
    );
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
    this.refreshBarExtras();
    this.rerender();
    try {
      await this.env.services.rpc.unlink(this.props.payload.model, ids);
      this.panelState.selectedIds = new Set();
      this.env.services.notification.success("Deleted", `${ids.length} record(s) removed.`);
      this.reloadCollection();
    } catch (err) {
      this.env.services.notification.error(
        "Delete failed",
        err instanceof SwcError ? err.message : String(err),
      );
    } finally {
      this.deleting = false;
      this.refreshBarExtras();
      this.rerender();
    }
  }

  private async runHeaderObject(archButton: SwcArchButton): Promise<void> {
    const ids = [...this.panelState.selectedIds];
    if (ids.length === 0 || this.toolbarBusy()) return;
    this.acting = true;
    this.refreshBarExtras();
    this.rerender();
    const navigated = await runObjectAction(this.env, {
      model: this.props.payload.model,
      methodName: archButton.name,
      recordId: ids[0],
      extraArgs: { active_ids: ids.join(",") },
      buttonLabel: archButton.string || archButton.name,
      onSuccess: () => this.reloadCollection(),
    });
    this.acting = false;
    if (!navigated) {
      this.refreshBarExtras();
      this.rerender();
    }
  }

  private renderRow(row: Record<string, unknown>) {
    const id = Number(row.id ?? 0);
    const cols = this.columns();
    return html`<tr class="sum-list-row sum-list-row--click" @click=${() => this.openRow(row)}>
      ${renderRowCheckbox(id, this.panelState.selectedIds.has(id), (rid, checked) =>
        this.toggleRow(rid, checked),
      )}
      ${cols.map((c) => html`<td class="sum-list-td">${formatFieldValue(row, c)}</td>`)}
    </tr>`;
  }

  override template() {
    const payload = this.props.payload;
    const cols = this.columns();
    const rows = this.pageRows();
    const ids = rows.map((r) => Number(r.id ?? 0)).filter((id) => id > 0);
    const allSelected = ids.length > 0 && ids.every((id) => this.panelState.selectedIds.has(id));

    return html`
      <div class="sum-collection-view sum-list-view">
        ${this.collectionBar.renderOrPatch()}
        ${renderControlPanel({
          payload,
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
