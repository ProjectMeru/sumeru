import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import type { DebugWorkspaceContext } from "../../devtools/debug.js";
import { ListView } from "../list/ListView.js";
import { SwcError } from "../../runtime/error.js";
import { registry, type ViewConstructor } from "../../runtime/registry.js";
import { logWorkspacePayload, logViewArch, updateDebugDrawer } from "../../devtools/debug.js";
import { collectFormFields } from "../form/form-sheet.js";
import { ShellPageView } from "../../shell/ShellPageView.js";
import { syncWorkspaceViewTabs } from "../../shell/view-tab-sync.js";
import { syncWorkspaceBreadcrumbs } from "../../shell/breadcrumb-sync.js";
import { ACTION_CLOSED, Q_ACTION, RECORD_UPDATED, SWC_API_BASE } from "../../constants/routes.js";
import { ACTIVITY_CONTEXT, activityContextFromWorkspace } from "../../shell/activity-context.js";
import { moduleNameFromWebPath, RouterService } from "../../services/router.js";
import { runWillStart } from "../../runtime/lifecycle.js";

type ViewInstance = SwcComponent & { callSetup(): void; render(): HTMLElement };

export class WorkspaceRouter extends SwcComponent {
  private payload: SwcWorkspacePayload | null = null;
  private loading = true;
  private error = "";
  private activeView: ViewInstance | null = null;
  private activeViewType = "";
  private onPopState: (() => void) | null = null;
  private unsubActionClosed: (() => void) | null = null;
  private unsubRecordUpdated: (() => void) | null = null;

  override setup(): void {
    void this.load();
  }

  override onMount(): void {
    this.onPopState = (): void => {
      void this.load();
    };
    window.addEventListener("popstate", this.onPopState);
    this.unsubActionClosed = this.env.services.bus.subscribe(ACTION_CLOSED, () => {
      if (this.payload?.target === "new") return;
      void this.load();
    });
    this.unsubRecordUpdated = this.env.services.bus.subscribe(RECORD_UPDATED, (payload) => {
      const msg = payload as { model?: string; id?: number };
      if (!this.payload || !msg.model) return;
      if (msg.model !== this.payload.model) return;
      if (msg.id && this.payload.recordId && msg.id !== this.payload.recordId) return;
      void this.load();
    });
  }

  override onWillUnmount(): void {
    if (this.onPopState) {
      window.removeEventListener("popstate", this.onPopState);
      this.onPopState = null;
    }
    this.unsubActionClosed?.();
    this.unsubActionClosed = null;
    this.unsubRecordUpdated?.();
    this.unsubRecordUpdated = null;
  }

  private deferModuleEntryToServer(): boolean {
    const { pathname, search, hash } = window.location;
    if (!moduleNameFromWebPath(pathname)) {
      return false;
    }
    if (new URLSearchParams(search).get(Q_ACTION)) {
      return false;
    }
    window.location.assign(`${pathname}${search}${hash}`);
    return true;
  }

  private async load(): Promise<void> {
    if (this.deferModuleEntryToServer()) {
      return;
    }
    this.loading = true;
    this.error = "";
    this.rerender();
    try {
      this.payload = await this.fetchWorkspace();
      logWorkspacePayload("workspace", this.payload);
      logViewArch(this.payload.arch);
      syncWorkspaceViewTabs(this.payload.viewTabs);
      syncWorkspaceBreadcrumbs(this.payload.breadcrumbs);
      if (!this.openActionDialog(this.payload)) {
        this.syncView();
      }
      this.emitActivityContext(this.payload);
      void updateDebugDrawer(this.env.bootstrap, this.debugContext(this.payload));
    } catch (err) {
      this.error = err instanceof SwcError ? err.message : String(err);
      this.emitActivityContext({
        actionId: 0,
        menuId: "",
        viewType: "",
        model: "",
        recordId: 0,
        formEdit: false,
        csrfToken: "",
        arch: { type: "", model: "", fields: [] },
        viewTabs: [],
        breadcrumbs: [],
      });
    } finally {
      this.loading = false;
      this.rerender();
    }
  }

