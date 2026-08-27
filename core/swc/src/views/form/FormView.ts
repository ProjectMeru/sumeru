import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcArchButton, SwcWorkspacePayload } from "../../types/workspace.js";
import { RecordStore, SwcRecord } from "../../model/record.js";
import { SwcError } from "../../runtime/error.js";
import {
  headerButton,
  renderNewButton,
  renderReportActions,
  visibleFieldNames,
} from "../shared/view-toolbar.js";
import { collectFormFields, renderFormSheet } from "./form-sheet.js";
import { initFormInteractions } from "./form-interactions.js";
import { validatePasswordMatchGroups } from "../../login/password-match.js";
import { FieldHost } from "../../widgets/field-host.js";
import { ChatterPanel } from "../chatter/ChatterPanel.js";
import { isFieldVisible } from "../../model/modifiers.js";
import { VIEW_FORM, VIEW_LIST } from "../../constants/routes.js";

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

  setup(): void {
    this.recordStore = new RecordStore(this.env.services.rpc);
    this.fieldHost = new FieldHost(this.env);
    this.initRecordState(this.props.payload);
    this.bump = () => {
      if (this.el?.isConnected) this.patch();
    };
    this.chatterPanel = new ChatterPanel(
      {
        model: this.props.payload.model,
        recordId: this.props.payload.recordId,
        csrfToken: this.props.payload.csrfToken,
      },
      this.env,
    );
    this.chatterPanel.setup?.();
  }

  onPropsChanged(props: FormViewProps): void {
    this.initRecordState(props.payload);
    this.chatterPanel.updateProps({
      model: props.payload.model,
      recordId: props.payload.recordId,
      csrfToken: props.payload.csrfToken,
    });
    this.fieldHost.clear();
  }

  private initRecordState(p: SwcWorkspacePayload): void {
    this.editing = p.formEdit || p.recordId <= 0;
    this.snapshot = { ...(p.record ?? {}) };
    this.record = this.recordStore.fromPayload(p.model, p.recordId, this.snapshot);
    this.record.onFieldChange = (field) => void this.handleFieldChange(field);
  }

  private bump: (() => void) | null = null;

  onMount(): void {
    this.bindFormInteractions();
  }

  onWillUnmount(): void {
    this.teardownInteractions?.();
    this.teardownInteractions = null;
    this.fieldHost.clear();
    this.chatterPanel.destroy();
  }

  patch(): void {
    this.teardownInteractions?.();
    if (!this.el?.parentElement) return;
    const parent = this.el.parentElement;
    const oldEl = this.el;
    const next = this.template().render();
    parent.replaceChild(next, oldEl);
    this.el = next;
    this.bindFormInteractions();
  }

  private bindFormInteractions(): void {
    if (this.el) {
      this.teardownInteractions = initFormInteractions(this.el);
    }
  }

  private async handleFieldChange(field: string): Promise<void> {
    if (this.isReadonly()) return;
    const result = await this.recordStore.applyOnchange(this.record, field);
    if (result?.warning) {
      this.env.services.notification.warning(result.warning.title, result.warning.message);
    }
    this.fieldHost.clear();
    this.bump?.();
  }

  private renderFieldCached = (
    field: import("../../types/workspace.js").SwcArchField,
    record: SwcRecord,
    readonly: boolean,
  ): HTMLElement => {
    if (!isFieldVisible(field, record)) {
      const el = document.createElement("div");
      el.hidden = true;
      return el;
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

  private headerButtons(): SwcArchButton[] {
    return this.props.payload.arch.header?.buttons ?? [];
  }

  private startEdit(): void {
    this.editing = true;
    this.error = "";
    this.bump?.();
  }

  private cancelEdit(): void {
    const p = this.props.payload;
    if (p.recordId <= 0) {
      const url = this.env.services.router.workspaceUrl({
        actionId: p.actionId,
        menuId: p.menuId,
        viewType: VIEW_LIST,
        recordId: 0,
        formEdit: false,
      });
      this.env.services.action.navigate(url);
      return;
    }
    this.record = this.recordStore.fromPayload(p.model, p.recordId, { ...this.snapshot });
    this.record.onFieldChange = (field) => void this.handleFieldChange(field);
    this.editing = false;
    this.error = "";
    this.bump?.();
  }

  private async reloadRecord(): Promise<void> {
    const p = this.props.payload;
    if (p.recordId <= 0) return;
    const fieldNames = this.fields().map((f) => f.name);
    const rows = await this.env.services.rpc.read(p.model, [p.recordId], fieldNames);
    if (!rows[0]) return;
    this.snapshot = { ...rows[0] };
    this.record = this.recordStore.fromPayload(p.model, p.recordId, this.snapshot);
    this.record.onFieldChange = (field) => void this.handleFieldChange(field);
    this.bump?.();
  }

  private async save(): Promise<void> {
    if (this.el && !validatePasswordMatchGroups(this.el)) {
      this.error = "Passwords do not match.";
      this.bump?.();
      return;
    }
    this.saving = true;
    this.error = "";
    this.bump?.();
    try {
      const required = this.fields().filter((f) => f.required).map((f) => f.name);
      this.recordStore.validate(this.record, required);
      const id = await this.recordStore.save(this.record);
      this.env.services.notification.success("Saved", "Record saved successfully.");
      const p = this.props.payload;
      if (p.recordId <= 0 && id > 0) {
        this.env.services.action.openRecord({
          actionId: p.actionId,
          menuId: p.menuId,
          recordId: id,
          viewType: VIEW_FORM,
        });
        return;
      }
      this.snapshot = { ...this.record.data };
      this.editing = false;
      this.bump?.();
    } catch (err) {
      const message = err instanceof SwcError ? err.message : String(err);
      if (err instanceof SwcError && err.code === "validation") {
        this.error = message;
      } else {
        this.env.services.notification.error("Save failed", message);
      }
    } finally {
      this.saving = false;
      this.bump?.();
    }
  }

  private async deleteRecord(): Promise<void> {
    const p = this.props.payload;
    if (p.recordId <= 0) return;
    const ok = await this.env.services.dialog.confirm("Delete record", "This cannot be undone.");
    if (!ok) return;
    try {
      await this.recordStore.unlink(this.record);
      this.env.services.notification.success("Deleted", "Record deleted.");
      this.env.services.action.navigate(
        this.env.services.router.workspaceUrl({
          actionId: p.actionId,
          menuId: p.menuId,
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
    const p = this.props.payload;
    if (p.recordId <= 0) return;
    try {
      const newId = await this.recordStore.duplicate(this.record);
      this.env.services.notification.success("Duplicated", "Record duplicated.");
      this.env.services.action.openRecord({
        actionId: p.actionId,
        menuId: p.menuId,
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

  private async runObjectButton(btn: SwcArchButton): Promise<void> {
    const p = this.props.payload;
    if (btn.type !== "object" || p.recordId <= 0) return;
    this.acting = true;
    this.error = "";
    this.bump?.();
    try {
      const result = await this.env.services.rpc.callMethod(p.model, btn.name, p.recordId);
      if (await this.env.services.action.applyCallResult(result)) {
        return;
      }
      this.env.services.notification.success(btn.string || btn.name, "Action completed.");
      await this.reloadRecord();
    } catch (err) {
      this.env.services.notification.error(
        btn.string || btn.name,
        err instanceof SwcError ? err.message : String(err),
      );
    } finally {
      this.acting = false;
      this.bump?.();
    }
  }

  private renderToolbarPrimary(): Array<HTMLElement> {
    const p = this.props.payload;
    const busy = this.toolbarBusy();
    const items: HTMLElement[] = [];

    if (p.recordId > 0 && this.isReadonly()) {
      if (!this.props.inDialog) {
        items.push(renderNewButton(p));
        items.push(headerButton("Edit", undefined, () => this.startEdit(), busy));
        items.push(headerButton("Duplicate", undefined, () => void this.duplicateRecord(), busy));
        items.push(
          headerButton("Delete", "sum-btn--danger", () => void this.deleteRecord(), busy),
        );
      } else {
        items.push(headerButton("Edit", undefined, () => this.startEdit(), busy));
      }
    } else {
      items.push(headerButton("Save", "sum_highlight", () => void this.save(), busy));
      items.push(headerButton("Cancel", undefined, () => this.cancelEdit(), busy || this.saving));
    }

    for (const btn of this.headerButtons()) {
      if (btn.type !== "object") continue;
      items.push(
        headerButton(btn.string || btn.name, btn.class, () => void this.runObjectButton(btn), busy),
      );
    }

    return items;
  }

  template() {
    const p = this.props.payload;
    const readonly = this.isReadonly();
    const headerFields = p.arch.header?.fields ?? [];
    const exportFields = visibleFieldNames(this.fields());
    const reportActions = p.recordId > 0 ? renderReportActions(p, exportFields, p.recordId) : null;
    const toolbarItems = this.renderToolbarPrimary();
    const busy = this.toolbarBusy();

    const sheet = renderFormSheet({
      env: this.env,
      sheet: p.arch.sheet,
      record: this.record,
      readonly,
      hasImageField: p.arch.formMeta?.hasImageField ?? false,
      activeNotebookPages: this.activeNotebookPages,
      onNotebookTab: (notebookIndex, pageIndex) => {
        this.activeNotebookPages = { ...this.activeNotebookPages, [notebookIndex]: pageIndex };
        this.bump?.();
      },
      renderField: this.renderFieldCached,
      onStatButton: (name) => void this.runObjectButton({ name, string: name, type: "object" }),
    });

    const footerButtons = p.arch.footer?.buttons ?? [];
    const showChatter = p.arch.hasChatter && p.recordId > 0;

    return html`
      <div class="sum-form-view sum-form-view--workspace-chrome${readonly ? " sum-form-view--readonly" : ""}">
        <div class="sum-ws-record-toolbar sum-view-toolbar sum-form-toolbar">
          <div class="sum-statusbar-buttons sum-view-toolbar-primary">${toolbarItems}</div>
          ${headerFields.length > 0
            ? html`<div class="sum-statusbar-status sum-ws-toolbar-right">
                ${headerFields.map((f) => this.renderFieldCached(f, this.record, readonly))}
              </div>`
            : ""}
          ${reportActions ?? ""}
        </div>
        ${this.error ? html`<div class="sum-flash sum-flash--error">${this.error}</div>` : ""}
        <div class="sum-form-layout${showChatter ? " sum-form-layout--with-chatter" : ""}">
          <div class="sum-form-sheet-bg">
            ${sheet}
            ${footerButtons.length > 0
              ? html`<div class="sum-form-footer">
                  ${footerButtons.map((btn) =>
                    headerButton(
                      btn.string || btn.name,
                      btn.class,
                      () => void this.runObjectButton(btn),
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
