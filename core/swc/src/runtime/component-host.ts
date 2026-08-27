import type { SwcEnv } from "./env.js";
import type { SwcComponent, ComponentConstructor } from "./component.js";
import type { TemplateResult } from "../template/html.js";
import { registerComponent, unregisterComponent } from "../devtools/bridge.js";

/** Mount a nested SwcComponent inside a parent template. */
export class ComponentHost<P extends object = Record<string, unknown>> implements TemplateResult {
  private instance: SwcComponent<P> | null = null;
  private el: HTMLElement | null = null;

  constructor(
    private readonly Component: ComponentConstructor<P>,
    private props: P,
    private readonly env: SwcEnv,
  ) {}

  updateProps(next: P): void {
    this.props = next;
    this.instance?.updateProps(next);
  }

  render(): HTMLElement {
    if (!this.instance) {
      this.instance = new this.Component(this.props, this.env);
      this.instance.setup?.();
      registerComponent(this.instance);
      this.el = this.instance.render();
      this.instance.onMount?.();
      return this.el;
    }
    this.instance.updateProps(this.props);
    return this.el ?? this.instance.render();
  }

  destroy(): void {
    if (this.instance) {
      unregisterComponent(this.instance);
      this.instance.destroy();
      this.instance = null;
      this.el = null;
    }
  }
}

export function mountComponent<P extends object>(
  Component: ComponentConstructor<P>,
  props: P,
  env: SwcEnv,
): ComponentHost<P> {
  return new ComponentHost(Component, props, env);
}
