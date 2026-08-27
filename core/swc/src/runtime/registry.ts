import type { SwcComponent } from "./component.js";

export type RegistryCategory = "fields" | "views" | "services" | "main_components" | "actions" | "systray";

export type RegistryEntry = new (props: object, env: import("./env.js").SwcEnv) => SwcComponent;

export class Registry {
  private readonly entries = new Map<RegistryCategory, Map<string, RegistryEntry>>();

  category(name: RegistryCategory): CategoryRegistry {
    if (!this.entries.has(name)) {
      this.entries.set(name, new Map());
    }
    return new CategoryRegistry(this.entries.get(name)!);
  }

  get(category: RegistryCategory, key: string): RegistryEntry | undefined {
    return this.entries.get(category)?.get(key);
  }
}

export class CategoryRegistry {
  constructor(private readonly map: Map<string, RegistryEntry>) {}

  add(key: string, value: RegistryEntry): void {
    this.map.set(key, value);
  }

  get(key: string): RegistryEntry | undefined {
    return this.map.get(key);
  }

  keys(): string[] {
    return [...this.map.keys()];
  }
}

export const registry = new Registry();
