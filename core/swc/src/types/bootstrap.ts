/** Bootstrap JSON injected by Go as window.__SWC_BOOTSTRAP__ */

export interface SwcBootstrapUser {
  id: number;
  name: string;
  login: string;
  image?: string;
  initials: string;
}

export interface SwcBootstrapCompany {
  id: number;
  name: string;
}

export interface SwcBootstrapMenuItem {
  id: string;
  name: string;
  action: string;
  module?: string;
  webIcon?: string;
}

export interface SwcBootstrapSidebarGroup {
  id: string;
  name: string;
  sequence: number;
  subMenus: SwcBootstrapMenuItem[];
}

export interface SwcBootstrapApp {
  kind?: "app" | "menu";
  module: string;
  name: string;
  action: string;
  description?: string;
  webIcon?: string;
  pinned?: boolean;
}

export interface SwcBootstrap {
  csrfToken: string;
  rpcUrl: string;
  swcApiBase: string;
  user: SwcBootstrapUser;
  company: SwcBootstrapCompany;
  companies: SwcBootstrapCompany[];
  activeCompanyId: number;
  showCompanySwitcher: boolean;
  topMenus: SwcBootstrapMenuItem[];
  sidebarMenus: SwcBootstrapSidebarGroup[];
  activeModuleId: string;
  activeMenuId: string;
  apps: SwcBootstrapApp[];
  pinnedApps: string[];
  appsNavAllowed: boolean;
  settingsNavAllowed: boolean;
  activityEnabled: boolean;
  busEnabled?: boolean;
  docsUrl: string;
  profileUrl: string;
  workspace?: SwcBootstrapWorkspace;
  translations?: Record<string, string>;
  toasts?: SwcToastMessage[];
}

export interface SwcBootstrapWorkspace {
  actionId: number;
  menuId: string;
  viewType: string;
  recordId: number;
  formEdit: boolean;
  listSearch: string;
}

export interface SwcToastMessage {
  kind: string;
  title: string;
  body: string;
  details?: string;
}

declare global {
  interface Window {
    __SWC_BOOTSTRAP__?: SwcBootstrap;
  }
}

export function readBootstrap(): SwcBootstrap {
  const boot = window.__SWC_BOOTSTRAP__;
  if (!boot) {
    throw new Error("SWC bootstrap missing on window.__SWC_BOOTSTRAP__");
  }
  return boot;
}
