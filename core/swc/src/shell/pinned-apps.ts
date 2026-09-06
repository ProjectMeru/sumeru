import type { HttpService } from "../services/http.js";
import { readJSON } from "../util/shell-storage.js";

const KEY_PINNED_LEGACY = "sumeru:pinned-apps";

let pinnedCache: string[] = [];
let cacheLoaded = false;

interface PinnedAppsResponse {
  modules: string[];
}

async function persistPinnedApps(http: HttpService, modules: string[]): Promise<string[]> {
  const data = await http.postJSON<PinnedAppsResponse>("/web/user/pinned-apps", { modules });
  return Array.isArray(data.modules) ? data.modules.map(String) : modules;
}

export function resetPinnedAppsState(modules: string[] = []): void {
  pinnedCache = modules.slice();
  cacheLoaded = true;
}

export function loadPinnedApps(initial: string[]): string[] {
  if (!cacheLoaded) {
    pinnedCache = initial.slice();
    cacheLoaded = true;
  }
  return pinnedCache.slice();
}

export function setPinnedCache(modules: string[]): void {
  pinnedCache = modules.slice();
  cacheLoaded = true;
}

export function getPinnedApps(): string[] {
  return pinnedCache.slice();
}

export function togglePinnedApp(http: HttpService, moduleName: string): string[] {
  const mod = String(moduleName || "").trim();
  if (!mod) return getPinnedApps();

  const previous = getPinnedApps();
  let pins = previous.slice();
  if (pins.includes(mod)) {
    pins = pins.filter((m) => m !== mod);
  } else {
    pins = [mod, ...pins];
  }
  setPinnedCache(pins);

  persistPinnedApps(http, pins)
    .then((saved) => {
      setPinnedCache(saved);
    })
    .catch(() => {
      setPinnedCache(previous);
    });

  return pins;
}

export function applyTopNavFilter(): void {
  const nav = document.querySelector(".sum-top-nav");
  if (!nav) return;

  const moduleItems = [...nav.querySelectorAll<HTMLElement>(".top-menu-item--module")];
  if (moduleItems.length === 0) return;

  const pins = getPinnedApps();
  const visibleMods = new Set<string>(pins);

  moduleItems.forEach((el) => {
    const mod = el.getAttribute("data-module") ?? "";
    el.classList.toggle("is-topbar-hidden", !visibleMods.has(mod));
  });

  const activeEl = nav.querySelector<HTMLElement>(".top-menu-item.active");
  activeEl?.scrollIntoView?.({ inline: "nearest", block: "nearest", behavior: "instant" });
}

export function initPinnedApps(http: HttpService, initial: string[]): void {
  loadPinnedApps(initial);

  const legacy = readJSON<string[]>(KEY_PINNED_LEGACY, []);
  if (getPinnedApps().length === 0 && legacy.length > 0) {
    persistPinnedApps(http, legacy)
      .then((saved) => {
        setPinnedCache(saved);
        try {
          localStorage.removeItem(KEY_PINNED_LEGACY);
        } catch (err) {
          console.warn("pinned-apps: legacy key cleanup failed", err);
        }
        applyTopNavFilter();
      })
      .catch((err) => {
        console.warn("pinned-apps: migration failed", err);
      });
  }

  applyTopNavFilter();
}
