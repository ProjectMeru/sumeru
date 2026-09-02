import { patchKeyedChildren } from "../runtime/patch/keyed.js";

export type TemplateValue =
  | string
  | number
  | boolean
  | null
  | undefined
  | TemplateResult
  | HTMLElement
  | TemplateValue[]
  | EventHandler;

export type ComponentProps = Record<string, unknown>;

export type EventHandler = (event: Event) => void;

export interface TemplateResult {
  render(): HTMLElement;
  patch(existing: HTMLElement): HTMLElement;
  key?: string;
}

interface VNode {
  tag: string;
  attrs: Record<string, string>;
  handlers: Record<string, EventHandler>;
  children: VNodeChild[];
  key?: string;
}

type VNodeChild = VNode | string | HTMLElement | TemplateResult;

const ALLOWED_ATTRS = new Set([
  "id",
  "for",
  "href",
  "type",
  "name",
  "value",
  "placeholder",
  "autocomplete",
  "step",
  "tabindex",
  "aria-label",
  "aria-labelledby",
  "aria-controls",
  "title",
  "role",
  "aria-selected",
  "checked",
  "src",
  "alt",
  "rows",
  "selected",
  "method",
  "action",
  "enctype",
  "accept",
  "open",
  "hidden",
  "disabled",
  "target",
  "rel",
]);

const elementHandlers = new WeakMap<HTMLElement, Record<string, EventHandler>>();

function isVNode(node: VNodeChild): node is VNode {
  return (
    typeof node === "object" &&
    node !== null &&
    !(node instanceof HTMLElement) &&
    !isTemplateResult(node) &&
    "tag" in node
  );
}

const VOID_ELEMENTS = new Set([
  "area",
  "base",
  "br",
  "col",
  "embed",
  "hr",
  "img",
  "input",
  "link",
  "meta",
  "param",
  "source",
  "track",
  "wbr",
]);

export function isTemplateResult(value: unknown): value is TemplateResult {
  return (
    typeof value === "object" &&
    value !== null &&
    "render" in value &&
    typeof (value as TemplateResult).render === "function" &&
    typeof (value as TemplateResult).patch === "function"
  );
}

function flattenValues(values: TemplateValue[]): VNodeChild[] {
  const out: VNodeChild[] = [];
  for (const v of values) {
    if (v == null || v === false) continue;
    if (isTemplateResult(v)) {
      out.push(v);
      continue;
    }
    if (v instanceof HTMLElement) {
      out.push(v);
      continue;
    }
    if (Array.isArray(v)) {
      out.push(...flattenValues(v as TemplateValue[]));
      continue;
    }
    if (typeof v === "function") continue;
    out.push(String(v));
  }
  return out;
}

function awaitingAttrName(partial: string): string | null {
  const m = partial.match(/([@\.\w-]+)=\s*$/);
  return m ? m[1] : null;
}

function stripAwaitingAttr(partial: string): string {
  const attr = awaitingAttrName(partial);
  if (!attr) return partial;
  return partial.slice(0, partial.length - attr.length - 1);
}

function isAttrWhitespace(ch: string): boolean {
  return ch === " " || ch === "\n" || ch === "\r" || ch === "\t";
}

function parseTagAttributes(raw: string): { tag: string; attrs: Record<string, string> } {
  const s = raw.trim();
  const tagMatch = s.match(/^([^\s/]+)/);
  const tag = tagMatch?.[1] || "div";
  const rest = s.slice(tag.length).trim();
  const attrs: Record<string, string> = {};
  let i = 0;
  while (i < rest.length) {
    while (i < rest.length && isAttrWhitespace(rest[i])) i++;
    if (i >= rest.length) break;

    const keyStart = i;
    while (i < rest.length && rest[i] !== "=" && !isAttrWhitespace(rest[i])) i++;
    const key = rest.slice(keyStart, i).trim();
    if (!key) break;

    if (i >= rest.length || rest[i] !== "=") {
      attrs[key] = "";
      continue;
    }
    i++;

    if (i >= rest.length) {
      attrs[key] = "";
      break;
    }

    if (rest[i] === '"' || rest[i] === "'") {
      const quote = rest[i++];
      let val = "";
      while (i < rest.length && rest[i] !== quote) val += rest[i++];
      if (i < rest.length) i++;
      attrs[key] = val;
      continue;
    }

    let val = "";
    while (i < rest.length && !isAttrWhitespace(rest[i])) val += rest[i++];
    attrs[key] = val;
  }
  return { tag, attrs };
}

