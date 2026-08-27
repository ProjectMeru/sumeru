/** Static block DOM with dynamic holes (incremental patch foundation). */

export interface StaticBlock {
  render(values: unknown[]): HTMLElement;
  patch(el: HTMLElement, values: unknown[]): void;
}

export function createBlock(
  staticHtml: string,
  holeCount: number,
  patchFn: (el: HTMLElement, values: unknown[]) => void,
): StaticBlock {
  const template = document.createElement("template");
  template.innerHTML = staticHtml.trim();

  return {
    render(values: unknown[]): HTMLElement {
      const clone = template.content.cloneNode(true) as DocumentFragment;
      const el = (clone.firstElementChild ?? clone) as HTMLElement;
      if (holeCount > 0) patchFn(el, values);
      return el;
    },
    patch(el: HTMLElement, values: unknown[]): void {
      patchFn(el, values);
    },
  };
}

export function textHole(el: HTMLElement, selector: string, index: number, values: unknown[]): void {
  const node = el.querySelector(selector);
  if (node) node.textContent = String(values[index] ?? "");
}

export function attrHole(
  el: HTMLElement,
  selector: string,
  attr: string,
  index: number,
  values: unknown[],
): void {
  const node = el.querySelector(selector);
  if (node) node.setAttribute(attr, String(values[index] ?? ""));
}