  private debugContext(payload: SwcWorkspacePayload): DebugWorkspaceContext {
    const viewType = payload.viewType || payload.arch.type;
    const fields =
      viewType === "form"
        ? collectFormFields(payload.arch.sheet, payload.arch.header?.fields ?? [])
        : undefined;
    return {
      model: payload.model,
      recordId: payload.recordId,
      actionId: payload.actionId,
      viewType,
      menuId: payload.menuId,
      fields,
      recordData: payload.record,
      arch: payload.arch,
      favorites: payload.favorites,
      listSearch: payload.listSearch,
      hasChatter: payload.arch.hasChatter,
    };
  }

  private async fetchWorkspace(): Promise<SwcWorkspacePayload> {
    const params = RouterService.searchParams(this.env.services.router.parse());
    const base = this.env.bootstrap.swcApiBase || SWC_API_BASE;
    return this.env.services.http.getJSON(`${base}/workspace?${params.toString()}`);
  }

  private createView(type: string, payload: SwcWorkspacePayload): ViewInstance {
    const ViewClass = (registry.category("views").get(type) ?? ListView) as ViewConstructor;
    const view = new ViewClass({ payload }, this.env) as unknown as ViewInstance;
    view.callSetup();
    void runWillStart(view).then(() => {
      if (view.rootElement?.isConnected) view.patch();
      else this.rerender();
    });
    return view;
  }

  private openActionDialog(payload: SwcWorkspacePayload): boolean {
    document.body.classList.toggle("sum-body--fullscreen", payload.target === "fullscreen");
    const viewType = payload.viewType || payload.arch.type;
    if (payload.target !== "new" || viewType !== "form") return false;
    this.env.services.action.openPayloadDialog(payload);
    this.activeView?.destroy();
    this.activeView = null;
    this.activeViewType = "";
    return true;
  }

  private syncView(): void {
    if (!this.payload) return;
    const type = this.payload.viewType || this.payload.arch.type;
    if (this.activeView && this.activeViewType === type) {
      this.activeView.updateProps({ payload: this.payload });
      return;
    }
    this.activeView?.destroy();
    this.activeView = this.createView(type, this.payload);
    this.activeViewType = type;
  }

  private emitActivityContext(payload: SwcWorkspacePayload): void {
    this.env.services.bus.emit(ACTIVITY_CONTEXT, activityContextFromWorkspace(payload));
  }

  private renderView(): HTMLElement {
    if (!this.payload || !this.activeView) return document.createElement("div");
    return this.activeView.renderOrPatch();
  }

  /** Reload workspace payload (e.g. after bus event). */
  reload(): void {
    void this.fetchWorkspace()
      .then((payload) => {
        this.payload = payload;
        syncWorkspaceViewTabs(payload.viewTabs);
        syncWorkspaceBreadcrumbs(payload.breadcrumbs);
        if (!this.openActionDialog(payload)) {
          this.syncView();
        }
        this.emitActivityContext(payload);
        this.patch();
        void updateDebugDrawer(this.env.bootstrap, this.debugContext(payload));
      })
      .catch((err) => {
        this.error = err instanceof SwcError ? err.message : String(err);
        this.patch();
      });
  }

  override template() {
    const route = this.env.services.router.parse();
    if (route.shell === "home" || route.shell === "apps" || route.shell === "settings") {
      document.body.classList.remove("sum-body--fullscreen");
      const page = route.shell as "home" | "apps" | "settings";
      const shellView = new ShellPageView({ boot: this.env.bootstrap, page }, this.env);
      return html`<div class="sum-workspace-root sum-workspace-root--shell">${shellView.render()}</div>`;
    }
    if (this.loading) {
      return html`<div class="sum-workspace-loading">Loading workspace…</div>`;
    }
    if (this.error) {
      return html`<div class="sum-flash sum-flash--error">${this.error}</div>`;
    }
    if (!this.payload) return html`<div></div>`;
    return html`<div class="sum-workspace-root sum-workspace-view">${this.renderView()}</div>`;
  }
}