function buildTree(strings: TemplateStringsArray, values: TemplateValue[]): VNodeChild[] {
  const root: VNodeChild[] = [];
  const stack: VNode[] = [];
  let partial: string | null = null;
  let partialHandlers: Record<string, EventHandler> = {};

  const parentChildren = (): VNodeChild[] => (stack.length > 0 ? stack[stack.length - 1].children : root);

  const appendChild = (child: VNodeChild): void => {
    parentChildren().push(child);
  };

  const openTag = (raw: string, handlers: Record<string, EventHandler> = {}): void => {
    const trimmed = raw.trim();
    const selfClosing = trimmed.endsWith("/");
    const inner = selfClosing ? trimmed.slice(0, -1).trim() : trimmed;
    const parsed = parseTagAttributes(inner);
    const node: VNode = {
      tag: parsed.tag,
      attrs: parsed.attrs,
      handlers: { ...handlers },
      children: [],
      key: parsed.attrs.key,
    };
    appendChild(node);
    if (!selfClosing && !VOID_ELEMENTS.has(parsed.tag.toLowerCase())) {
      stack.push(node);
    }
  };

  const closeTag = (raw: string): void => {
    const name = raw.slice(1).trim().split(/\s+/)[0]?.toLowerCase() ?? "";
    if (!name) return;
    while (stack.length > 0) {
      const top = stack[stack.length - 1];
      if (top.tag.toLowerCase() === name) {
        stack.pop();
        break;
      }
      stack.pop();
    }
  };

  const completePartial = (rest: string): void => {
    if (partial === null) return;
    let raw = partial + rest;
    if (raw.startsWith("<")) raw = raw.slice(1);
    openTag(raw, partialHandlers);
    partial = null;
    partialHandlers = {};
  };

  const processText = (text: string): void => {
    let i = 0;
    while (i < text.length) {
      if (partial !== null) {
        const gt = text.indexOf(">", i);
        if (gt === -1) {
          partial += text.slice(i);
          return;
        }
        completePartial(text.slice(i, gt));
        i = gt + 1;
        continue;
      }

      const lt = text.indexOf("<", i);
      if (lt === -1) {
        const tail = text.slice(i);
        if (tail) appendChild(tail);
        return;
      }

      if (lt > i) {
        appendChild(text.slice(i, lt));
      }

      const gt = text.indexOf(">", lt);
      if (gt === -1) {
        partial = text.slice(lt);
        return;
      }

      const inner = text.slice(lt + 1, gt);
      i = gt + 1;

      if (inner.startsWith("/")) {
        closeTag(inner);
      } else {
        openTag(inner);
      }
    }
  };

  const processValue = (value: TemplateValue): void => {
    if (partial !== null) {
      const attr = awaitingAttrName(partial);
      if (value == null || value === false) {
        if (attr) partial = stripAwaitingAttr(partial);
        return;
      }
      if (typeof value === "function") {
        if (attr) partialHandlers[attr.startsWith("@") ? attr.slice(1) : attr] = value as EventHandler;
        return;
      }
      if (isTemplateResult(value)) {
        // Optional attribute suffix fragments (e.g. html``) must not stringify to [object Object].
        return;
      }
      const text = String(value);
      const needsQuotes =
        attr &&
        !attr.startsWith("@") &&
        (text === "" ||
          text.includes(" ") ||
          text.includes("\n") ||
          text.includes("\t") ||
          text.includes(".") ||
          text.includes("=") ||
          text.includes("<") ||
          text.includes(">"));
      if (needsQuotes) {
        partial += `"${text.replace(/"/g, "&quot;")}"`;
      } else {
        partial += text;
      }
      return;
    }
    for (const child of flattenValues([value])) {
      appendChild(child);
    }
  };

  for (let i = 0; i < strings.length; i++) {
    if (strings[i]) processText(strings[i]);
    if (i < values.length) processValue(values[i]);
  }

  if (partial) appendChild(partial);

  return root;
}

