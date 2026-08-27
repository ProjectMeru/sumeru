import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { VIEW_PIVOT } from "../../constants/routes.js";
import { CollectionBarHost, mountCollectionBar } from "../shared/collection-bar-host.js";

interface PivotViewProps {
  payload: SwcWorkspacePayload;
}

export class PivotView extends SwcComponent<PivotViewProps> {
  private collectionBar!: CollectionBarHost;

  override setup(): void {
    this.collectionBar = mountCollectionBar(this.props.payload, VIEW_PIVOT, this.env);
  }

  override onPropsChanged(props: PivotViewProps): void {
    this.collectionBar.updateProps({ payload: props.payload, viewType: VIEW_PIVOT });
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
  }

  override template() {
    const pivot = this.props.payload.arch.pivot;
    if (!pivot) {
      return html`<div class="sum-collection-view sum-pivot-view sum-pivot-view--empty">
        ${this.collectionBar.renderOrPatch()}
        No pivot data
      </div>`;
    }
    return html`
      <div class="sum-collection-view sum-pivot-view">
        ${this.collectionBar.renderOrPatch()}
        <table class="sum-pivot-table">
          <thead>
            <tr>
              <th></th>
              ${pivot.colLabels.map((c) => html`<th>${c}</th>`)}
            </tr>
          </thead>
          <tbody>
            ${pivot.rowLabels.map(
              (row) => html`<tr>
                <th>${row}</th>
                ${pivot.colLabels.map((col) => {
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
