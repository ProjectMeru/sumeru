import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  buildLauncherIndex,
  filterLauncherItems,
  fuzzyScore,
  initAppLauncher,
  resetAppLauncherState,
} from "../../src/shell/app-launcher.js";
import type { SwcBootstrap } from "../../src/types/bootstrap.js";

function sampleBoot(): SwcBootstrap {
  return {
    csrfToken: "t",
    rpcUrl: "/api/rpc",
    apiDatabase: "test_db",
    swcApiBase: "/web/swc",
    user: { id: 1, name: "Admin", login: "admin", initials: "A" },
    company: { id: 1, name: "Main" },
    companies: [{ id: 1, name: "Main" }],
    activeCompanyId: 1,
    showCompanySwitcher: false,
    topMenus: [
      { id: "10", name: "Contacts", action: "/web?menu_id=10&action=42", module: "crm" },
    ],
    sidebarMenus: [
      {
        id: "crm",
        name: "CRM",
        sequence: 1,
        subMenus: [{ id: "11", name: "Leads", action: "/web?menu_id=11&action=43", module: "crm" }],
      },
    ],
    activeModuleId: "crm",
    activeMenuId: "10",
    apps: [
      {
        kind: "app",
        module: "crm",
        name: "CRM",
        action: "/web?menu_id=10",
        description: "Customer relationship",
      },
      {
        kind: "menu",
        module: "11",
        name: "Leads",
        action: "/web?menu_id=11&action=43",
        description: "crm",
      },
    ],
    pinnedApps: [],
    appsNavAllowed: true,
    settingsNavAllowed: true,
    activityEnabled: false,
    docsUrl: "/docs",
    profileUrl: "/web/settings",
  };
}

describe("app-launcher", () => {
  beforeEach(() => {
    resetAppLauncherState();
    document.body.innerHTML = "";
  });

  it("buildLauncherIndex deduplicates by action URL", () => {
    const items = buildLauncherIndex(sampleBoot());
    const actions = items.map((item) => item.action);
    expect(new Set(actions).size).toBe(actions.length);
    expect(items.some((item) => item.name === "CRM" && item.kind === "app")).toBe(true);
    expect(items.some((item) => item.name === "Leads" && item.kind === "menu")).toBe(true);
  });

  it("filterLauncherItems fuzzy-matches name and module", () => {
    const items = buildLauncherIndex(sampleBoot());
    const hits = filterLauncherItems(items, "cont");
    expect(hits.some((item) => item.name === "Contacts")).toBe(true);
    const crmHits = filterLauncherItems(items, "crm");
    expect(crmHits.length).toBeGreaterThan(0);
  });

  it("fuzzyScore prefers prefix matches", () => {
    expect(fuzzyScore("cr", "crm")).toBeGreaterThan(fuzzyScore("mx", "crm"));
  });

  it("initAppLauncher opens dialog on Cmd+K and renders results", () => {
    document.body.innerHTML = `
      <button id="sum-topbar-search-open"></button>
      <input id="sum-topbar-search-field" />
      <dialog id="sum-app-launcher">
        <input id="sum-app-launcher-input" />
        <ul id="sum-app-launcher-results"></ul>
      </dialog>
    `;

    const dlg = document.getElementById("sum-app-launcher") as HTMLDialogElement;
    dlg.showModal = () => {
      dlg.open = true;
    };
    dlg.close = () => {
      dlg.open = false;
    };

    const navigate = vi.fn();
    initAppLauncher(sampleBoot(), { navigate } as never);

    document.dispatchEvent(new KeyboardEvent("keydown", { key: "k", metaKey: true, bubbles: true }));
    expect(dlg.open).toBe(true);

    const input = document.getElementById("sum-app-launcher-input") as HTMLInputElement;
    input.value = "lead";
    input.dispatchEvent(new Event("input", { bubbles: true }));

    const rows = document.querySelectorAll(".sum-app-launcher-result");
    expect(rows.length).toBeGreaterThan(0);
    expect(document.querySelector(".sum-app-launcher-result-kind")).toBeTruthy();
  });

  it("arrow keys move selection without mouseenter resetting it", () => {
    document.body.innerHTML = `
      <button id="sum-topbar-search-open"></button>
      <input id="sum-topbar-search-field" />
      <dialog id="sum-app-launcher">
        <input id="sum-app-launcher-input" type="search" />
        <ul id="sum-app-launcher-results"></ul>
      </dialog>
    `;

    const dlg = document.getElementById("sum-app-launcher") as HTMLDialogElement;
    dlg.showModal = () => {
      dlg.open = true;
    };
    dlg.close = () => {
      dlg.open = false;
    };

    const boot: SwcBootstrap = {
      ...sampleBoot(),
      apps: [
        { kind: "app", module: "a", name: "Alpha", action: "/a" },
        { kind: "app", module: "b", name: "Beta", action: "/b" },
        { kind: "app", module: "c", name: "Gamma", action: "/c" },
      ],
      topMenus: [],
      sidebarMenus: [],
    };

    initAppLauncher(boot, { navigate: vi.fn() } as never);

    document.dispatchEvent(new KeyboardEvent("keydown", { key: "k", metaKey: true, bubbles: true }));

    const input = document.getElementById("sum-app-launcher-input") as HTMLInputElement;
    input.focus();

    const results = document.getElementById("sum-app-launcher-results") as HTMLUListElement;
    const firstRow = results.querySelector(".sum-app-launcher-result") as HTMLElement;

    firstRow.dispatchEvent(new MouseEvent("mouseenter", { bubbles: true }));
    firstRow.dispatchEvent(new MouseEvent("mousemove", { bubbles: true }));
    results.dispatchEvent(new MouseEvent("mousemove", { bubbles: true }));

    input.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowDown", bubbles: true }));
    let active = results.querySelector(".sum-app-launcher-result.is-active");
    expect(active?.querySelector(".sum-app-launcher-result-name")?.textContent).toBe("Beta");

    input.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowDown", bubbles: true }));
    active = results.querySelector(".sum-app-launcher-result.is-active");
    expect(active?.querySelector(".sum-app-launcher-result-name")?.textContent).toBe("Gamma");

    input.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowUp", bubbles: true }));
    active = results.querySelector(".sum-app-launcher-result.is-active");
    expect(active?.querySelector(".sum-app-launcher-result-name")?.textContent).toBe("Beta");

    input.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowDown", bubbles: true }));
    active = results.querySelector(".sum-app-launcher-result.is-active");
    expect(active?.querySelector(".sum-app-launcher-result-name")?.textContent).toBe("Gamma");

    input.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowDown", bubbles: true }));
    active = results.querySelector(".sum-app-launcher-result.is-active");
    expect(active?.querySelector(".sum-app-launcher-result-name")?.textContent).toBe("Gamma");

    input.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowUp", bubbles: true }));
    input.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowUp", bubbles: true }));
    active = results.querySelector(".sum-app-launcher-result.is-active");
    expect(active?.querySelector(".sum-app-launcher-result-name")?.textContent).toBe("Alpha");
  });
});
