import type { ActionService } from "../services/action.js";
import type { CommandService } from "../services/command.js";
import type { SwcBootstrap, SwcBootstrapApp } from "../types/bootstrap.js";

export interface LauncherItem {
  kind: "app" | "menu" | "command";
  name: string;
  module: string;
  action: string;
  description?: string;
}

export function buildLauncherIndex(boot: SwcBootstrap): LauncherItem[] {
  const seen = new Set<string>();
  const out: LauncherItem[] = [];

  const add = (item: LauncherItem): void => {
    const key = item.action.trim();
    if (!key || seen.has(key)) return;
    seen.add(key);
    out.push(item);
  };

  for (const app of boot.apps ?? []) {
    add(appToLauncherItem(app));
  }

  for (const menu of boot.topMenus ?? []) {
    add({
      kind: "menu",
      name: menu.name,
      module: menu.module?.trim() || "Menu",
      action: menu.action,
    });
  }

  for (const group of boot.sidebarMenus ?? []) {
    for (const menu of group.subMenus ?? []) {
      add({
        kind: "menu",
        name: menu.name,
        module: menu.module?.trim() || group.name,
        action: menu.action,
      });
    }
  }

  return out;
}

function appToLauncherItem(app: SwcBootstrapApp): LauncherItem {
  return {
    kind: app.kind === "menu" ? "menu" : "app",
    name: app.name,
    module: app.module,
    action: app.action,
    description: app.description,
  };
}

export function fuzzyScore(query: string, text: string): number {
  const q = query.trim().toLowerCase();
  const t = text.trim().toLowerCase();
  if (!q) return 1;
  if (!t) return 0;
  if (t === q) return 100;
  if (t.startsWith(q)) return 80;
  if (t.includes(q)) return 60;

  let qi = 0;
  for (let i = 0; i < t.length && qi < q.length; i++) {
    if (t[i] === q[qi]) qi++;
  }
  return qi === q.length ? 40 : 0;
}

export function scoreLauncherItem(query: string, item: LauncherItem): number {
  const q = query.trim();
  if (!q) return 1;
  const fields = [item.name, item.module, item.description ?? "", item.kind];
  return Math.max(...fields.map((field) => fuzzyScore(q, field)));
}

export function filterLauncherItems(items: LauncherItem[], query: string): LauncherItem[] {
  const q = query.trim();
  if (!q) return items;
  return items
    .map((item) => ({ item, score: scoreLauncherItem(q, item) }))
    .filter(({ score }) => score > 0)
    .sort((a, b) => b.score - a.score || a.item.name.localeCompare(b.item.name))
    .map(({ item }) => item);
}

let initialized = false;

export function filterCommands(command: CommandService, query: string): { id: string; label: string }[] {
  const q = query.trim();
  const list = command.list();
  if (!q) return list.map((c) => ({ id: c.id, label: c.label }));
  return list
    .map((c) => ({ c, score: fuzzyScore(q, c.label) }))
    .filter(({ score }) => score > 0)
    .sort((a, b) => b.score - a.score || a.c.label.localeCompare(b.c.label))
    .map(({ c }) => ({ id: c.id, label: c.label }));
}

