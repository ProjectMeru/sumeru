import { html, type TemplateResult } from "../../template/html.js";
import type { SwcArchButton, SwcArchField, SwcWorkspacePayload } from "../../types/workspace.js";
import type { SwcRecord } from "../../model/record.js";
import { isButtonVisible } from "../../model/modifiers.js";
import {
  exportFieldNamesCsv,
  headerButton,
  renderNewButton,
  renderReportActions,
} from "../shared/view-toolbar.js";

export interface FormToolbarOptions {
  payload: SwcWorkspacePayload;
  inDialog?: boolean;
  readonly: boolean;
  busy: boolean;
  saving?: boolean;
  record: SwcRecord;
  fields: SwcArchField[];
  onStartEdit: () => void;
  onCancelEdit: () => void;
  onSave: () => void;
  onDelete: () => void;
  onDuplicate: () => void;
  onObjectButton: (archButton: SwcArchButton) => void;
  renderField: (field: SwcArchField, record: SwcRecord, readonly: boolean) => HTMLElement;
}

/** Primary Save/Cancel/Edit/New/object buttons for the form record toolbar. */
export function renderFormToolbarPrimary(options: FormToolbarOptions): HTMLElement[] {
  const {
    payload,
    inDialog,
    readonly,
    busy,
    saving = false,
    record,
    onStartEdit,
    onCancelEdit,
    onSave,
    onDelete,
    onDuplicate,
    onObjectButton,
  } = options;
  const items: HTMLElement[] = [];
  const headerButtons = payload.arch.header?.buttons ?? [];

  if (payload.recordId > 0 && readonly) {
    if (!inDialog) {
      items.push(renderNewButton(payload));
      items.push(headerButton("Edit", undefined, onStartEdit, busy));
      items.push(headerButton("Duplicate", undefined, onDuplicate, busy));
      items.push(headerButton("Delete", "sum-btn--danger", onDelete, busy));
    } else {
      items.push(headerButton("Edit", undefined, onStartEdit, busy));
    }
  } else {
    items.push(headerButton("Save", "sum_highlight", onSave, busy));
    items.push(headerButton("Cancel", undefined, onCancelEdit, busy || saving));
  }

  for (const archButton of headerButtons) {
    if (archButton.type !== "object") continue;
    if (!isButtonVisible(archButton, record)) continue;
    items.push(
      headerButton(
        archButton.string || archButton.name,
        archButton.class,
        () => onObjectButton(archButton),
        busy,
      ),
    );
  }

  return items;
}

/** Full form workspace toolbar: primary buttons, header status fields, report actions. */
export function renderFormToolbar(options: FormToolbarOptions): TemplateResult {
  const { payload, readonly, fields, record, renderField } = options;
  const headerFields = payload.arch.header?.fields ?? [];
  const exportFields = exportFieldNamesCsv(fields);
  const reportActions =
    payload.recordId > 0 ? renderReportActions(payload, exportFields, payload.recordId) : null;

  return html`
    <div class="sum-ws-record-toolbar sum-view-toolbar sum-form-toolbar">
      <div class="sum-statusbar-buttons sum-view-toolbar-primary">
        ${renderFormToolbarPrimary(options)}
      </div>
      ${headerFields.length > 0
        ? html`<div class="sum-statusbar-status sum-ws-toolbar-right">
            ${headerFields.map((f) => renderField(f, record, readonly))}
          </div>`
        : ""}
      ${reportActions ?? ""}
    </div>
  `;
}
