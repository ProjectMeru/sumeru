import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcArchButton, SwcWorkspacePayload } from "../../types/workspace.js";
import { RecordStore, SwcRecord } from "../../model/record.js";
import { takePendingChildren } from "../../model/pending-children.js";
import { SwcError } from "../../runtime/error.js";
import { headerButton } from "../shared/view-toolbar.js";
import { renderFormToolbar } from "./form-chrome.js";
import { collectFormFields, renderFormSheet } from "./form-sheet.js";
import { initFormInteractions } from "./form-interactions.js";
import { validatePasswordMatchGroups } from "../../widgets/password-match.js";
import { FieldHost } from "../../widgets/field-host.js";
import { ChatterPanel } from "../chatter/ChatterPanel.js";
import { isFieldVisible } from "../../model/modifiers.js";
import { VIEW_FORM, VIEW_LIST } from "../../constants/routes.js";
import { runObjectAction } from "../shared/object-action.js";

interface FormViewProps {
  payload: SwcWorkspacePayload;
  inDialog?: boolean;
}

export class FormView extends SwcComponent<FormViewProps> {
  private recordStore!: RecordStore;
  private record!: SwcRecord;
  private snapshot: Record<string, unknown> = {};
  private editing = false;
  private saving = false;
  private acting = false;
  private error = "";
  private activeNotebookPages: Record<number, number> = {};
  private teardownInteractions: (() => void) | null = null;
  private fieldHost!: FieldHost;
  private chatterPanel!: ChatterPanel;

  override setup(): void {
    this.recordStore = new RecordStore(this.env.services.rpc);
    this.fieldHost = new FieldHost(this.env);
    this.initRecordState(this.props.payload);
    this.chatterPanel = new ChatterPanel(
      {
        model: this.props.payload.model,
        recordId: this.props.payload.recordId,
        csrfToken: this.props.payload.csrfToken,
      },
      this.env,
    );
    this.chatterPanel.callSetup();
  }

  override onPropsChanged(props: FormViewProps): void {
    this.initRecordState(props.payload);
    this.chatterPanel.updateProps({
      model: props.payload.model,
      recordId: props.payload.recordId,
      csrfToken: props.payload.csrfToken,
    });
    this.fieldHost.clear();
  }

  private initRecordState(payload: SwcWorkspacePayload): void {
    this.editing = payload.formEdit || payload.recordId <= 0;
    this.snapshot = { ...(payload.record ?? {}) };
    this.bindRecord(this.recordStore.fromPayload(payload.model, payload.recordId, this.snapshot));
  }

  override onMount(): void {
    this.bindFormInteractions();
  }

  override onWillUnmount(): void {
    this.teardownInteractions?.();
    this.teardownInteractions = null;
    this.fieldHost.clear();
    this.chatterPanel.destroy();
  }

  override patch(): void {
    this.teardownInteractions?.();
    super.patch();
  }

  override afterPatch(): void {
    this.bindFormInteractions();
  }

  private bindFormInteractions(): void {
    if (this.rootElement) {
      this.teardownInteractions = initFormInteractions(this.rootElement);
    }
  }

  private bindRecord(record: SwcRecord): void {
    this.record = record;
    this.record.onFieldChange = (field) => void this.handleFieldChange(field);
  }

  private async handleFieldChange(field: string): Promise<void> {
    if (this.isReadonly()) return;
    const result = await this.recordStore.applyOnchange(this.record, field);
    if (result?.warning) {
      this.env.services.notification.warning(result.warning.title, result.warning.message);
    }
    this.fieldHost.invalidate(field);
    this.rerender();
  }

  private renderFieldCached = (
    field: import("../../types/workspace.js").SwcArchField,
    record: SwcRecord,
    readonly: boolean,
  ): HTMLElement => {
    if (!isFieldVisible(field, record)) {
      const element = document.createElement("div");
      element.hidden = true;
      return element;
    }
    return this.fieldHost.render(field, record, readonly);
  };

  private isReadonly(): boolean {
    return !this.editing;
  }

  private toolbarBusy(): boolean {
    return this.saving || this.acting;
  }

  private fields() {
    const arch = this.props.payload.arch;
    return collectFormFields(arch.sheet, arch.header?.fields ?? []);
  }

  private startEdit(): void {
    this.editing = true;
    this.error = "";
    this.rerender();
  }

  private cancelEdit(): void {
    const payload = this.props.payload;
    if (payload.recordId <= 0) {
      const url = this.env.services.router.workspaceUrl({
        actionId: payload.actionId,
        menuId: payload.menuId,
        viewType: VIEW_LIST,
        recordId: 0,
        formEdit: false,
      });
      this.env.services.action.navigate(url);
      return;
    }
    this.bindRecord(this.recordStore.fromPayload(payload.model, payload.recordId, { ...this.snapshot }));
    this.editing = false;
    this.error = "";
    this.rerender();
  }

  private async reloadRecord(): Promise<void> {
    const payload = this.props.payload;
    if (payload.recordId <= 0) return;
    const fieldNames = this.fields().map((f) => f.name);
    const rows = await this.env.services.rpc.read(payload.model, [payload.recordId], fieldNames);
    if (!rows[0]) return;
    this.snapshot = { ...rows[0] };
    this.bindRecord(this.recordStore.fromPayload(payload.model, payload.recordId, this.snapshot));
    this.rerender();
  }

