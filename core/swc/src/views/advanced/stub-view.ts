import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";

interface StubViewProps {
  payload: SwcWorkspacePayload;
}

export function renderStubView(title: string, payload: SwcWorkspacePayload) {
  const rows = payload.records ?? [];
  return html`
    <div class="sum-advanced-view">
      <h2>${title}</h2>
      <p class="sum-advanced-view-hint">${rows.length} record(s) loaded.</p>
      <ul>
        ${rows.slice(0, 20).map(
          (row) => html`<li>${String(row.name ?? row.display_name ?? row.id ?? "")}</li>`,
        )}
      </ul>
    </div>
  `;
}

function titleFallback(type: string): string {
  const trimmed = type.trim();
  if (!trimmed) return "View";
  return trimmed.charAt(0).toUpperCase() + trimmed.slice(1);
}

export class StubView extends SwcComponent<StubViewProps> {
  template() {
    const type = this.props.payload.arch.type ?? this.props.payload.viewType ?? "";
    return renderStubView(this.props.payload.arch.title ?? titleFallback(type), this.props.payload);
  }
}