export function html(strings: TemplateStringsArray, ...values: TemplateValue[]): TemplateResult {
  const vnodes = buildTree(strings, values);

  return {
    render(): HTMLElement {
      return materialize(vnodes);
    },
    patch(existing: HTMLElement): HTMLElement {
      return patchRoot(existing, vnodes);
    },
  };
}

function applyStyle(el: HTMLElement, raw: string): void {
  el.style.cssText = "";
  for (const part of raw.split(";")) {
    const idx = part.indexOf(":");
    if (idx === -1) continue;
    const prop = part.slice(0, idx).trim();
    const val = part.slice(idx + 1).trim();
    if (prop) el.style.setProperty(prop, val);
  }
}

function applyAttrs(el: HTMLElement, attrs: Record<string, string>, key?: string): void {
  if (key) el.dataset.swcKey = key;
  const nextNames = new Set<string>();
  const classes: string[] = [];
  for (const [k, v] of Object.entries(attrs)) {
    if (k.startsWith("@") || k === "key") continue;
    if (k.startsWith(".")) {
      classes.push(k.slice(1));
      continue;
    }
    if (k === "class") {
      for (const c of v.split(/\s+/)) {
        if (c) classes.push(c);
      }
      continue;
    }
    if (k === "style") {
      applyStyle(el, v);
      nextNames.add("style");
      continue;
    }
    if (k === "ref") {
      el.setAttribute("data-ref", v);
      nextNames.add("data-ref");
      continue;
    }
    if (k === "value") {
      el.setAttribute("value", v);
      nextNames.add("value");
      if (el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement) {
        if (el.value !== v) el.value = v;
      }
      continue;
    }
    if (k.startsWith("data-") || ALLOWED_ATTRS.has(k)) {
      el.setAttribute(k, v);
      nextNames.add(k);
    }
  }
  if (classes.length > 0) {
    el.className = classes.join(" ");
    nextNames.add("class");
  } else if (el.className) {
    el.removeAttribute("class");
  }
  for (const name of [...el.getAttributeNames()]) {
    if (name === "class" || name === "style" || name.startsWith("data-swc")) continue;
    if (!nextNames.has(name) && (name.startsWith("data-") || ALLOWED_ATTRS.has(name))) {
      el.removeAttribute(name);
    }
  }
}

function syncHandlers(el: HTMLElement, next: Record<string, EventHandler>): void {
  const previous = elementHandlers.get(el) ?? {};
  for (const [event, handler] of Object.entries(previous)) {
    if (next[event] !== handler) el.removeEventListener(event, handler);
  }
  for (const [event, handler] of Object.entries(next)) {
    if (previous[event] !== handler) el.addEventListener(event, handler);
  }
  elementHandlers.set(el, { ...next });
}

function renderChild(node: VNodeChild): Node {
  if (typeof node === "string") return document.createTextNode(node);
  if (node instanceof HTMLElement) return node;
  if (isTemplateResult(node)) return node.render();
  return renderVNode(node);
}

function materialize(vnodes: VNodeChild[]): HTMLElement {
  const root = document.createElement("div");
  root.style.display = "contents";
  for (const node of vnodes) {
    root.appendChild(renderChild(node));
  }
  if (root.childNodes.length === 1 && root.firstElementChild) {
    return root.firstElementChild as HTMLElement;
  }
  return root;
}

