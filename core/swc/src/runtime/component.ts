import type { ComponentProps } from "../template/html.js";
import { patchKeyedChildren } from "./patch/keyed.js";
import { runWillPatch, runPatched } from "./lifecycle.js";
import { registerComponent, unregisterComponent } from "../devtools/bridge.js";

export type ComponentConstructor<P extends object = ComponentProps> = new (
  props: P,
  env: import("./env.js").SwcEnv,
) => SwcComponent<P>;

export abstract class SwcComponent<P extends object = ComponentProps> {
  readonly props: P;
  readonly env: import("./env.js").SwcEnv;
  el: HTMLElement | null = null;
  private mounted = false;

  constructor(props: P, env: import("./env.js").SwcEnv) {
    this.props = props;
    this.env = env;
  }

  setup?(): void;

  abstract template(): import("../template/html.js").TemplateResult;

  onMount?(): void;
  onWillUnmount?(): void;

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

  render(): HTMLElement {
    const result = this.template();
    const root = result.render();
    this.el = root;
    if (!this.mounted) {
      this.mounted = true;
      registerComponent(this);
      this.onMount?.();
    }
    return root;
  }

  /** Patch keyed tbody/list regions in-place when possible. */
  protected patchKeyedTbody(tbody: HTMLTableSectionElement, rows: Array<{ key: string; render: () => HTMLElement }>): boolean {
    if (!tbody) return false;
    patchKeyedChildren(tbody, rows);
    return true;
  }

  patch(): void {
    if (!this.el?.parentElement) return;
    runWillPatch();
    const parent = this.el.parentElement;
    const oldEl = this.el;
    const next = this.template().render();
    parent.replaceChild(next, oldEl);
    this.el = next;
    runPatched();
  }

  destroy(): void {
    this.onWillUnmount?.();
    unregisterComponent(this);
    this.el?.remove();
    this.el = null;
    this.mounted = false;
  }
}
