import { html } from "../../template/html.js";
import { VIEW_PIVOT } from "../../constants/routes.js";
import { CollectionView } from "../shared/collection-view.js";
import { renderPivotExportLink } from "../shared/view-toolbar.js";

export class PivotView extends CollectionView {
  protected readonly collectionViewType = VIEW_PIVOT;
  private collapsedRows = new Set<string>();

  private toggleRow(row: string): void {
    if (this.collapsedRows.has(row)) {
      this.collapsedRows.delete(row);
    } else {
      this.collapsedRows.add(row);
    }
    this.rerender();
  }

  override template() {
    const pivot = this.props.payload.arch.pivot;
    const exportLink = renderPivotExportLink(this.props.payload);
    if (!pivot) {
      return html`<div class="sum-collection-view sum-pivot-view sum-pivot-view--empty">
        ${this.collectionBar.renderOrPatch()}
        No pivot data
      </div>`;
    }
    return html`
      <div class="sum-collection-view sum-pivot-view">
        ${this.collectionBar.renderOrPatch()}
        ${exportLink ?? ""}
        <table class="sum-pivot-table">
          <thead>
            <tr>
              <th></th>
              ${pivot.colLabels.map((c) => html`<th>${c}</th>`)}
            </tr>
          </thead>
          <tbody>
            ${pivot.rowLabels.map(
              (row) => html`<tr class=${this.collapsedRows.has(row) ? "is-collapsed" : ""}>
                <th>
                  <button type="button" class="sum-pivot-row-toggle" @click=${() => this.toggleRow(row)}>
                    ${this.collapsedRows.has(row) ? "+" : "−"}
                  </button>
                  ${row}
                </th>
                ${this.collapsedRows.has(row)
                  ? pivot.colLabels.map(() => html`<td>…</td>`)
                  : pivot.colLabels.map((col) => {
                      const fieldValue = pivot.values[row]?.[col] ?? 0;
                      return html`<td>${String(fieldValue)}</td>`;
                    })}
              </tr>`,
            )}
          </tbody>
        </table>
        <p class="sum-pivot-measure">${pivot.measureLabel}</p>
      </div>
    `;
  }
}
