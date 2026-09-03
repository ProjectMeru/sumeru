import { VIEW_HIERARCHY } from "../../constants/routes.js";
import { CollectionView } from "../shared/collection-view.js";
import { listColumns } from "../shared/arch-fields.js";
import { openWorkspaceRecord } from "../shared/collection-navigation.js";
import { renderArchListTable } from "../shared/list-table.js";

/** Hierarchy view — indented tree table using parent_field on arch. */
export class HierarchyView extends CollectionView {
  protected readonly collectionViewType = VIEW_HIERARCHY;

  private columns() {
    return listColumns(this.props.payload.arch);
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

  override template() {
    const payload = this.props.payload;
    const cols = this.columns();
    const rows = [...(payload.records ?? [])];
    const depthCache = new Map<number, number>();

    return this.renderShell(
      renderArchListTable({
        columns: cols,
        rows,
        onRowClick: (row) => openWorkspaceRecord(this.env, payload, row),
        firstCellStyle: (row) => {
          const depth = this.depth(row, rows, depthCache);
          return `padding-left: calc(0.75rem + ${depth * 1.25}rem)`;
        },
      }),
    );
  }
}
