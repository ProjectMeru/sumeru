import type { HttpService } from "../services/http.js";
import { NotificationService } from "../services/notification.js";
import type { SwcBootstrap } from "../types/bootstrap.js";
import type { SwcEnv } from "../runtime/env.js";
import { initActivityPanel } from "./activity-panel.js";
import { initActivityMessages } from "./activity-messages.js";
import { initActivityLog } from "./activity-log.js";
import { initInternalMessagesBanner } from "./activity-im-banner.js";
import { initMessagesUnreadPoll } from "./messages-unread.js";
import { ACTIVITY_CONTEXT, activityBoardHref, emptyActivityContext, type ActivityContextPayload } from "./activity-context.js";
import { initHomeDashboard } from "./home-dashboard.js";
import { initPinnedApps } from "./pinned-apps.js";
import { initSidebar } from "./sidebar.js";
import { initCompanySwitcher } from "./company-switcher.js";
import { initViewTabNavigation } from "./view-tab-sync.js";
import { initBreadcrumbNavigation } from "./breadcrumb-sync.js";
import { initNotificationBell } from "./notification-bell.js";
import type { BusService } from "../services/bus.js";

export function initShellChrome(boot: SwcBootstrap, http: HttpService, bus?: BusService, env?: SwcEnv): void {
  const shell = document.getElementById("sum-shell");
  if (!shell) return;

  initSidebar(shell);
  initViewTabNavigation();
  initBreadcrumbNavigation();

  initInternalMessagesBanner(boot);

  if (boot.activityEnabled) {
    initActivityPanel(shell);
    if (env) {
      initActivityBoardLink(env);
      initActivityMessages(env);
      initActivityLog(env);
      initMessagesUnreadPoll(http, boot.swcApiBase);
      env.services.bus.emit(ACTIVITY_CONTEXT, emptyActivityContext());
    }
  }

  initPinnedApps(http, boot.pinnedApps ?? []);
  initHomeDashboard(http);
  initCompanySwitcher(boot, http);

  if (bus && boot.user?.id) {
    initNotificationBell(http, bus, boot.user.id);
  }

  new NotificationService().bootstrap(boot.toasts);
}

function initActivityBoardLink(env: SwcEnv): void {
  const link = document.getElementById("sum-activity-board");
  if (!link) return;
  const paint = (ctx: ActivityContextPayload): void => {
    link.setAttribute("href", activityBoardHref(ctx));
  };
  paint(emptyActivityContext());
  env.services.bus.subscribe(ACTIVITY_CONTEXT, (payload) => {
    paint(payload as ActivityContextPayload);
  });
}
