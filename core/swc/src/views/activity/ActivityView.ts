import { VIEW_ACTIVITY } from "../../constants/routes.js";
import { CollectionView } from "../shared/collection-view.js";
import { listColumns } from "../shared/arch-fields.js";
import { openWorkspaceRecord } from "../shared/collection-navigation.js";
import { renderArchListTable } from "../shared/list-table.js";

/** Activity view — list-style table for sys.activity (or activity arch on any model). */
export class ActivityView extends CollectionView {
  protected readonly collectionViewType = VIEW_ACTIVITY;

  private columns() {
    return listColumns(this.props.payload.arch);
  }

  override template() {
    const payload = this.props.payload;

    return this.renderShell(
      renderArchListTable({
        columns: this.columns(),
        rows: [...(payload.records ?? [])],
        onRowClick: (row) => openWorkspaceRecord(this.env, payload, row),
      }),
    );
  }
}
