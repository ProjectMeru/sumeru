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

export type EventHandler = (ev: Event) => void;

export interface TemplateResult {
  render(): HTMLElement;
}

interface VNode {
  tag: string;
  attrs: Record<string, string>;
  handlers: Record<string, EventHandler>;
  children: VNodeChild[];
  key?: string;
}

type VNodeChild = VNode | string | HTMLElement;

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

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function flattenValues(values: TemplateValue[]): VNodeChild[] {
  const out: VNodeChild[] = [];
  for (const v of values) {
    if (v == null || v === false) continue;
    if (typeof v === "object" && "render" in v && typeof (v as TemplateResult).render === "function") {
      out.push((v as TemplateResult).render());
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
      const root = document.createElement("div");
      root.style.display = "contents";
      for (const node of vnodes) {
        if (typeof node === "string") {
          root.appendChild(document.createTextNode(node));
          continue;
        }
        if (node instanceof HTMLElement) {
          root.appendChild(node);
          continue;
        }
        root.appendChild(renderVNode(node));
      }
      if (root.childNodes.length === 1 && root.firstElementChild) {
        return root.firstElementChild as HTMLElement;
      }
      return root;
    },
  };
}

function applyStyle(el: HTMLElement, raw: string): void {
  for (const part of raw.split(";")) {
    const idx = part.indexOf(":");
    if (idx === -1) continue;
    const prop = part.slice(0, idx).trim();
    const val = part.slice(idx + 1).trim();
    if (prop) el.style.setProperty(prop, val);
  }
}

function renderVNode(vn: VNode): HTMLElement {
  const el = document.createElement(vn.tag);
  if (vn.key) el.dataset.swcKey = vn.key;
  for (const [k, v] of Object.entries(vn.attrs)) {
    if (k.startsWith("@")) {
      continue;
    }
    if (k.startsWith(".")) {
      el.classList.add(k.slice(1));
    } else if (k === "class" && v) {
      for (const c of v.split(/\s+/)) {
        if (c) el.classList.add(c);
      }
    } else if (k === "style" && v) {
      applyStyle(el, v);
    } else if (
      k.startsWith("data-") ||
      k === "id" ||
      k === "for" ||
      k === "href" ||
      k === "type" ||
      k === "name" ||
      k === "value" ||
      k === "placeholder" ||
      k === "autocomplete" ||
      k === "step" ||
      k === "tabindex" ||
      k === "aria-label" ||
      k === "aria-labelledby" ||
      k === "aria-controls" ||
      k === "title" ||
      k === "role" ||
      k === "aria-selected" ||
      k === "checked" ||
      k === "src" ||
      k === "alt" ||
      k === "rows" ||
      k === "selected" ||
      k === "method" ||
      k === "action" ||
      k === "enctype" ||
      k === "accept" ||
      k === "open" ||
      k === "hidden" ||
      k === "disabled"
    ) {
      el.setAttribute(k, v);
    }
  }
  for (const [event, handler] of Object.entries(vn.handlers)) {
    el.addEventListener(event, handler);
  }
  for (const child of vn.children) {
    if (typeof child === "string") {
      el.insertAdjacentHTML("beforeend", escapeHtml(child));
    } else if (child instanceof HTMLElement) {
      el.appendChild(child);
    } else {
      el.appendChild(renderVNode(child));
    }
  }
  return el;
}

export function fragment(children: TemplateValue[]): TemplateResult {
  return html`${children as unknown as TemplateValue}`;
}
