import type { SwcBootstrap } from "../types/bootstrap.js";

const STORAGE_KEY = "sumeru_shell_tabs";
const CHANNEL_NAME = "sumeru_shell_tabs";
const TAB_ID_KEY = "sumeru_shell_tab_id";
const LOGOUT_ROUTE = "/web/logout";

export type TabRegistry = Record<string, number>;

type TabMessage =
  | { type: "sync"; registry: TabRegistry }
  | { type: "touch"; tabId: string; at: number };

function readRegistry(): TabRegistry {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return {};
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== "object" || parsed === null) return {};
    return parsed as TabRegistry;
  } catch {
    return {};
  }
}

function writeRegistry(registry: TabRegistry): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(registry));
}

export function pruneStaleTabs(registry: TabRegistry, staleMs: number, now = Date.now()): TabRegistry {
  const out: TabRegistry = {};
  for (const [id, seen] of Object.entries(registry)) {
    if (now-seen <= staleMs) {
      out[id] = seen;
    }
  }
  return out;
}

export function registerTab(registry: TabRegistry, tabId: string, now = Date.now()): TabRegistry {
  return { ...registry, [tabId]: now };
}

export function unregisterTab(registry: TabRegistry, tabId: string): TabRegistry {
  const next = { ...registry };
  delete next[tabId];
  return next;
}

export function registryIsEmpty(registry: TabRegistry): boolean {
  return Object.keys(registry).length === 0;
}

function csrfTokenFromPage(boot: SwcBootstrap): string {
  const meta = document.querySelector('meta[name="sum-csrf-token"]');
  const fromMeta = meta?.getAttribute("content")?.trim();
  if (fromMeta) return fromMeta;
  return boot.csrfToken;
}

function postLogoutBeacon(csrfToken: string): void {
  if (!csrfToken) return;
  const body = new URLSearchParams({ csrf_token: csrfToken });
  if (typeof navigator.sendBeacon === "function") {
    navigator.sendBeacon(LOGOUT_ROUTE, body);
    return;
  }
  void fetch(LOGOUT_ROUTE, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body,
    keepalive: true,
  });
}

function touchSession(csrfToken: string): void {
  if (!csrfToken) return;
  const body = new URLSearchParams({ csrf_token: csrfToken });
  void fetch("/web/session/touch", {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body,
    keepalive: true,
  });
}

function tabIdFromSession(): string {
  let id = sessionStorage.getItem(TAB_ID_KEY);
  if (!id) {
    id = crypto.randomUUID();
    sessionStorage.setItem(TAB_ID_KEY, id);
  }
  return id;
}

export function isReloadNavigation(): boolean {
  if (typeof performance === "undefined") return false;
  const nav = performance.getEntriesByType("navigation")[0] as PerformanceNavigationTiming | undefined;
  return nav?.type === "reload";
}

export function initShellTabPresence(boot: SwcBootstrap): void {
  if (!boot.closeSessionOnLastTab || typeof window === "undefined") {
    return;
  }

  const tabId = tabIdFromSession();
  const csrfToken = csrfTokenFromPage(boot);
  const pingMs = Math.max(5, boot.sessionBrowserPingSeconds ?? 30) * 1000;
  const staleMs = Math.max(pingMs * 2, (boot.sessionBrowserGraceSeconds ?? 90) * 1000);

  let registry = registerTab(pruneStaleTabs(readRegistry(), staleMs), tabId);
  writeRegistry(registry);

  const channel =
    typeof BroadcastChannel !== "undefined" ? new BroadcastChannel(CHANNEL_NAME) : null;

  const broadcast = (msg: TabMessage): void => {
    channel?.postMessage(msg);
  };

  const applyRegistry = (next: TabRegistry): void => {
    registry = next;
    writeRegistry(registry);
    broadcast({ type: "sync", registry });
  };

  channel?.addEventListener("message", (event: MessageEvent<TabMessage>) => {
    const msg = event.data;
    if (!msg || typeof msg !== "object") return;
    if (msg.type === "sync") {
      registry = registerTab(pruneStaleTabs(msg.registry, staleMs), tabId);
      writeRegistry(registry);
    }
    if (msg.type === "touch") {
      registry = registerTab(readRegistry(), msg.tabId, msg.at);
      writeRegistry(registry);
    }
  });

  window.addEventListener("pageshow", () => {
    applyRegistry(registerTab(readRegistry(), tabId));
    touchSession(csrfToken);
  });

  window.addEventListener("pagehide", (ev: PageTransitionEvent) => {
    if (ev.persisted || isReloadNavigation()) {
      return;
    }
    const next = unregisterTab(readRegistry(), tabId);
    writeRegistry(next);
    broadcast({ type: "sync", registry: next });
    if (registryIsEmpty(next)) {
      postLogoutBeacon(csrfToken);
    }
  });

  const tick = (): void => {
    const now = Date.now();
    const next = registerTab(pruneStaleTabs(readRegistry(), staleMs, now), tabId, now);
    writeRegistry(next);
    broadcast({ type: "touch", tabId, at: now });
    touchSession(csrfToken);
  };

  tick();
  window.setInterval(tick, pingMs);
}
