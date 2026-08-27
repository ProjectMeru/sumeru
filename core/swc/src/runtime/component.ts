import type { ComponentProps } from "../template/html.js";
import { runWillPatch, runPatched } from "./lifecycle.js";
import {
  runMountEffects,
  runUnmountEffects,
  withActiveHost,
  type EffectFn,
  type StartFn,
} from "./hooks.js";
import { scheduleRender } from "./scheduler.js";
import { registerComponent, unregisterComponent } from "../devtools/bridge.js";
import { applyPortals, restorePortals } from "./portals.js";

export type ComponentConstructor<P extends object = ComponentProps> = new (
  props: P,
  env: import("./env.js").SwcEnv,
) => SwcComponent<P>;

export abstract class SwcComponent<P extends object = ComponentProps> {
  readonly props: P;
  readonly env: import("./env.js").SwcEnv;
  rootElement: HTMLElement | null = null;
  private mounted = false;

  readonly hookState: unknown[] = [];
  hookIndex = 0;
  readonly willStart: StartFn[] = [];
  readonly willPatch: Array<() => void> = [];
  readonly patched: Array<() => void> = [];
  readonly mountEffects: EffectFn[] = [];
  readonly unmountEffects: EffectFn[] = [];

  constructor(props: P, env: import("./env.js").SwcEnv) {
    this.props = props;
    this.env = env;
  }

  consumeHookSlot(): number {
    const index = this.hookIndex;
    this.hookIndex += 1;
    return index;
  }

  /** Run `setup` with this instance as the active hook host. */
  callSetup(): void {
    this.hookIndex = 0;
    withActiveHost(this, () => {
      this.setup?.();
    });
  }

  setup?(): void;

  abstract template(): import("../template/html.js").TemplateResult;

  onMount?(): void;
  onWillUnmount?(): void;

  /** Called after the root node is patched or replaced. */
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

  /** Queue a patch if this component is still in the document. */
  rerender(): void {
    if (this.rootElement?.isConnected) scheduleRender(this);
  }

  /** Patch in place when a root already exists; otherwise produce a new root. */
  renderOrPatch(): HTMLElement {
    if (this.rootElement) {
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
      applyPortals(root);
      runMountEffects(this);
      this.onMount?.();
    }
    return root;
  }

  patch(): void {
    if (!this.rootElement) return;
    runWillPatch(this);
    const previousRoot = this.rootElement;
    const result = this.template();
    const next = result.patch(previousRoot);
    if (next !== previousRoot && previousRoot.parentNode) {
      previousRoot.replaceWith(next);
    }
    this.rootElement = next;
    applyPortals(next);
    runPatched(this);
    this.afterPatch?.();
  }

  destroy(): void {
    runUnmountEffects(this);
    this.onWillUnmount?.();
    restorePortals(this.rootElement);
    unregisterComponent(this);
    this.rootElement?.remove();
    this.rootElement = null;
    this.mounted = false;
  }
}
