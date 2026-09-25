/** Debug mode v2 — ?debug=1|true|assets and gear menu registry. */

import { initDevtoolsBridge } from "./bridge.js";
import { mountDevtoolsPanel, enablePicker } from "./panel.js";
import type { SwcBootstrap } from "../types/bootstrap.js";

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
    document.body.appendChild(el);
  }
  mountDebugGear();
  enablePicker();
}

export interface DebugWorkspaceContext {
  model: string;
  recordId: number;
}

/** Refresh debug aside with bootstrap + optional server ACL trace. */
export async function updateDebugPanel(
  boot: SwcBootstrap,
  workspace?: DebugWorkspaceContext,
): Promise<void> {
  if (!isDebugMode()) return;
  const el = document.getElementById("sum-debug-panel");
  if (!el) return;

  const lines: string[] = [
    `<h4>SWC Debug</h4>`,
    `<p>Mode: <strong>${getDebugMode()}</strong> · User ${boot.user.id} · Company ${boot.activeCompanyId}</p>`,
  ];
  if (workspace?.model) {
    lines.push(`<p>Record: <code>${workspace.model}:${workspace.recordId}</code></p>`);
    try {
      const url = `/web/debug/access?model=${encodeURIComponent(workspace.model)}`;
      const res = await fetch(url, { credentials: "same-origin", headers: { Accept: "application/json" } });
      if (res.ok) {
        const trace = (await res.json()) as {
          group_ids?: number[];
          rule_domain?: string;
          rule_domain_truncated?: boolean;
          fields?: { name: string; read_denied?: boolean; write_denied?: boolean }[];
        };
        const denied = (trace.fields ?? []).filter((f) => f.read_denied || f.write_denied);
        lines.push(`<p>Groups: ${(trace.group_ids ?? []).join(", ") || "—"}</p>`);
        if (denied.length > 0) {
          lines.push(
            `<p>Field ACL: ${denied.map((f) => `${f.name}${f.read_denied ? " (read)" : ""}${f.write_denied ? " (write)" : ""}`).join(", ")}</p>`,
          );
        }
        if (trace.rule_domain) {
          const suffix = trace.rule_domain_truncated ? " …" : "";
          lines.push(`<pre class="sum-debug-trace">${escapeHtml(trace.rule_domain)}${suffix}</pre>`);
        }
      } else {
        lines.push(`<p class="sum-debug-trace-err">ACL trace: HTTP ${res.status}</p>`);
      }
    } catch {
      lines.push(`<p class="sum-debug-trace-err">ACL trace failed</p>`);
    }
  } else {
    lines.push(`<p>Arch and RPC logging enabled. Alt+click to inspect.</p>`);
  }
  el.innerHTML = lines.join("\n");
}

function escapeHtml(raw: string): string {
  return raw.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
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
