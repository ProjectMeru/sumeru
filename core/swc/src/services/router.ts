import {
  EDIT_ENABLED,
  Q_ACTION,
  Q_EDIT,
  Q_DOMAIN,
  Q_FILTER,
  Q_GROUPBY,
  Q_MENU_ID,
  Q_MODEL,
  Q_OFFSET,
  Q_RECORD_ID,
  Q_SEARCH,
  Q_SHELL,
  Q_SORT,
  Q_VIEW_TYPE,
  WEB_ROUTE,
} from "../constants/routes.js";

export interface WorkspaceRoute {
  actionId: number;
  menuId: string;
  viewType: string;
  recordId: number;
  formEdit: boolean;
  listSearch: string;
  model?: string;
  listFilter?: string;
  listSort?: string;
  listOffset?: number;
  listGroupBy?: string;
  listDomain?: string;
  /** SPA shell route: home | apps | settings */
  shell?: string;
}

export class RouterService {
  static searchParams(route: Partial<WorkspaceRoute>): URLSearchParams {
    const params = new URLSearchParams();
    if (route.actionId) params.set(Q_ACTION, String(route.actionId));
    if (route.menuId) params.set(Q_MENU_ID, route.menuId);
    if (route.viewType) params.set(Q_VIEW_TYPE, route.viewType);
    if (route.recordId) params.set(Q_RECORD_ID, String(route.recordId));
    if (route.formEdit) params.set(Q_EDIT, EDIT_ENABLED);
    if (route.listSearch) params.set(Q_SEARCH, route.listSearch);
    if (route.model) params.set(Q_MODEL, route.model);
    if (route.listFilter) params.set(Q_FILTER, route.listFilter);
    if (route.listSort) params.set(Q_SORT, route.listSort);
    if (route.listOffset) params.set(Q_OFFSET, String(route.listOffset));
    if (route.listGroupBy) params.set(Q_GROUPBY, route.listGroupBy);
    if (route.listDomain) params.set(Q_DOMAIN, route.listDomain);
    if (route.shell) params.set(Q_SHELL, route.shell);
    return params;
  }

  static buildUrl(route: Partial<WorkspaceRoute>): string {
    return `${WEB_ROUTE}?${RouterService.searchParams(route).toString()}`;
  }

  parse(location: { search: string } = window.location): WorkspaceRoute {
    const q = new URLSearchParams(location.search);
    return {
      actionId: Number(q.get(Q_ACTION) ?? "0"),
      menuId: q.get(Q_MENU_ID) ?? "",
      viewType: q.get(Q_VIEW_TYPE) ?? "",
      recordId: Number(q.get(Q_RECORD_ID) ?? "0"),
      formEdit: q.get(Q_EDIT) === EDIT_ENABLED,
      listSearch: q.get(Q_SEARCH) ?? "",
      model: q.get(Q_MODEL) ?? "",
      listFilter: q.get(Q_FILTER) ?? "",
      listSort: q.get(Q_SORT) ?? "",
      listOffset: Number(q.get(Q_OFFSET) ?? "0"),
      listGroupBy: q.get(Q_GROUPBY) ?? "",
      listDomain: q.get(Q_DOMAIN) ?? "",
      shell: q.get(Q_SHELL) ?? "",
    };
  }

  workspaceUrl(route: Partial<WorkspaceRoute>): string {
    return RouterService.buildUrl({ ...this.parse(), ...route });
  }

  push(route: Partial<WorkspaceRoute>): void {
    const url = this.workspaceUrl(route);
    window.history.pushState({}, "", url);
    window.dispatchEvent(new PopStateEvent("popstate"));
  }

  /** Replace the workspace query with an absolute route (no merge with current). */
  assign(route: WorkspaceRoute): void {
    const url = RouterService.buildUrl(route);
    window.history.pushState({}, "", url);
    window.dispatchEvent(new PopStateEvent("popstate"));
  }
}
