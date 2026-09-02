import type { SwcBreadcrumb } from "../types/workspace.js";
import { WEB_ROUTE } from "../constants/routes.js";

/** Keep server-rendered breadcrumb trail in sync after SPA navigation. */
export function syncWorkspaceBreadcrumbs(breadcrumbs: SwcBreadcrumb[]): void {
  const nav = document.querySelector<HTMLElement>("nav.sum-breadcrumb");
  if (!nav || breadcrumbs.length === 0) return;

  nav.replaceChildren();
  breadcrumbs.forEach((crumb, index) => {
    if (index > 0) {
      const sep = document.createElement("span");
      sep.className = "sum-bc-sep";
      sep.setAttribute("aria-hidden", "true");
      sep.textContent = "/";
      nav.appendChild(sep);
    }
    if (crumb.href) {
      const link = document.createElement("a");
      link.href = crumb.href;
      link.className = "sum-bc-link";
      link.textContent = crumb.label;
      nav.appendChild(link);
    } else {
      const current = document.createElement("span");
      current.className = "sum-bc-current";
      current.setAttribute("aria-current", "page");
      current.textContent = crumb.label;
      nav.appendChild(current);
    }
  });
}

/** Route breadcrumb workspace links through the SPA instead of full page loads. */
export function initBreadcrumbNavigation(): void {
  document.addEventListener("click", (ev) => {
    const link = (ev.target as Element).closest<HTMLAnchorElement>("nav.sum-breadcrumb a.sum-bc-link[href]");
    if (!link?.href.includes(`${WEB_ROUTE}?`)) return;

    ev.preventDefault();
    const url = new URL(link.href, window.location.origin);
    window.history.pushState({}, "", `${url.pathname}${url.search}`);
    window.dispatchEvent(new PopStateEvent("popstate"));
  });
}
