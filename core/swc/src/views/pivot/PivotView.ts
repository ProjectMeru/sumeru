import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";

interface PivotViewProps {
  payload: SwcWorkspacePayload;
}

export class PivotView extends SwcComponent<PivotViewProps> {
  template() {
    const pivot = this.props.payload.arch.pivot;
    if (!pivot) {
      return html`<div class="sum-pivot-view sum-pivot-view--empty">No pivot data</div>`;
    }
    return html`
      <div class="sum-pivot-view">
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
                  const val = pivot.values[row]?.[col] ?? 0;
                  return html`<td>${String(val)}</td>`;
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
