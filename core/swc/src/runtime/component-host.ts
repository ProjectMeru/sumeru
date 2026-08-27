import type { SwcEnv } from "./env.js";
import type { SwcComponent, ComponentConstructor } from "./component.js";
import type { TemplateResult } from "../template/html.js";

/** Mount a nested SwcComponent inside a parent template. */
export class ComponentHost<P extends object = Record<string, unknown>> implements TemplateResult {
  private instance: SwcComponent<P> | null = null;
  private rootElement: HTMLElement | null = null;

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
      this.rootElement = this.instance.render();
      return this.rootElement;
    }
    this.instance.updateProps(this.props);
    return this.rootElement ?? this.instance.render();
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
): ComponentHost<P> {
  return new ComponentHost(Component, props, env);
}