  private async save(): Promise<void> {
    if (this.rootElement && !validatePasswordMatchGroups(this.rootElement)) {
      this.error = "Passwords do not match.";
      this.rerender();
      return;
    }
    this.saving = true;
    this.error = "";
    this.rerender();
    try {
      const required = this.fields().filter((f) => f.required).map((f) => f.name);
      this.recordStore.validate(this.record, required);
      const payload = this.props.payload;
      const isNew = payload.recordId <= 0;
      const id = await this.recordStore.save(this.record);
      if (isNew && id > 0) {
        await this.savePendingChildren(id);
      }
      this.env.services.notification.success("Saved", "Record saved successfully.");
      if (isNew && id > 0) {
        this.env.services.action.openRecord({
          actionId: payload.actionId,
          menuId: payload.menuId,
          recordId: id,
          viewType: VIEW_FORM,
        });
        return;
      }
      this.editing = false;
      // Re-fetch the record so server-recomputed fields (e.g. invoice totals
      // after line edits) are reflected in the UI.
      await this.reloadRecord();
    } catch (err) {
      const message = err instanceof SwcError ? err.message : String(err);
      if (err instanceof SwcError && err.code === "validation") {
        this.error = message;
      } else {
        this.env.services.notification.error("Save failed", message);
      }
    } finally {
      this.saving = false;
      this.rerender();
    }
  }

  private async savePendingChildren(parentId: number): Promise<void> {
    const children = takePendingChildren(this.record);
    for (const child of children) {
      if (!child.comodel || !child.inverse) continue;
      const values: Record<string, unknown> = { ...child.values };
      values[child.inverse] = parentId;
      try {
        await this.env.services.rpc.create(child.comodel, values);
      } catch (err) {
        const message = err instanceof SwcError ? err.message : String(err);
        this.env.services.notification.error(
          "Save failed",
          `Could not create ${child.comodel} line: ${message}`,
        );
      }
    }
  }

  private async deleteRecord(): Promise<void> {
    const payload = this.props.payload;
    if (payload.recordId <= 0) return;
    const ok = await this.env.services.dialog.confirm("Delete record", "This cannot be undone.");
    if (!ok) return;
    try {
      await this.recordStore.unlink(this.record);
      this.env.services.notification.success("Deleted", "Record deleted.");
      this.env.services.action.navigate(
        this.env.services.router.workspaceUrl({
          actionId: payload.actionId,
          menuId: payload.menuId,
          viewType: VIEW_LIST,
          recordId: 0,
        }),
      );
    } catch (err) {
      this.env.services.notification.error(
        "Delete failed",
        err instanceof SwcError ? err.message : String(err),
      );
    }
  }

  private async duplicateRecord(): Promise<void> {
    const payload = this.props.payload;
    if (payload.recordId <= 0) return;
    try {
      const newId = await this.recordStore.duplicate(this.record);
      this.env.services.notification.success("Duplicated", "Record duplicated.");
      this.env.services.action.openRecord({
        actionId: payload.actionId,
        menuId: payload.menuId,
        recordId: newId,
        viewType: VIEW_FORM,
      });
    } catch (err) {
      this.env.services.notification.error(
        "Duplicate failed",
        err instanceof SwcError ? err.message : String(err),
      );
    }
  }

  private async runObjectButton(archButton: SwcArchButton): Promise<void> {
    const payload = this.props.payload;
    if (archButton.type !== "object" || payload.recordId <= 0) return;
    if (archButton.confirm && !window.confirm(archButton.confirm)) return;
    this.acting = true;
    this.error = "";
    this.rerender();
    const navigated = await runObjectAction(this.env, {
      model: payload.model,
      methodName: archButton.name,
      recordId: payload.recordId,
      buttonLabel: archButton.string || archButton.name,
      onSuccess: () => this.reloadRecord(),
    });
    this.acting = false;
    if (!navigated) this.rerender();
  }

  override template() {
    const payload = this.props.payload;
    const readonly = this.isReadonly();
    const busy = this.toolbarBusy();

    const sheet = renderFormSheet({
      env: this.env,
      sheet: payload.arch.sheet,
      record: this.record,
      readonly,
      hasImageField: payload.arch.formMeta?.hasImageField ?? false,
      activeNotebookPages: this.activeNotebookPages,
      onNotebookTab: (notebookIndex, pageIndex) => {
        this.activeNotebookPages = { ...this.activeNotebookPages, [notebookIndex]: pageIndex };
        this.rerender();
      },
      renderField: this.renderFieldCached,
      onStatButton: (name) => void this.runObjectButton({ name, string: name, type: "object" }),
    });

    const footerButtons = payload.arch.footer?.buttons ?? [];
    const showChatter = payload.arch.hasChatter && payload.recordId > 0;

    return html`
      <div class="sum-form-view sum-form-view--workspace-chrome${readonly ? " sum-form-view--readonly" : ""}">
        ${renderFormToolbar({
          payload,
          inDialog: this.props.inDialog,
          readonly,
          busy,
          saving: this.saving,
          record: this.record,
          fields: this.fields(),
          onStartEdit: () => this.startEdit(),
          onCancelEdit: () => this.cancelEdit(),
          onSave: () => void this.save(),
          onDelete: () => void this.deleteRecord(),
          onDuplicate: () => void this.duplicateRecord(),
          onObjectButton: (btn) => void this.runObjectButton(btn),
          renderField: this.renderFieldCached,
        })}
        ${this.error ? html`<div class="sum-flash sum-flash--error">${this.error}</div>` : ""}
        <div class="sum-form-layout${showChatter ? " sum-form-layout--with-chatter" : ""}">
          <div class="sum-form-sheet-bg">
            ${sheet}
            ${footerButtons.length > 0
              ? html`<div class="sum-form-footer">
                  ${footerButtons.map((archButton) =>
                    headerButton(
                      archButton.string || archButton.name,
                      archButton.class,
                      () => void this.runObjectButton(archButton),
                      busy,
                    ),
                  )}
                </div>`
              : ""}
          </div>
          ${showChatter ? this.chatterPanel.render() : ""}
        </div>
      </div>
    `;
  }
}
