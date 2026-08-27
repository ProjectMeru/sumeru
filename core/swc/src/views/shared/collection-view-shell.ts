import type { SwcComponent } from "../../runtime/component.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { CollectionBarHost, mountCollectionBar } from "./collection-bar-host.js";

/** Attach collection control bar lifecycle to a collection view. */
export function attachCollectionBar(
  view: SwcComponent<{ payload: SwcWorkspacePayload }>,
  viewType: string,
): CollectionBarHost {
  const bar = mountCollectionBar(view.props.payload, viewType, view.env);
  return bar;
}

export function syncCollectionBar(bar: CollectionBarHost, payload: SwcWorkspacePayload, viewType: string): void {
  bar.updateProps({ payload, viewType });
}

export function destroyCollectionBar(bar: CollectionBarHost | undefined): void {
  bar?.destroy();
}
