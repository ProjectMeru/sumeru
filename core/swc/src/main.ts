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
import { RecordService } from "./model/record.js";
import { CommandService } from "./services/command.js";
import { getDebugMode, mountDebugPanel, updateDebugPanel } from "./devtools/debug.js";
import type { SwcServices } from "./runtime/env.js";
import type { SwcBootstrap } from "./types/bootstrap.js";
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
import { IframeView } from "./views/iframe/IframeView.js";
import { loadTranslations } from "./i18n/translate.js";
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
  iframe: IframeView,
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

function registerCoreCommands(boot: SwcBootstrap, services: SwcServices): void {
  const { command, action } = services;
  command.register({
    id: "nav.home",
    label: "Go to home",
    run: () => action.navigate("/web"),
  });
  command.register({
    id: "nav.apps",
    label: "Open app launcher",
    run: () => document.getElementById("sum-topbar-search-open")?.click(),
  });
  command.register({
    id: "debug.toggle",
    label: "Toggle debug mode",
    run: () => {
      const next = getDebugMode() === "off" ? "1" : "off";
      const url = new URL(window.location.href);
      if (next === "off") {
        sessionStorage.removeItem("sum.debug.mode");
        url.searchParams.delete("debug");
      } else {
        sessionStorage.setItem("sum.debug.mode", next);
        url.searchParams.set("debug", next);
      }
      window.location.assign(url.toString());
    },
  });
  if (boot.showCompanySwitcher) {
    command.register({
      id: "company.focus",
      label: "Focus company switcher",
      run: () => document.querySelector<HTMLSelectElement>(".sum-company-switcher-select")?.focus(),
    });
  }
}

function buildEnv(boot: ReturnType<typeof readBootstrap>): SwcEnv {
  const router = new RouterService();
  const bus = new BusService();
  const command = new CommandService();
  const rpc = new RpcService(boot.rpcUrl, boot.csrfToken);
  const services = {
    rpc,
    http: new HttpService(boot.csrfToken),
    notification: new NotificationService(),
    action: new ActionService(router),
    router,
    bus,
    dialog: new DialogService(),
    record: new RecordService(rpc, bus),
    command,
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
  registerCoreCommands(boot, env.services);
  mountDebugPanel();
  void updateDebugPanel(boot);
  initShellChrome(boot, env.services.http);
  initAppLauncher(boot, env.services.action, env.services.command);

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
