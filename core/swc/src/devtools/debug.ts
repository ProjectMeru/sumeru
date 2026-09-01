/** Debug mode v2 — ?debug=1|true|assets and gear menu registry. */

import { initDevtoolsBridge } from "./bridge.js";
import { mountDevtoolsPanel, enablePicker } from "./panel.js";

const DEBUG_STORAGE_KEY = "sum.debug.mode";

export type DebugMode = "off" | "1" | "assets";

export function parseDebugParam(raw: string | null): DebugMode {
  if (!raw) return "off";
  const v = raw.trim().toLowerCase();
  if (v === "1" || v === "true" || v === "yes" || v === "on") return "1";
  if (v === "assets") return "assets";
  return "off";
}

export function getDebugMode(): DebugMode {
  if (typeof window === "undefined") return "off";
  const fromUrl = parseDebugParam(new URLSearchParams(window.location.search).get("debug"));
  if (fromUrl !== "off") {
    sessionStorage.setItem(DEBUG_STORAGE_KEY, fromUrl);
    return fromUrl;
  }
  const stored = sessionStorage.getItem(DEBUG_STORAGE_KEY);
  if (stored === "1" || stored === "assets") return stored;
  return "off";
}

export function isDebugMode(): boolean {
  return getDebugMode() !== "off";
}

export function isAssetsDebugMode(): boolean {
  return getDebugMode() === "assets";
}

export interface DebugMenuItem {
  id: string;
  label: string;
  section?: string;
  run: () => void;
}

const debugMenuItems: DebugMenuItem[] = [];

export function registerDebugMenuItem(item: DebugMenuItem): void {
  if (debugMenuItems.some((existing) => existing.id === item.id)) return;
  debugMenuItems.push(item);
}

export function listDebugMenuItems(): DebugMenuItem[] {
  return [...debugMenuItems];
}

function mountDebugGear(): void {
  if (document.getElementById("sum-debug-gear")) return;
  const wrap = document.createElement("div");
  wrap.id = "sum-debug-gear";
  wrap.className = "sum-debug-gear";
  wrap.innerHTML = `<button type="button" id="sum-debug-gear-btn" title="Developer tools">Debug</button>
    <div id="sum-debug-gear-menu" class="sum-debug-gear-menu" hidden></div>`;
  document.body.appendChild(wrap);
  const menu = wrap.querySelector("#sum-debug-gear-menu") as HTMLDivElement;
  const btn = wrap.querySelector("#sum-debug-gear-btn") as HTMLButtonElement;
  const renderMenu = () => {
    menu.innerHTML = "";
    for (const item of listDebugMenuItems()) {
      const b = document.createElement("button");
      b.type = "button";
      b.textContent = item.label;
      b.addEventListener("click", () => {
        item.run();
        menu.hidden = true;
      });
      menu.appendChild(b);
    }
  };
  btn.addEventListener("click", () => {
    renderMenu();
    menu.hidden = !menu.hidden;
  });
}

export function mountDebugPanel(): void {
  initDevtoolsBridge();
  registerDebugMenuItem({
    id: "open-vision",
    label: "Open SWC Vision",
    section: "tools",
    run: () => mountDevtoolsPanel(),
  });
  registerDebugMenuItem({
    id: "log-arch",
    label: "Log view arch to console",
    section: "tools",
    run: () => console.debug("[SWC] Use logViewArch from view load hooks"),
  });
  if (!isDebugMode()) return;
  if (!document.getElementById("sum-debug-panel")) {
    const el = document.createElement("aside");
    el.id = "sum-debug-panel";
    el.className = "sum-debug-panel";
    el.innerHTML = `<h4>SWC Debug</h4><p>Mode: ${getDebugMode()}. Arch and RPC logging enabled. Alt+click to inspect.</p>`;
    document.body.appendChild(el);
  }
  mountDebugGear();
  enablePicker();
}

export function logWorkspacePayload(label: string, payload: unknown): void {
  if (!isDebugMode()) return;
  console.debug(`[SWC ${label}]`, payload);
}

export function logViewArch(arch: unknown): void {
  if (!isDebugMode()) return;
  console.debug("[SWC arch]", arch);
}

export function debugFieldTitle(model: string, field: string, type?: string): string | undefined {
  if (!isDebugMode()) return undefined;
  const parts = [model, field];
  if (type) parts.push(type);
  return parts.join(" · ");
}
