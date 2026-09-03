import { describe, it, expect, afterEach, vi } from "vitest";
import {
  stripAppsFlashQueryParam,
  handleAppsListRowClick,
  initAppsPage,
} from "../../src/apps/apps-page.js";

describe("stripAppsFlashQueryParam", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("removes msg from the URL via replaceState", () => {
    const replaceState = vi.fn();
    vi.stubGlobal("history", { replaceState });
    stripAppsFlashQueryParam({ href: "https://example.com/web/apps?module=sale&msg=installed_sale&layout=grid" });
    expect(replaceState).toHaveBeenCalledWith(
      {},
      "",
      "/web/apps?module=sale&layout=grid",
    );
  });

  it("no-ops when msg is absent", () => {
    const replaceState = vi.fn();
    vi.stubGlobal("history", { replaceState });
    stripAppsFlashQueryParam({ href: "https://example.com/web/apps?module=sale" });
    expect(replaceState).not.toHaveBeenCalled();
  });
});

describe("handleAppsListRowClick", () => {
  it("navigates when row is clicked outside actions", () => {
    const row = document.createElement("tr");
    row.dataset.detailUrl = "/web/apps?module=sale";
    const cell = document.createElement("td");
    cell.textContent = "Sales";
    row.appendChild(cell);
    document.body.appendChild(row);

    const assign = vi.fn();
    vi.stubGlobal("location", { href: "", assign });

    handleAppsListRowClick({ currentTarget: row, target: cell } as unknown as MouseEvent);
    expect(location.href).toBe("/web/apps?module=sale");

    document.body.removeChild(row);
  });

  it("ignores clicks on action controls", () => {
    const row = document.createElement("tr");
    row.dataset.detailUrl = "/web/apps?module=sale";
    const actions = document.createElement("td");
    actions.className = "sum-apps-list-actions";
    const button = document.createElement("button");
    button.type = "button";
    actions.appendChild(button);
    row.appendChild(actions);
    document.body.appendChild(row);

    vi.stubGlobal("location", { href: "https://example.com/web/apps" });

    handleAppsListRowClick({ currentTarget: row, target: button } as unknown as MouseEvent);
    expect(location.href).toBe("https://example.com/web/apps");

    document.body.removeChild(row);
  });
});

describe("initAppsPage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    document.body.innerHTML = "";
  });

  it("auto-submits control bar when a select changes", () => {
    const form = document.createElement("form");
    form.className = "sum-apps-control-bar";
    const select = document.createElement("select");
    select.setAttribute("data-auto-submit", "");
    select.name = "filter";
    form.appendChild(select);
    document.body.appendChild(form);

    const requestSubmit = vi.fn();
    form.requestSubmit = requestSubmit;
    vi.stubGlobal("history", { replaceState: vi.fn() });

    initAppsPage(document);
    select.dispatchEvent(new Event("change", { bubbles: true }));
    expect(requestSubmit).toHaveBeenCalled();
  });
});
