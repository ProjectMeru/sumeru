import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";

/** Embeds a sys.action.url target inside workspace chrome. */
export class IframeView extends SwcComponent {
  private src = "";

  override setup(): void {
    this.syncSrc(this.props.payload as SwcWorkspacePayload);
  }

  override onPropsChanged(): void {
    this.syncSrc(this.props.payload as SwcWorkspacePayload);
  }

  private syncSrc(payload: SwcWorkspacePayload): void {
    this.src = (payload.iframeUrl ?? "").trim();
  }

  override template() {
    if (!this.src) {
      return html`<div class="sum-flash sum-flash--error">Missing iframe URL</div>`;
    }
    return html`
      <div class="sum-iframe-view">
        <iframe class="sum-iframe-view__frame" src="${this.src}" title="Embedded content"></iframe>
      </div>
    `;
  }
}
