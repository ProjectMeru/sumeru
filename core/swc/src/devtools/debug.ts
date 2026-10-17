/** Debug mode — ?debug=1|assets, top-bar manager, drawer, field inspector. */

import { initDevtoolsBridge } from "./bridge.js";
import { enablePicker } from "./panel.js";
import type { SwcBootstrap } from "../types/bootstrap.js";
import type { SwcArchField, SwcSavedSearch, SwcViewArch } from "../types/workspace.js";
import { buildFieldDebugInfo, parseDebugFieldKey, type FieldDebugInfo } from "./debug-field-info.js";
import { SwcRecord } from "../model/record.js";

const DEBUG_STORAGE_KEY = "sum.debug.mode";
const FIELD_INSPECTOR_KEY = "sum.debug.fieldInspector";

export type DebugMode = "off" | "1" | "assets";

export function parseDebugParam(raw: string | null): DebugMode {
  if (!raw) return "off";
  let v = raw.trim().toLowerCase();
  if (v.startsWith("true/")) {
    v = "true";
  }
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

export function navigateDebugMode(mode: DebugMode): void {
  const url = new URL(window.location.href);
  if (mode === "off") {
    sessionStorage.removeItem(DEBUG_STORAGE_KEY);
    url.searchParams.delete("debug");
  } else {
    sessionStorage.setItem(DEBUG_STORAGE_KEY, mode);
    url.searchParams.set("debug", mode);
  }
  window.location.assign(url.toString());
}

export function toggleDebugMode(): void {
  navigateDebugMode(getDebugMode() === "off" ? "1" : "off");
}

export function isFieldInspectorEnabled(): boolean {
  if (typeof sessionStorage === "undefined") return false;
  return sessionStorage.getItem(FIELD_INSPECTOR_KEY) === "1";
}

export function toggleFieldInspector(): void {
  const next = !isFieldInspectorEnabled();
  if (next) sessionStorage.setItem(FIELD_INSPECTOR_KEY, "1");
  else sessionStorage.removeItem(FIELD_INSPECTOR_KEY);
}

export function applyDebugBodyClasses(): void {
  if (typeof document === "undefined") return;
  const on = isDebugMode();
  document.body.classList.toggle("sum-debug--active", on);
  document.body.classList.toggle("sum-debug--assets", getDebugMode() === "assets");
  document.body.classList.toggle("sum-debug--field-inspector", on && isFieldInspectorEnabled());
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

export interface DebugWorkspaceContext {
  model: string;
  recordId: number;
  actionId?: number;
  viewType?: string;
  menuId?: string;
  fields?: SwcArchField[];
  recordData?: Record<string, unknown>;
  arch?: SwcViewArch;
  favorites?: SwcSavedSearch[];
  listSearch?: string;
  hasChatter?: boolean;
}

export function getDebugMenuContext(): {
  boot: SwcBootstrap | null;
  workspace: DebugWorkspaceContext | undefined;
  aclTrace: AclTrace | null;
} {
  return { boot: lastBoot, workspace: lastWorkspace, aclTrace: lastAclTrace };
}

let lastBoot: SwcBootstrap | null = null;
let lastWorkspace: DebugWorkspaceContext | undefined;
let lastAclTrace: AclTrace | null = null;
let selectedFieldKey: string | null = null;
let selectedFieldEl: HTMLElement | null = null;
const selectionListeners: Array<() => void> = [];

interface AclTrace {
  group_ids?: number[];
  rule_domain?: string;
  rule_domain_truncated?: boolean;
  fields?: { name: string; read_denied?: boolean; write_denied?: boolean }[];
}

export function onDebugSelectionChange(fn: () => void): void {
  selectionListeners.push(fn);
}

function notifySelectionChange(): void {
  for (const fn of selectionListeners) fn();
}

export function selectDebugFieldFromElement(el: HTMLElement): void {
  if (selectedFieldEl) selectedFieldEl.classList.remove("sum-debug-field--selected");
  selectedFieldEl = el;
  selectedFieldKey = el.getAttribute("data-sum-debug-field");
  el.classList.add("sum-debug-field--selected");
  notifySelectionChange();
}

export function ensureDebugDrawer(): HTMLElement {
  let el = document.getElementById("sum-debug-drawer");
  if (el) return el;
  el = document.createElement("aside");
  el.id = "sum-debug-drawer";
  el.className = "sum-debug-drawer sum-debug-drawer--collapsed";
  el.setAttribute("aria-label", "Developer debug");
  el.innerHTML = `<header class="sum-debug-drawer__header">
      <h2 class="sum-debug-drawer__title">Developer</h2>
      <button type="button" class="sum-icon-btn sum-debug-drawer__close" aria-label="Close debug panel">×</button>
    </header>
    <div class="sum-debug-drawer__body"></div>`;
  document.body.appendChild(el);
  el.querySelector(".sum-debug-drawer__close")?.addEventListener("click", () => {
    el?.classList.add("sum-debug-drawer--collapsed");
  });
  return el;
}

function section(title: string, body: HTMLElement, sectionId?: string): HTMLElement {
  const wrap = document.createElement("section");
  wrap.className = "sum-debug-drawer__section";
  if (sectionId) wrap.dataset.debugSection = sectionId;
  const h = document.createElement("h3");
  h.className = "sum-debug-drawer__section-title";
  h.textContent = title;
  wrap.append(h, body);
  return wrap;
}

function para(text: string): HTMLParagraphElement {
  const p = document.createElement("p");
  p.className = "sum-debug-drawer__text";
  p.textContent = text;
  return p;
}

function kvRow(label: string, value: string): HTMLDivElement {
  const row = document.createElement("div");
  row.className = "sum-debug-kv";
  const dt = document.createElement("div");
  dt.className = "sum-debug-kv__label";
  dt.textContent = label;
  const dd = document.createElement("div");
  dd.className = "sum-debug-kv__value";
  dd.textContent = value;
  row.append(dt, dd);
  return row;
}

function kvList(pairs: Array<[string, string]>): HTMLDivElement {
  const wrap = document.createElement("div");
  wrap.className = "sum-debug-kv-list";
  for (const [label, value] of pairs) {
    if (value === "") continue;
    wrap.appendChild(kvRow(label, value));
  }
  return wrap;
}

export function openDebugDrawer(sectionId?: string): void {
  const drawer = ensureDebugDrawer();
  drawer.classList.remove("sum-debug-drawer--collapsed");
  if (sectionId) {
    requestAnimationFrame(() => {
      drawer
        .querySelector(`[data-debug-section="${sectionId}"]`)
        ?.scrollIntoView({ behavior: "smooth", block: "start" });
    });
  }
}

function renderFieldDetailSection(info: FieldDebugInfo): HTMLElement {
  const rows: Array<[string, string]> = [
    ["Model", info.model],
    ["Field", info.fieldName],
    ["Label", info.label],
    ["Type", info.type],
    ["Widget", info.widget],
    ["Relation", info.relation],
    ["Value", info.valueDisplay],
    ["Evaluated", info.modifierSummary],
    ["Static invisible", String(info.staticInvisible)],
    ["Static readonly", String(info.staticReadonly)],
    ["Static required", String(info.staticRequired)],
    ["Expr invisible", info.invisibleExpr],
    ["Expr readonly", info.readonlyExpr],
    ["Expr required", info.requiredExpr],
    ["Domain", info.domainDisplay],
    ["Context", info.contextDisplay],
    ["Related", info.relatedDisplay],
    [
      "ACL",
      info.aclReadDenied || info.aclWriteDenied
        ? [info.aclReadDenied ? "read denied" : "", info.aclWriteDenied ? "write denied" : ""]
            .filter(Boolean)
            .join(", ")
        : "allowed",
    ],
  ];
  return kvList(rows);
}

export function mountDebugEnvironment(boot: SwcBootstrap): void {
  lastBoot = boot;
  initDevtoolsBridge();
  applyDebugBodyClasses();
  if (!isDebugMode()) return;
  ensureDebugDrawer();
  enablePicker();
}

export async function refreshDebugDrawer(): Promise<void> {
  if (lastBoot) await updateDebugDrawer(lastBoot, lastWorkspace);
}

/** Refresh debug drawer with bootstrap + optional server ACL trace. */
export async function updateDebugDrawer(
  boot: SwcBootstrap,
  workspace?: DebugWorkspaceContext,
): Promise<void> {
  lastBoot = boot;
  lastWorkspace = workspace;
  if (!isDebugMode()) {
    document.getElementById("sum-debug-drawer")?.remove();
    return;
  }
  if (workspace?.model) {
    try {
      const url = `/web/debug/access?model=${encodeURIComponent(workspace.model)}`;
      const res = await fetch(url, { credentials: "same-origin", headers: { Accept: "application/json" } });
      if (res.ok) lastAclTrace = (await res.json()) as AclTrace;
      else lastAclTrace = null;
    } catch {
      lastAclTrace = null;
    }
  }
  const drawer = ensureDebugDrawer();
  const body = drawer.querySelector(".sum-debug-drawer__body") as HTMLDivElement;
  body.replaceChildren();

  const statusRows: Array<[string, string]> = [
    ["Mode", getDebugMode()],
    ["User", String(boot.user.id)],
    ["Company", String(boot.activeCompanyId)],
  ];
  body.append(section("Status", kvList(statusRows), "status"));

  if (workspace?.fields?.length) {
    const list = document.createElement("div");
    list.className = "sum-debug-field-list";
    for (const f of workspace.fields) {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "sum-debug-field-list__item";
      btn.textContent = f.name;
      btn.addEventListener("click", () => {
        const key = `${workspace.model}.${f.name}`;
        const el = document.querySelector<HTMLElement>(`[data-sum-debug-field="${key}"]`);
        if (el) selectDebugFieldFromElement(el);
        void refreshDebugDrawer();
      });
      list.append(btn);
    }
    body.append(section("Fields", list, "fields"));
  }

  if (workspace?.model) {
    const viewRows: Array<[string, string]> = [
      ["Model", workspace.model],
      ["Record", String(workspace.recordId)],
      ["View", workspace.viewType ?? "—"],
      ["Action", workspace.actionId != null ? String(workspace.actionId) : "—"],
      ["Menu", workspace.menuId ?? "—"],
    ];
    body.append(section("View", kvList(viewRows), "view"));
  } else {
    body.append(section("View", para("Arch and RPC logging enabled. Alt+click to inspect components."), "view"));
  }

  if (selectedFieldKey && selectedFieldEl) {
    const { fieldName } = parseDebugFieldKey(selectedFieldKey);
    const archField = workspace?.fields?.find((f) => f.name === fieldName);
    const record =
      workspace?.recordData && workspace.model
        ? new SwcRecord(workspace.model, workspace.recordId, workspace.recordData)
        : undefined;
    const info = buildFieldDebugInfo({
      fieldKey: selectedFieldKey,
      archField,
      domType: selectedFieldEl.getAttribute("data-sum-debug-type") ?? undefined,
      domWidget: selectedFieldEl.getAttribute("data-sum-debug-widget") ?? undefined,
      domLabel: selectedFieldEl.getAttribute("data-sum-debug-label") ?? undefined,
      record,
      userId: boot.user.id,
      companyId: boot.activeCompanyId,
      recordData: workspace?.recordData,
      aclField: lastAclTrace?.fields?.find((f) => f.name === fieldName),
      viewContext: { user_id: boot.user.id, company_id: boot.activeCompanyId },
    });
    body.append(section("Field", renderFieldDetailSection(info), "field"));
  } else if (isDebugMode()) {
    body.append(section("Field", para("Hover the info icon or click a field to inspect."), "field"));
  }

  if (workspace?.model) {
    const aclWrap = document.createElement("div");
    if (lastAclTrace) {
      aclWrap.append(
        kvList([["Groups", (lastAclTrace.group_ids ?? []).join(", ") || "—"]]),
      );
      const denied = (lastAclTrace.fields ?? []).filter((f) => f.read_denied || f.write_denied);
      if (denied.length > 0) {
        aclWrap.append(
          kvList([
            [
              "Denied fields",
              denied
                .map((f) => `${f.name}${f.read_denied ? " (read)" : ""}${f.write_denied ? " (write)" : ""}`)
                .join(", "),
            ],
          ]),
        );
      }
      if (lastAclTrace.rule_domain) {
        const details = document.createElement("details");
        details.className = "sum-debug-details";
        const sum = document.createElement("summary");
        sum.textContent = "Record rule domain";
        const pre = document.createElement("pre");
        pre.className = "sum-debug-trace";
        pre.textContent = lastAclTrace.rule_domain + (lastAclTrace.rule_domain_truncated ? " …" : "");
        details.append(sum, pre);
        aclWrap.append(details);
      }
    } else {
      aclWrap.append(para("ACL trace unavailable"));
    }
    body.append(section("ACL", aclWrap, "acl"));
  }

  const toolsBody = document.createElement("div");
  for (const item of listDebugMenuItems()) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "sum-debug-drawer__tool";
    btn.textContent = item.label;
    btn.addEventListener("click", () => item.run());
    toolsBody.append(btn);
  }
  body.append(section("Tools", toolsBody));
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