function childKey(node: VNodeChild): string | undefined {
  if (typeof node === "string") return undefined;
  if (node instanceof HTMLElement) return node.dataset.swcKey;
  if (isTemplateResult(node)) return node.key;
  return node.key;
}

function patchRoot(existing: HTMLElement, vnodes: VNodeChild[]): HTMLElement {
  if (vnodes.length === 1) {
    const only = vnodes[0];
    if (isTemplateResult(only) && typeof only.patch === "function") {
      return only.patch(existing);
    }
    if (only instanceof HTMLElement) {
      if (only === existing) return existing;
      return only;
    }
    if (isVNode(only) && existing.style.display !== "contents") {
      if (existing.tagName.toLowerCase() === only.tag.toLowerCase()) {
        patchVNode(existing, only);
        return existing;
      }
    }
  }
  if (existing.style.display === "contents") {
    patchChildren(existing, vnodes);
    return existing;
  }
  return materialize(vnodes);
}

function patchVNode(el: HTMLElement, vn: VNode): void {
  applyAttrs(el, vn.attrs, vn.key);
  syncHandlers(el, vn.handlers);
  patchChildren(el, vn.children);
}

function patchChildren(container: HTMLElement, children: VNodeChild[]): void {
  const meaningful = children.filter((c) => c !== "" && c != null);
  const keys = meaningful.map(childKey);
  if (meaningful.length > 0 && keys.every((k) => k)) {
    patchKeyedChildren(
      container,
      meaningful.map((child, index) => ({
        key: keys[index] as string,
        render: () => {
          const node = renderChild(child);
          return node instanceof HTMLElement ? node : wrapNode(node);
        },
        patch: (element: HTMLElement) => patchChildElement(element, child),
      })),
    );
    return;
  }

  const existingNodes = [...container.childNodes];
  let index = 0;
  for (const child of meaningful) {
    const current = existingNodes[index];
    const next = patchOrCreate(current, child);
    if (current && next === current) {
      index += 1;
      continue;
    }
    if (current) {
      container.replaceChild(next, current);
    } else {
      container.appendChild(next);
    }
    index += 1;
  }
  while (container.childNodes.length > index) {
    container.lastChild?.remove();
  }
}

function wrapNode(node: Node): HTMLElement {
  if (node instanceof HTMLElement) return node;
  const span = document.createElement("span");
  span.style.display = "contents";
  span.appendChild(node);
  return span;
}

function patchChildElement(existing: HTMLElement, child: VNodeChild): HTMLElement {
  if (isTemplateResult(child)) return child.patch(existing);
  if (child instanceof HTMLElement) return child;
  if (isVNode(child) && existing.tagName.toLowerCase() === child.tag.toLowerCase()) {
    patchVNode(existing, child);
    return existing;
  }
  const rendered = renderChild(child);
  return rendered instanceof HTMLElement ? rendered : wrapNode(rendered);
}

function patchOrCreate(current: ChildNode | undefined, child: VNodeChild): Node {
  if (!current) return renderChild(child);
  if (typeof child === "string") {
    if (current.nodeType === Node.TEXT_NODE) {
      if (current.textContent !== child) current.textContent = child;
      return current;
    }
    return document.createTextNode(child);
  }
  if (current instanceof HTMLElement) {
    return patchChildElement(current, child);
  }
  return renderChild(child);
}

function renderVNode(vn: VNode): HTMLElement {
  const el = document.createElement(vn.tag);
  applyAttrs(el, vn.attrs, vn.key);
  syncHandlers(el, vn.handlers);
  for (const child of vn.children) {
    el.appendChild(renderChild(child));
  }
  return el;
}

export function fragment(children: TemplateValue[]): TemplateResult {
  return html`${children as unknown as TemplateValue}`;
}
