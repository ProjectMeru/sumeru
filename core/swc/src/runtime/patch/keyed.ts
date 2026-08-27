/** Keyed DOM reconciliation for list rows (`data-swc-key` siblings). */

export function collectKeyedChildren(container: HTMLElement): Map<string, HTMLElement> {
  const map = new Map<string, HTMLElement>();
  for (const child of container.children) {
    if (!(child instanceof HTMLElement)) continue;
    const key = child.dataset.swcKey;
    if (key) map.set(key, child);
  }
  return map;
}

export function patchKeyedChildren(
  container: HTMLElement,
  items: Array<{
    key: string;
    render: () => HTMLElement;
    patch?: (element: HTMLElement) => HTMLElement;
  }>,
): void {
  const prev = collectKeyedChildren(container);
  const nextKeys = new Set<string>();
  const ordered: HTMLElement[] = [];

  for (const item of items) {
    nextKeys.add(item.key);
    let element = prev.get(item.key);
    if (!element) {
      element = item.render();
      element.dataset.swcKey = item.key;
    } else if (item.patch) {
      element = item.patch(element);
      element.dataset.swcKey = item.key;
    }
    ordered.push(element);
  }

  for (const [key, element] of prev) {
    if (!nextKeys.has(key)) element.remove();
  }

  for (let index = 0; index < ordered.length; index++) {
    const element = ordered[index];
    const current = container.children[index] as HTMLElement | undefined;
    if (current !== element) {
      container.insertBefore(element, current ?? null);
    }
  }

  while (container.children.length > ordered.length) {
    container.lastElementChild?.remove();
  }
}
