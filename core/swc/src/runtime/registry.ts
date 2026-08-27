import type { SwcComponent } from "./component.js";
import type { SwcEnv } from "./env.js";
import type { SwcWorkspacePayload } from "../types/workspace.js";
import type { FieldWidgetProps } from "../widgets/field-props.js";

export type RegistryCategory = "fields" | "views" | "services" | "main_components";

export type ViewConstructor = new (
  props: { payload: SwcWorkspacePayload },
  env: SwcEnv,
) => SwcComponent<{ payload: SwcWorkspacePayload }>;

export type FieldWidgetConstructor = new (props: FieldWidgetProps, env: SwcEnv) => SwcComponent<FieldWidgetProps>;

export type ServiceInstance = object;

export type MainComponentConstructor = new (props: object, env: SwcEnv) => SwcComponent<object>;

function debugDuplicatesEnabled(): boolean {
  return typeof location !== "undefined" && /(?:^|[?&])debug=1(?:&|$)/.test(location.search);
}

export class CategoryRegistry<T> {
  constructor(private readonly map: Map<string, T>) {}

  add(key: string, value: T): void {
    if (this.map.has(key) && debugDuplicatesEnabled()) {
      throw new Error(`Registry already has "${key}"`);
    }
    this.map.set(key, value);
  }

  get(key: string): T | undefined {
    return this.map.get(key);
  }

  keys(): string[] {
    return [...this.map.keys()];
  }
}

export class Registry {
  private readonly maps = {
    fields: new Map<string, FieldWidgetConstructor>(),
    views: new Map<string, ViewConstructor>(),
    services: new Map<string, ServiceInstance>(),
    main_components: new Map<string, MainComponentConstructor>(),
  } satisfies Record<RegistryCategory, Map<string, unknown>>;

  category(name: "fields"): CategoryRegistry<FieldWidgetConstructor>;
  category(name: "views"): CategoryRegistry<ViewConstructor>;
  category(name: "services"): CategoryRegistry<ServiceInstance>;
  category(name: "main_components"): CategoryRegistry<MainComponentConstructor>;
  category(name: RegistryCategory): CategoryRegistry<unknown> {
    const store = this.maps[name];
    if (!store) throw new Error(`Unknown registry category: ${String(name)}`);
    return new CategoryRegistry(store);
  }

  get(category: "fields", key: string): FieldWidgetConstructor | undefined;
  get(category: "views", key: string): ViewConstructor | undefined;
  get(category: "services", key: string): ServiceInstance | undefined;
  get(category: "main_components", key: string): MainComponentConstructor | undefined;
  get(category: RegistryCategory, key: string): unknown {
    const store = this.maps[category];
    if (!store) throw new Error(`Unknown registry category: ${String(category)}`);
    return store.get(key);
  }
}

export const registry = new Registry();
