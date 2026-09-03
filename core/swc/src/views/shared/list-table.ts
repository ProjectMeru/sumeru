import { html, type TemplateResult } from "../../template/html.js";
import type { SwcArchField } from "../../types/workspace.js";
import { forEach } from "../../template/helpers.js";
import { formatFieldValue } from "./field-display.js";

export interface ArchListTableOptions {
  columns: SwcArchField[];
  rows: Record<string, unknown>[];
  onRowClick: (row: Record<string, unknown>) => void;
  rowKey?: (row: Record<string, unknown>) => number;
  firstCellStyle?: (row: Record<string, unknown>) => string | undefined;
}

/** Simple arch-driven table for hierarchy, activity, and similar collection views. */
export function renderArchListTable(options: ArchListTableOptions): TemplateResult {
  const rowKey = options.rowKey ?? ((row) => Number(row.id ?? 0));

  return html`
    <div class="sum-list-table-wrap">
      <table class="sum-list-table">
        <thead>
          <tr>
            ${options.columns.map((c) => html`<th class="sum-list-th">${c.string ?? c.name}</th>`)}
          </tr>
        </thead>
        <tbody>
          ${forEach(options.rows, rowKey, (row) => {
            const firstStyle = options.firstCellStyle?.(row);
            return html`<tr class="sum-list-row sum-list-row--click" @click=${() => options.onRowClick(row)}>
              ${options.columns.map((c, i) =>
                html`<td
                  class="sum-list-td"
                  style=${i === 0 && firstStyle ? firstStyle : undefined}
                >
                  ${formatFieldValue(row, c)}
                </td>`,
              )}
            </tr>`;
          })}
        </tbody>
      </table>
    </div>
  `;
}
