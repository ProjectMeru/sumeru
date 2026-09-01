import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { VIEW_ACTIVITY, VIEW_FORM } from "../../constants/routes.js";
import { CollectionBarHost, mountCollectionBar } from "../shared/collection-bar-host.js";
import { formatFieldValue } from "../shared/field-display.js";
import { forEach } from "../../template/helpers.js";

interface ActivityViewProps {
  payload: SwcWorkspacePayload;
}

/** Activity view — list-style table for sys.activity (or activity arch on any model). */
export class ActivityView extends SwcComponent<ActivityViewProps> {
  private collectionBar!: CollectionBarHost;

  override setup(): void {
    this.collectionBar = mountCollectionBar(this.props.payload, VIEW_ACTIVITY, this.env);
  }

  override onPropsChanged(props: ActivityViewProps): void {
    this.collectionBar.updateProps({ payload: props.payload, viewType: VIEW_ACTIVITY });
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
  }

  private columns() {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
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

  override template() {
    const payload = this.props.payload;
    const cols = this.columns();
    const rows = [...(payload.records ?? [])];

    return html`
      <div class="sum-collection-view sum-activity-view">
        ${this.collectionBar.renderOrPatch()}
        <div class="sum-list-table-wrap">
          <table class="sum-list-table">
            <thead>
              <tr>
                ${cols.map((c) => html`<th class="sum-list-th">${c.string ?? c.name}</th>`)}
              </tr>
            </thead>
            <tbody>
              ${forEach(rows, (row) => Number(row.id ?? 0), (row) =>
                html`<tr class="sum-list-row sum-list-row--click" @click=${() => this.openRow(row)}>
                  ${cols.map((c) => html`<td class="sum-list-td">${formatFieldValue(row, c)}</td>`)}
                </tr>`,
              )}
            </tbody>
          </table>
        </div>
      </div>
    `;
  }
}
