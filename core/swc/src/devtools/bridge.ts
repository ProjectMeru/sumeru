import type { TemplateSourceMeta } from "../template/sum/meta.js";

export interface DevtoolsComponent {
  rootElement: HTMLElement | null;
  constructor: { name: string };
  props: object;
}

export interface ComponentRecord {
  id: number;
  name: string;
  component: DevtoolsComponent;
  parentId: number | null;
}

export interface SwcDevtoolsGlobal {
  apps: unknown[];
  components: ComponentRecord[];
  getComponentForElement(element: Element): ComponentRecord | null;
  getTemplateSource(component: ComponentRecord): TemplateSourceMeta | null;
}

let nextId = 1;
const components = new Map<number, ComponentRecord>();
const byElement = new WeakMap<Element, number>();

declare global {
  interface Window {
    __SWC_DEVTOOLS__?: SwcDevtoolsGlobal;
  }
}

export function registerComponent(component: DevtoolsComponent, parentId: number | null = null): number {
  const id = nextId++;
  const name = component.constructor.name || "Anonymous";
  const record: ComponentRecord = { id, name, component, parentId };
  components.set(id, record);
  if (component.rootElement) byElement.set(component.rootElement, id);
  publish();
  return id;
}

export function unregisterComponent(component: DevtoolsComponent): void {
  for (const [id, record] of components) {
    if (record.component === component) {
      components.delete(id);
      if (component.rootElement) byElement.delete(component.rootElement);
      publish();
      return;
    }
  }
}

export function getComponentForElement(element: Element): ComponentRecord | null {
  const id = byElement.get(element);
  if (id === undefined) return null;
  return components.get(id) ?? null;
}

export function getTemplateSource(_component: ComponentRecord): TemplateSourceMeta | null {
  return null;
}

function publish(): void {
  if (typeof window === "undefined") return;
  window.__SWC_DEVTOOLS__ = {
    apps: [],
    components: [...components.values()],
    getComponentForElement,
    getTemplateSource,
  };
}

export function initDevtoolsBridge(): void {
  publish();
}
