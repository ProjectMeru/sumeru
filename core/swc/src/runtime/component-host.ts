import type { SwcEnv } from "./env.js";
import type { SwcComponent, ComponentConstructor } from "./component.js";
import type { TemplateResult } from "../template/html.js";
import { runWillStart } from "./lifecycle.js";

function fillSlots(root: HTMLElement, slots: Record<string, TemplateResult>): void {
  for (const [name, result] of Object.entries(slots)) {
    const target = root.querySelector(`[data-slot="${name}"]`);
    if (!(target instanceof HTMLElement)) continue;
    if (target.firstElementChild) {
      const next = result.patch(target.firstElementChild as HTMLElement);
      if (next !== target.firstElementChild) target.replaceChildren(next);
    } else {
      target.replaceChildren(result.render());
    }
  }
}

/** Mount a nested SwcComponent inside a parent template. */
export class ComponentHost<P extends object = Record<string, unknown>> implements TemplateResult {
  private instance: SwcComponent<P> | null = null;
  private rootElement: HTMLElement | null = null;

  constructor(
    private readonly Component: ComponentConstructor<P>,
    private props: P,
    private readonly env: SwcEnv,
    private readonly slots: Record<string, TemplateResult> = {},
  ) {}

  updateProps(next: P): void {
    this.props = next;
    this.instance?.updateProps(next);
  }

  render(): HTMLElement {
    if (!this.instance) {
      this.instance = new this.Component(this.props, this.env);
      this.instance.callSetup();
      void runWillStart(this.instance).then(() => {
        if (this.instance?.rootElement?.isConnected) this.instance.patch();
      });
      this.rootElement = this.instance.render();
      fillSlots(this.rootElement, this.slots);
      return this.rootElement;
    }
    this.instance.updateProps(this.props);
    this.rootElement = this.instance.rootElement ?? this.instance.render();
    fillSlots(this.rootElement, this.slots);
    return this.rootElement;
  }

  patch(existing: HTMLElement): HTMLElement {
    this.rootElement = existing;
    if (!this.instance) {
      const root = this.render();
      if (root !== existing) existing.replaceWith(root);
      return root;
    }
    this.instance.updateProps(this.props);
    const root = this.instance.rootElement ?? existing;
    fillSlots(root, this.slots);
    this.rootElement = root;
    return root;
  }

  destroy(): void {
    if (this.instance) {
      this.instance.destroy();
      this.instance = null;
      this.rootElement = null;
    }
  }
}

export function mountComponent<P extends object>(
  Component: ComponentConstructor<P>,
  props: P,
  env: SwcEnv,
  slots: Record<string, TemplateResult> = {},
): ComponentHost<P> {
  return new ComponentHost(Component, props, env, slots);
}
