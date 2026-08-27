import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcBootstrap } from "../types/bootstrap.js";

interface ShellPageProps {
  boot: SwcBootstrap;
  page: "home" | "apps" | "settings";
}

function appHref(action: string): string {
  const id = action.replace(/\D/g, "") || action;
  return `/web?action=${encodeURIComponent(id)}`;
}

export class ShellPageView extends SwcComponent<ShellPageProps> {
  template() {
    const { boot, page } = this.props;
    if (page === "apps") {
      return html`<div class="sum-shell-page sum-shell-apps">
        <h1>Applications</h1>
        <div class="sum-shell-app-grid">
          ${boot.apps.map(
            (app) => html`<a class="sum-shell-app-tile" href=${appHref(app.action)}>
              <span class="sum-shell-app-name">${app.name}</span>
            </a>`,
          )}
        </div>
      </div>`;
    }
    if (page === "settings") {
      return html`<div class="sum-shell-page sum-shell-settings">
        <h1>Settings</h1>
        <p>Company and user preferences (SPA shell route).</p>
      </div>`;
    }
    return html`<div class="sum-shell-page sum-shell-home">
      <h1>Home</h1>
      <p>Welcome, ${boot.user.name}.</p>
      <div class="sum-shell-app-grid">
        ${boot.apps.slice(0, 6).map(
          (app) => html`<a class="sum-shell-app-tile" href=${appHref(app.action)}>${app.name}</a>`,
        )}
      </div>
    </div>`;
  }
}