export function initAppLauncher(
  boot: SwcBootstrap,
  action: ActionService,
  command?: CommandService,
): void {
  if (initialized) return;
  initialized = true;

  const dlg = document.getElementById("sum-app-launcher") as HTMLDialogElement | null;
  const input = document.getElementById("sum-app-launcher-input") as HTMLInputElement | null;
  const results = document.getElementById("sum-app-launcher-results") as HTMLUListElement | null;
  const searchBtn = document.getElementById("sum-topbar-search-open");
  const searchField = document.getElementById("sum-topbar-search-field");

  if (!dlg || !input || !results) return;

  const items = buildLauncherIndex(boot);
  let query = "";
  let activeIndex = 0;
  let open = false;
  let pointerNav = false;
  let pane: "apps" | "commands" = "apps";

  const filtered = (): LauncherItem[] => {
    if (pane === "commands" && command) {
      return filterCommands(command, query).map((c) => ({
        kind: "command" as const,
        name: c.label,
        module: "Command",
        action: c.id,
      }));
    }
    return filterLauncherItems(items, query);
  };

  const scrollActiveIntoView = (): void => {
    const active = results.querySelector(".sum-app-launcher-result.is-active");
    if (active && typeof (active as HTMLElement).scrollIntoView === "function") {
      active.scrollIntoView({ block: "nearest" });
    }
  };

  const syncActiveRow = (): void => {
    const rows = results.querySelectorAll<HTMLElement>(".sum-app-launcher-result");
    rows.forEach((row, index) => {
      const selected = index === activeIndex;
      row.classList.toggle("is-active", selected);
      row.setAttribute("aria-selected", selected ? "true" : "false");
    });
    scrollActiveIntoView();
  };

  const setActiveIndex = (index: number): void => {
    const list = filtered();
    if (list.length === 0) {
      activeIndex = 0;
      return;
    }
    activeIndex = Math.max(0, Math.min(index, list.length - 1));
    syncActiveRow();
  };

  const close = (): void => {
    if (!open) return;
    open = false;
    pointerNav = false;
    query = "";
    activeIndex = 0;
    input.value = "";
    results.replaceChildren();
    if (dlg.open) dlg.close();
  };

  const renderResults = (): void => {
    const list = filtered();
    if (activeIndex >= list.length) activeIndex = Math.max(0, list.length - 1);
    results.replaceChildren();

    list.forEach((item, index) => {
      const row = document.createElement("li");
      row.className = "sum-app-launcher-result";
      if (index === activeIndex) row.classList.add("is-active");
      row.setAttribute("role", "option");
      row.setAttribute("aria-selected", index === activeIndex ? "true" : "false");

      const letter = document.createElement("span");
      letter.className = "sum-app-launcher-result-letter";
      letter.textContent = (item.name.trim()[0] ?? "?").toUpperCase();

      const body = document.createElement("span");
      body.className = "sum-app-launcher-result-body";

      const nameRow = document.createElement("span");
      nameRow.className = "sum-app-launcher-result-name-row";

      const name = document.createElement("span");
      name.className = "sum-app-launcher-result-name";
      name.textContent = item.name;

      const kind = document.createElement("span");
      kind.className = `sum-app-launcher-result-kind sum-app-launcher-result-kind--${item.kind}`;
      kind.textContent = item.kind === "app" ? "App" : item.kind === "command" ? "Cmd" : "Menu";

      nameRow.append(name, kind);

      const meta = document.createElement("span");
      meta.className = "sum-app-launcher-result-meta";
      meta.textContent = item.description?.trim() || item.module;

      body.append(nameRow, meta);
      row.append(letter, body);

      row.addEventListener("mouseenter", () => {
        if (!pointerNav) return;
        setActiveIndex(index);
      });
      row.addEventListener("click", () => {
        if (item.kind === "command" && command) {
          command.run(item.action);
        } else {
          action.navigate(item.action);
        }
        close();
      });
      results.appendChild(row);
    });
  };

  const openLauncher = (): void => {
    if (open) return;
    open = true;
    pointerNav = false;
    query = "";
    activeIndex = 0;
    pane = "apps";
    input.value = "";
    ensurePaneTabs();
    renderResults();
    if (!dlg.open) dlg.showModal();
    queueMicrotask(() => input.focus());
  };

  const toggle = (): void => {
    if (open) close();
    else openLauncher();
  };

  const activate = (): void => {
    const list = filtered();
    const item = list[activeIndex];
    if (!item) return;
    if (item.kind === "command" && command) {
      command.run(item.action);
    } else {
      action.navigate(item.action);
    }
    close();
  };

  const onInputKeydown = (ev: KeyboardEvent): void => {
    if (!open) return;
    const list = filtered();

    if (ev.key === "ArrowDown") {
      ev.preventDefault();
      pointerNav = false;
      if (list.length === 0) return;
      setActiveIndex(activeIndex + 1);
      return;
    }

    if (ev.key === "ArrowUp") {
      ev.preventDefault();
      pointerNav = false;
      if (list.length === 0) return;
      setActiveIndex(activeIndex - 1);
      return;
    }

    if (ev.key === "Enter") {
      ev.preventDefault();
      activate();
      return;
    }

    if (ev.key === "Escape") {
      ev.preventDefault();
      close();
    }
  };

  const onGlobalKeydown = (ev: KeyboardEvent): void => {
    const target = ev.target;
    if (
      target instanceof HTMLElement &&
      (target.isContentEditable ||
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.tagName === "SELECT")
    ) {
      return;
    }
    if ((ev.ctrlKey || ev.metaKey) && ev.key.toLowerCase() === "k") {
      ev.preventDefault();
      toggle();
    }
  };

  const ensurePaneTabs = (): void => {
    if (!command || dlg.querySelector(".sum-app-launcher-panes")) return;
    const panes = document.createElement("div");
    panes.className = "sum-app-launcher-panes";
    const appsBtn = document.createElement("button");
    appsBtn.type = "button";
    appsBtn.textContent = "Apps";
    const cmdBtn = document.createElement("button");
    cmdBtn.type = "button";
    cmdBtn.textContent = "Commands";
    const syncPane = (): void => {
      appsBtn.classList.toggle("is-active", pane === "apps");
      cmdBtn.classList.toggle("is-active", pane === "commands");
      input.placeholder = pane === "commands" ? "Search commands…" : "Search apps and menus…";
    };
    appsBtn.addEventListener("click", () => {
      pane = "apps";
      activeIndex = 0;
      syncPane();
      renderResults();
    });
    cmdBtn.addEventListener("click", () => {
      pane = "commands";
      activeIndex = 0;
      syncPane();
      renderResults();
    });
    panes.append(appsBtn, cmdBtn);
    dlg.insertBefore(panes, dlg.firstChild);
    syncPane();
  };

  input.addEventListener("input", () => {
    query = input.value;
    activeIndex = 0;
    pointerNav = false;
    renderResults();
  });

  results.addEventListener("mousemove", () => {
    pointerNav = true;
  });

  dlg.addEventListener("close", () => {
    open = false;
    pointerNav = false;
    query = "";
    activeIndex = 0;
    input.value = "";
    results.replaceChildren();
  });

  input.addEventListener("keydown", onInputKeydown);
  document.addEventListener("keydown", onGlobalKeydown);
  searchBtn?.addEventListener("click", toggle);
  searchField?.addEventListener("click", toggle);
}

/** Test helper — reset singleton guard between tests. */
export function resetAppLauncherState(): void {
  initialized = false;
}
