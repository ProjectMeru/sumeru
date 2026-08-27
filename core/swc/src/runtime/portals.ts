const originalParent = new WeakMap<HTMLElement, { parent: Node; next: Node | null }>();
const movedFromRoot = new WeakMap<HTMLElement, HTMLElement[]>();

function portalNodesUnder(root: HTMLElement): HTMLElement[] {
  const nodes = [...root.querySelectorAll<HTMLElement>("[data-portal]")];
  if (root.matches("[data-portal]")) nodes.unshift(root);
  return nodes;
}

/** Move `[data-portal]` nodes into the matching selector (document.body if missing). */
export function applyPortals(root: HTMLElement | null): void {
  if (!root) return;
  const tracked = movedFromRoot.get(root) ?? [];
  for (const node of portalNodesUnder(root)) {
    const selector = node.dataset.portal?.trim();
    if (!selector) continue;
    if (!originalParent.has(node)) {
      originalParent.set(node, { parent: node.parentNode ?? root, next: node.nextSibling });
    }
    if (!tracked.includes(node)) tracked.push(node);
    const target = document.querySelector(selector) ?? document.body;
    if (node.parentNode !== target) {
      target.appendChild(node);
    }
  }
  movedFromRoot.set(root, tracked);
}

/** Put portaled nodes back under their original parent (or remove). */
export function restorePortals(root: HTMLElement | null): void {
  if (!root) return;
  const nodes = movedFromRoot.get(root) ?? portalNodesUnder(root);
  movedFromRoot.delete(root);
  for (const node of nodes) {
    const origin = originalParent.get(node);
    originalParent.delete(node);
    if (!origin) {
      node.remove();
      continue;
    }
    origin.parent.insertBefore(node, origin.next);
  }
}
