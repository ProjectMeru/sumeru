import { SwcApp } from "./runtime/app.js";
import { SwcEnv } from "./runtime/env.js";
import { readBootstrap } from "./types/bootstrap.js";
import { RpcService } from "./services/rpc.js";
import { HttpService } from "./services/http.js";
import { NotificationService } from "./services/notification.js";
import { ActionService } from "./services/action.js";
import { RouterService } from "./services/router.js";
import { BusService } from "./services/bus.js";
import { DialogService } from "./services/dialog.js";
import { registerCoreServices } from "./services/service-registry.js";
import { ShellLayout } from "./shell/ShellLayout.js";
import { initShellChrome } from "./shell/shell-chrome.js";
import { initAppLauncher } from "./shell/app-launcher.js";
import { registerDefaultWidgets } from "./widgets/registry.js";
import { registry, type ViewConstructor, type MainComponentConstructor } from "./runtime/registry.js";
import { AddonLoader } from "./addon/loader.js";
import { ListView } from "./views/list/ListView.js";
import { FormView } from "./views/form/FormView.js";
import { KanbanView } from "./views/kanban/KanbanView.js";
import { PivotView } from "./views/pivot/PivotView.js";
import { GraphView } from "./views/graph/GraphView.js";
import { CalendarView } from "./views/calendar/CalendarView.js";
import { GanttView } from "./views/gantt/GanttView.js";
import { MapView } from "./views/map/MapView.js";
import { CohortView } from "./views/cohort/CohortView.js";
import { HierarchyView } from "./views/hierarchy/HierarchyView.js";
import { ActivityView } from "./views/activity/ActivityView.js";
import { loadTranslations } from "./i18n/translate.js";
import { mountDebugPanel } from "./devtools/debug.js";
import { initDevtoolsBridge } from "./devtools/bridge.js";

const VIEW_CONSTRUCTORS = {
  list: ListView,
  form: FormView,
  kanban: KanbanView,
  pivot: PivotView,
  graph: GraphView,
  calendar: CalendarView,
  gantt: GanttView,
  map: MapView,
  cohort: CohortView,
  hierarchy: HierarchyView,
  activity: ActivityView,
} satisfies Record<string, ViewConstructor>;

function registerCore(): void {
  registerDefaultWidgets();
  const views = registry.category("views");
  for (const [name, ViewClass] of Object.entries(VIEW_CONSTRUCTORS)) {
    views.add(name, ViewClass);
  }
  const main = registry.category("main_components");
  main.add("shell", ShellLayout as MainComponentConstructor);
}

function buildEnv(boot: ReturnType<typeof readBootstrap>): SwcEnv {
  const router = new RouterService();
  const services = {
    rpc: new RpcService(boot.rpcUrl, boot.csrfToken),
    http: new HttpService(boot.csrfToken),
    notification: new NotificationService(),
    action: new ActionService(router),
    router,
    bus: new BusService(),
    dialog: new DialogService(),
  };
  registerCoreServices(services);
  return new SwcEnv(boot, services);
}

function bootstrap(): void {
  registerCore();
  AddonLoader.registerFromGlobal();

  let boot;
  try {
    boot = readBootstrap();
  } catch {
    return;
  }

  const env = buildEnv(boot);
  env.services.action.setEnv(env);
  loadTranslations(boot.translations);
  initDevtoolsBridge();
  mountDebugPanel();
  initShellChrome(boot, env.services.http);
  initAppLauncher(boot, env.services.action);

  const mountEl = document.getElementById("swc-workspace");
  if (mountEl) {
    SwcApp.start(mountEl, env, ShellLayout);
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", bootstrap);
} else {
  bootstrap();
}

export { SwcApp, registry, SwcEnv };
