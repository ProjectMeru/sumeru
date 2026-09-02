import { html } from "../../template/html.js";
import { VIEW_ACTIVITY, VIEW_FORM } from "../../constants/routes.js";
import { CollectionView } from "../shared/collection-view.js";
import { formatFieldValue } from "../shared/field-display.js";
import { forEach } from "../../template/helpers.js";

/** Activity view — list-style table for sys.activity (or activity arch on any model). */
export class ActivityView extends CollectionView {
  protected readonly collectionViewType = VIEW_ACTIVITY;

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
