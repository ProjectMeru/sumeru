import { describe, expect, it, vi } from "vitest";
import { initBreadcrumbNavigation, syncWorkspaceBreadcrumbs } from "../../src/shell/breadcrumb-sync.js";
import type { SwcBreadcrumb } from "../../src/types/workspace.js";

function mountBreadcrumb(): void {
  document.body.innerHTML = `
    <nav class="sum-breadcrumb" aria-label="Breadcrumb">
      <span class="sum-bc-module">CRM</span>
    </nav>
  `;
}

describe("breadcrumb-sync", () => {
  it("rewrites breadcrumb trail from workspace payload", () => {
    mountBreadcrumb();
    const crumbs: SwcBreadcrumb[] = [
      { label: "CRM", href: "/web?menu_id=1" },
      { label: "Sales", href: "/web?menu_id=2" },
      { label: "Pipeline", href: "" },
    ];
    syncWorkspaceBreadcrumbs(crumbs);

    const nav = document.querySelector("nav.sum-breadcrumb");
    expect(nav?.querySelectorAll(".sum-bc-sep").length).toBe(2);
    expect(nav?.querySelector(".sum-bc-current")?.textContent).toBe("Pipeline");
    expect(nav?.querySelectorAll("a.sum-bc-link").length).toBe(2);
  });

  it("routes breadcrumb workspace links through popstate", () => {
    mountBreadcrumb();
    syncWorkspaceBreadcrumbs([
      { label: "CRM", href: "/web?menu_id=1&action=9" },
      { label: "Pipeline", href: "" },
    ]);

    const onPop = vi.fn();
    window.addEventListener("popstate", onPop);
    initBreadcrumbNavigation();

    const link = document.querySelector("a.sum-bc-link") as HTMLAnchorElement;
    link.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));

    expect(window.location.pathname + window.location.search).toBe("/web?menu_id=1&action=9");
    expect(onPop).toHaveBeenCalledTimes(1);

    window.removeEventListener("popstate", onPop);
  });
});
