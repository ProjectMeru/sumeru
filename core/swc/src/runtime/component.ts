import type { ComponentProps } from "../template/html.js";
import { runWillPatch, runPatched } from "./lifecycle.js";
import { registerComponent, unregisterComponent } from "../devtools/bridge.js";

export type ComponentConstructor<P extends object = ComponentProps> = new (
  props: P,
  env: import("./env.js").SwcEnv,
) => SwcComponent<P>;

export abstract class SwcComponent<P extends object = ComponentProps> {
  readonly props: P;
  readonly env: import("./env.js").SwcEnv;
  rootElement: HTMLElement | null = null;
  private mounted = false;

  constructor(props: P, env: import("./env.js").SwcEnv) {
    this.props = props;
    this.env = env;
  }

  setup?(): void;

  abstract template(): import("../template/html.js").TemplateResult;

  onMount?(): void;
  onWillUnmount?(): void;

  /** Called after the root node is replaced in the DOM. */
  afterPatch?(): void;

  /** Called when props are updated on an existing instance (SPA navigation). */
  onPropsChanged(_props: P): void {
    // override in subclasses
  }

  /** Replace props and re-render without recreating the component instance. */
  updateProps(next: P): void {
    (this as { props: P }).props = next;
    this.onPropsChanged(next);
    this.patch();
  }

  /** Re-render this component if it is still in the document. */
  rerender(): void {
    if (this.rootElement?.isConnected) this.patch();
  }

  /** Patch in place when mounted; otherwise produce a new root. */
  renderOrPatch(): HTMLElement {
    if (this.rootElement?.isConnected) {
      this.patch();
      return this.rootElement;
    }
    return this.render();
  }

  render(): HTMLElement {
    const result = this.template();
    const root = result.render();
    this.rootElement = root;
    if (!this.mounted) {
      this.mounted = true;
      registerComponent(this);
      this.onMount?.();
    }
    return root;
  }

  patch(): void {
    if (!this.rootElement?.parentElement) return;
    runWillPatch();
    const parent = this.rootElement.parentElement;
    const previousRoot = this.rootElement;
    const next = this.template().render();
    parent.replaceChild(next, previousRoot);
    this.rootElement = next;
    runPatched();
    this.afterPatch?.();
  }

  destroy(): void {
    this.onWillUnmount?.();
    unregisterComponent(this);
    this.rootElement?.remove();
    this.rootElement = null;
    this.mounted = false;
  }
}
