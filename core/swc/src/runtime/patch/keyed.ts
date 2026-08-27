/** Keyed DOM reconciliation for list rows and t-key siblings. */

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
  items: Array<{ key: string; render: () => HTMLElement }>,
): void {
  const prev = collectKeyedChildren(container);
  const nextKeys = new Set<string>();
  const ordered: HTMLElement[] = [];

  for (const item of items) {
    nextKeys.add(item.key);
    let el = prev.get(item.key);
    if (!el) {
      el = item.render();
      el.dataset.swcKey = item.key;
    }
    ordered.push(el);
  }

  for (const [key, el] of prev) {
    if (!nextKeys.has(key)) el.remove();
  }

  for (let i = 0; i < ordered.length; i++) {
    const el = ordered[i];
    const current = container.children[i] as HTMLElement | undefined;
    if (current !== el) {
      container.insertBefore(el, current ?? null);
    }
  }

  while (container.children.length > ordered.length) {
    container.lastElementChild?.remove();
  }
}

/** Patch a keyed-list wrapper (display:contents div from KeyedListResult). */
export function patchKeyedListWrapper(
  wrapper: HTMLElement,
  items: Array<{ key: string; render: () => HTMLElement }>,
): void {
  patchKeyedChildren(wrapper, items);
}
