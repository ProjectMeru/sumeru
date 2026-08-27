import type { TemplateSourceMeta } from "../template/sum/meta.js";

export interface DevtoolsComponent {
  el: HTMLElement | null;
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
  getComponentForElement(el: Element): ComponentRecord | null;
  getTemplateSource(comp: ComponentRecord): TemplateSourceMeta | null;
}

let nextId = 1;
const components = new Map<number, ComponentRecord>();
const byElement = new WeakMap<Element, number>();

declare global {
  interface Window {
    __SWC_DEVTOOLS__?: SwcDevtoolsGlobal;
  }
}

export function registerComponent(comp: DevtoolsComponent, parentId: number | null = null): number {
  const id = nextId++;
  const name = comp.constructor.name || "Anonymous";
  const record: ComponentRecord = { id, name, component: comp, parentId };
  components.set(id, record);
  if (comp.el) byElement.set(comp.el, id);
  publish();
  return id;
}

export function unregisterComponent(comp: DevtoolsComponent): void {
  for (const [id, rec] of components) {
    if (rec.component === comp) {
      components.delete(id);
      if (comp.el) byElement.delete(comp.el);
      publish();
      return;
    }
  }
}

export function getComponentForElement(el: Element): ComponentRecord | null {
  const id = byElement.get(el);
  if (id === undefined) return null;
  return components.get(id) ?? null;
}

export function getTemplateSource(_comp: ComponentRecord): TemplateSourceMeta | null {
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
