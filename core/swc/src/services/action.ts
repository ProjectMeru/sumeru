import {
  ACTION_CLOSED,
  Q_ACTION,
  Q_EDIT,
  Q_MENU_ID,
  Q_VIEW_TYPE,
  SWC_API_BASE,
  VIEW_FORM,
  WEB_ROUTE,
  EDIT_ENABLED,
} from "../constants/routes.js";
import type { SwcEnv } from "../runtime/env.js";
import type { SwcWorkspacePayload } from "../types/workspace.js";
import { FormView } from "../views/form/FormView.js";
import { RouterService, type WorkspaceRoute } from "./router.js";

export type OpenRecordOpts = {
  actionId: number;
  menuId: string;
  recordId: number;
  viewType?: string;
};

export type ActionOpenSpec = {
  model?: string;
  actionId?: number;
  viewType?: string;
  recordId?: number;
  target?: string;
};

export type ActionCallResult = {
  redirect?: string;
  close?: boolean;
  open?: ActionOpenSpec;
};

export class ActionService {
  private env?: SwcEnv;
  private dialogView: { destroy(): void } | null = null;

  constructor(private readonly router?: RouterService) {}

  setEnv(env: SwcEnv): void {
    this.env = env;
  }

  navigate(url: string): void {
    if ((url.startsWith(`${WEB_ROUTE}?`) || url === WEB_ROUTE) && this.router) {
      const parsed = new URL(url, window.location.origin);
      this.router.assign(this.router.parse({ search: parsed.search }));
      return;
    }
    window.location.assign(url);
  }

  openWindowAction(actionId: number, menuId?: string, extra?: Record<string, string>): void {
    const params = new URLSearchParams({ [Q_ACTION]: String(actionId) });
    if (menuId) params.set(Q_MENU_ID, menuId);
    for (const [k, v] of Object.entries(extra ?? {})) {
      if (v) params.set(k, v);
    }
    this.navigate(`${WEB_ROUTE}?${params.toString()}`);
  }

  openRecord({ actionId, menuId, recordId, viewType = VIEW_FORM }: OpenRecordOpts): void {
    const route: WorkspaceRoute = {
      actionId,
      menuId,
      recordId,
      viewType,
      formEdit: false,
      listSearch: "",
    };
    if (this.router) {
      this.router.push(route);
      return;
    }
    this.navigate(RouterService.buildUrl(route));
  }

  /** Apply an object-action RPC result. Returns true when navigation or a dialog handled it. */
  async applyCallResult(result: unknown): Promise<boolean> {
    if (result === true || result === false || result == null) return false;
    const body = result as ActionCallResult;
    if (body.close) {
      this.closeDialog();
      this.env?.services.bus.emit(ACTION_CLOSED, {});
      return true;
    }
    if (body.open) {
      const target = body.open.target || "dialog";
      if (target === "dialog") {
        await this.openFormDialog(body.open);
        return true;
      }
      this.navigateCurrent(body.open);
      return true;
    }
    if (body.redirect) {
      const parsed = this.parseRedirectOpen(body.redirect);
      if (parsed) {
        return this.applyCallResult({ open: parsed });
      }
      this.navigate(body.redirect);
      return true;
    }
    return false;
  }

  private navigateCurrent(open: ActionOpenSpec): void {
    this.navigate(
      RouterService.buildUrl({
        actionId: open.actionId ?? 0,
        viewType: open.viewType || VIEW_FORM,
        recordId: open.recordId ?? 0,
        model: open.model,
        formEdit: false,
        listSearch: "",
        menuId: "",
      }),
    );
  }

  private parseRedirectOpen(redirect: string): ActionOpenSpec | null {
    if (!redirect.startsWith(`${WEB_ROUTE}?`) && !redirect.startsWith("?")) return null;
    const q = new URLSearchParams(redirect.slice(redirect.indexOf("?") + 1));
    const model = q.get("model") ?? "";
    const recordId = Number(q.get("id") ?? "0");
    if (!model || recordId <= 0) return null;
    return {
      model,
      actionId: Number(q.get(Q_ACTION) ?? "0") || undefined,
      viewType: q.get(Q_VIEW_TYPE) || VIEW_FORM,
      recordId,
      target: q.get("target") || "dialog",
    };
  }

  private async openFormDialog(open: ActionOpenSpec): Promise<void> {
    const env = this.env;
    if (!env) {
      this.navigateCurrent({ ...open, target: "current" });
      return;
    }
    const params = new URLSearchParams();
    if (open.model) params.set("model", open.model);
    if (open.actionId) params.set(Q_ACTION, String(open.actionId));
    if (open.recordId) params.set("id", String(open.recordId));
    params.set(Q_VIEW_TYPE, open.viewType || VIEW_FORM);
    params.set(Q_EDIT, EDIT_ENABLED);
    const base = env.bootstrap.swcApiBase || SWC_API_BASE;
    const payload = await env.services.http.getJSON<SwcWorkspacePayload>(
      `${base}/workspace?${params.toString()}`,
    );
    this.closeDialog();
    const view = new FormView({ payload, inDialog: true }, env);
    view.setup?.();
    this.dialogView = view;
    const title = payload.arch.title || payload.arch.model || "Wizard";
    void env.services.dialog.openHost(title, view.render()).then(() => {
      view.destroy();
      if (this.dialogView === view) this.dialogView = null;
    });
  }

  closeDialog(): void {
    this.env?.services.dialog.close();
  }
}
