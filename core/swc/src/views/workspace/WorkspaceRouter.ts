import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { ListView } from "../list/ListView.js";
import { useEffect } from "../../runtime/hooks.js";
import { SwcError } from "../../runtime/error.js";
import { registry, type ViewConstructor } from "../../runtime/registry.js";
import { logWorkspacePayload, logViewArch } from "../../devtools/debug.js";
import { ShellPageView } from "../../shell/ShellPageView.js";
import { syncWorkspaceViewTabs } from "../../shell/view-tab-sync.js";
import { ACTION_CLOSED, RECORD_UPDATED, SWC_API_BASE } from "../../constants/routes.js";
import { RouterService } from "../../services/router.js";
import { runWillStart } from "../../runtime/lifecycle.js";

type ViewInstance = SwcComponent & { callSetup(): void; render(): HTMLElement };

export class WorkspaceRouter extends SwcComponent {
  private payload: SwcWorkspacePayload | null = null;
  private loading = true;
  private error = "";
  private activeView: ViewInstance | null = null;
  private activeViewType = "";

  override setup(): void {
    const load = async (): Promise<void> => {
      this.loading = true;
      this.error = "";
      this.rerender();
      try {
        this.payload = await this.fetchWorkspace();
        logWorkspacePayload("workspace", this.payload);
        logViewArch(this.payload.arch);
        syncWorkspaceViewTabs(this.payload.viewTabs);
        this.syncView();
      } catch (err) {
        this.error = err instanceof SwcError ? err.message : String(err);
      } finally {
        this.loading = false;
        this.rerender();
      }
    };

    void load();
    useEffect(() => {
      const onNav = (): void => void load();
      window.addEventListener("popstate", onNav);
      return () => window.removeEventListener("popstate", onNav);
    });

    useEffect(() => {
      return this.env.services.bus.subscribe(ACTION_CLOSED, () => {
        void load();
      });
    });

    useEffect(() => {
      return this.env.services.bus.subscribe(RECORD_UPDATED, (payload) => {
        const msg = payload as { model?: string; id?: number };
        if (!this.payload || !msg.model) return;
        if (msg.model !== this.payload.model) return;
        if (msg.id && this.payload.recordId && msg.id !== this.payload.recordId) return;
        void load();
      });
    });
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
        this.syncView();
        this.patch();
      })
      .catch((err) => {
        this.error = err instanceof SwcError ? err.message : String(err);
        this.patch();
      });
  }

  override template() {
    const route = this.env.services.router.parse();
    if (route.shell === "home" || route.shell === "apps" || route.shell === "settings") {
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
