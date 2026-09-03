import { SwcComponent } from "../../runtime/component.js";
import type { TemplateResult, TemplateValue } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { CollectionBarHost, mountCollectionBar } from "./collection-bar-host.js";
import { renderCollectionShell, type CollectionShellOptions } from "./collection-layout.js";

export type CollectionViewProps = { payload: SwcWorkspacePayload };

/** Base class for list/kanban/graph and other collection workspace views with a control bar. */
export abstract class CollectionView<P extends CollectionViewProps = CollectionViewProps> extends SwcComponent<P> {
  protected collectionBar!: CollectionBarHost;
  protected abstract readonly collectionViewType: string;

  protected collectionBarExtra?(): TemplateValue | undefined {
    return undefined;
  }

  override setup(): void {
    this.collectionBar = mountCollectionBar(
      this.props.payload,
      this.collectionViewType,
      this.env,
      this.collectionBarExtra?.(),
    );
    this.onCollectionSetup?.();
  }

  override onPropsChanged(props: P): void {
    this.collectionBar.updateProps({
      payload: props.payload,
      viewType: this.collectionViewType,
      extraPrimary: this.collectionBarExtra?.(),
    });
    this.onCollectionPropsChanged?.(props);
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
    this.onCollectionTeardown?.();
  }

  protected onCollectionSetup?(): void;
  protected onCollectionPropsChanged?(props: P): void;
  protected onCollectionTeardown?(): void;

  protected renderShell(body: TemplateValue, options?: CollectionShellOptions): TemplateResult {
    return renderCollectionShell(this.collectionViewType, this.collectionBar, body, options);
  }
}
