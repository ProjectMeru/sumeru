import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { WorkspaceRouter } from "../views/workspace/WorkspaceRouter.js";

export class ShellLayout extends SwcComponent {
  private workspaceRouter!: WorkspaceRouter;

  setup(): void {
    this.workspaceRouter = new WorkspaceRouter({}, this.env);
    this.workspaceRouter.setup?.();

    if (this.env.bootstrap.busEnabled) {
      this.env.services.bus.connect();
    }
  }

  private workspaceView(): HTMLElement {
    if (this.workspaceRouter.el?.isConnected) {
      this.workspaceRouter.patch();
      return this.workspaceRouter.el;
    }
    return this.workspaceRouter.render();
  }

  template() {
    return html`
      <div id="swc-root-inner">
        <main class="sum-workspace-inner">
          ${this.workspaceView()}
        </main>
      </div>
    `;
  }
}
