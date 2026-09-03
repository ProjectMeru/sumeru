import { html, type TemplateResult, type TemplateValue } from "../../template/html.js";
import type { CollectionBarHost } from "./collection-bar-host.js";

export interface CollectionShellOptions {
  /** Inserted after the control bar and before body (e.g. list pagination). */
  extraBeforeBody?: TemplateValue;
  /** Extra root class names on the collection wrapper. */
  rootClass?: string;
}

/** Standard workspace collection wrapper: control bar + optional chrome + view body. */
export function renderCollectionShell(
  viewType: string,
  collectionBar: CollectionBarHost,
  body: TemplateValue,
  options: CollectionShellOptions = {},
): TemplateResult {
  const rootClass = ["sum-collection-view", `sum-${viewType}-view`, options.rootClass]
    .filter(Boolean)
    .join(" ");
  return html`
    <div class=${rootClass}>
      ${collectionBar.renderOrPatch()}
      ${options.extraBeforeBody ?? ""}
      ${body}
    </div>
  `;
}
