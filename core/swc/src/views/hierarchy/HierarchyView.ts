import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { VIEW_FORM, VIEW_HIERARCHY } from "../../constants/routes.js";
import { CollectionBarHost, mountCollectionBar } from "../shared/collection-bar-host.js";
import { formatFieldValue } from "../shared/field-display.js";
import { forEach } from "../../template/helpers.js";

interface HierarchyViewProps {
  payload: SwcWorkspacePayload;
}

/** Hierarchy view — indented tree table using parent_field on arch. */
export class HierarchyView extends SwcComponent<HierarchyViewProps> {
  private collectionBar!: CollectionBarHost;

  override setup(): void {
    this.collectionBar = mountCollectionBar(this.props.payload, VIEW_HIERARCHY, this.env);
  }

  override onPropsChanged(props: HierarchyViewProps): void {
    this.collectionBar.updateProps({ payload: props.payload, viewType: VIEW_HIERARCHY });
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
  }

  private columns() {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
  }

  private parentField(): string {
    return this.props.payload.arch.hierarchy?.parentField ?? "parent_id";
  }

  private depth(row: Record<string, unknown>, rows: Record<string, unknown>[], cache: Map<number, number>): number {
    const id = Number(row.id ?? 0);
    if (cache.has(id)) return cache.get(id)!;
    const parentKey = this.parentField();
    const parentId = Number(row[parentKey] ?? 0);
    if (parentId <= 0) {
      cache.set(id, 0);
      return 0;
    }
    const parent = rows.find((r) => Number(r.id ?? 0) === parentId);
    const d = parent ? this.depth(parent, rows, cache) + 1 : 0;
    cache.set(id, d);
    return d;
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
    const depthCache = new Map<number, number>();

    return html`
      <div class="sum-collection-view sum-hierarchy-view">
        ${this.collectionBar.renderOrPatch()}
        <div class="sum-list-table-wrap">
          <table class="sum-list-table">
            <thead>
              <tr>
                ${cols.map((c) => html`<th class="sum-list-th">${c.string ?? c.name}</th>`)}
              </tr>
            </thead>
            <tbody>
              ${forEach(rows, (row) => Number(row.id ?? 0), (row) => {
                const depth = this.depth(row, rows, depthCache);
                const pad = `${depth * 1.25}rem`;
                return html`<tr class="sum-list-row sum-list-row--click" @click=${() => this.openRow(row)}>
                  ${cols.map((c, i) =>
                    html`<td class="sum-list-td" style=${i === 0 ? `padding-left: calc(0.75rem + ${pad})` : ""}>
                      ${formatFieldValue(row, c)}
                    </td>`,
                  )}
                </tr>`;
              })}
            </tbody>
          </table>
        </div>
      </div>
    `;
  }
}
