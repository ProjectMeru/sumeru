import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { WorkspaceRouter } from "../views/workspace/WorkspaceRouter.js";

export class ShellLayout extends SwcComponent {
  private workspaceRouter!: WorkspaceRouter;

  override setup(): void {
    this.workspaceRouter = new WorkspaceRouter({}, this.env);
    this.workspaceRouter.callSetup();

    if (this.env.bootstrap.busEnabled) {
      this.env.services.bus.connect();
    }
  }

  private workspaceView(): HTMLElement {
    return this.workspaceRouter.renderOrPatch();
  }

  override template() {
    return html`
      <div id="swc-root-inner">
        <main class="sum-workspace-inner">
          ${this.workspaceView()}
        </main>
      </div>
    `;
  }
}
