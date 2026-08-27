"use strict";
var SumeruSWC = (() => {
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __export = (target, all) => {
    for (var name in all)
      __defProp(target, name, { get: all[name], enumerable: true });
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

  // src/main.ts
  var main_exports = {};
  __export(main_exports, {
    SwcApp: () => SwcApp,
    SwcEnv: () => SwcEnv,
    registry: () => registry
  });

  // src/runtime/scheduler.ts
  var pending = /* @__PURE__ */ new Set();
  var frameHandle = 0;
  function queueFrame(callback) {
    if (typeof requestAnimationFrame === "function") {
      return requestAnimationFrame(callback);
    }
    return setTimeout(callback, 0);
  }
  function cancelFrame(handle) {
    if (typeof cancelAnimationFrame === "function") {
      cancelAnimationFrame(handle);
      return;
    }
    clearTimeout(handle);
  }
  function scheduleRender(component) {
    pending.add(component);
    if (frameHandle) return;
    frameHandle = queueFrame(flushScheduledRenders);
  }
  function flushScheduledRenders() {
    if (frameHandle) {
      cancelFrame(frameHandle);
      frameHandle = 0;
    }
    const batch = [...pending];
    pending.clear();
    for (const component of batch) {
      if (component.rootElement?.isConnected) component.patch();
    }
  }

  // src/runtime/hooks.ts
  var activeHost = null;
  function getActiveHost() {
    return activeHost;
  }
  function requireActiveHost() {
    if (!activeHost) {
      throw new Error("Hooks must run inside callSetup()");
    }
    return activeHost;
  }
  function withActiveHost(host, fn) {
    const previous = activeHost;
    activeHost = host;
    try {
      return fn();
    } finally {
      activeHost = previous;
    }
  }
  function onMount(fn) {
    requireActiveHost().mountEffects.push(fn);
  }
  function onWillUnmount(fn) {
    requireActiveHost().unmountEffects.push(fn);
  }
  function useEffect(fn) {
    onMount(() => {
      const cleanup = fn();
      if (typeof cleanup === "function") {
        onWillUnmount(cleanup);
      }
    });
  }
  function useState(initial) {
    const host = requireActiveHost();
    const index = host.consumeHookSlot();
    if (host.hookState[index] === void 0) {
      host.hookState[index] = initial;
    }
    const box = {
      get value() {
        return host.hookState[index];
      }
    };
    const setValue = (next) => {
      const previous = host.hookState[index];
      const value = typeof next === "function" ? next(previous) : next;
      if (Object.is(value, previous)) return;
      host.hookState[index] = value;
      scheduleRender(host);
    };
    return [box, setValue];
  }
  function runMountEffects(host) {
    for (const fn of host.mountEffects) {
      fn();
    }
  }
  function runUnmountEffects(host) {
    for (const fn of host.unmountEffects) {
      fn();
    }
  }

  // src/runtime/lifecycle.ts
  function requireActiveHost2() {
    const host = getActiveHost();
    if (!host) {
      throw new Error("Lifecycle hooks must run inside callSetup()");
    }
    return host;
  }
  function onWillStart(fn) {
    requireActiveHost2().willStart.push(fn);
  }
  async function runWillStart(host) {
    for (const fn of host.willStart) {
      await fn();
    }
  }
  function runWillPatch(host) {
    for (const fn of host.willPatch) {
      fn();
    }
  }
  function runPatched(host) {
    for (const fn of host.patched) {
      fn();
    }
  }

  // src/devtools/bridge.ts
  var nextId = 1;
  var components = /* @__PURE__ */ new Map();
  var byElement = /* @__PURE__ */ new WeakMap();
  function registerComponent(component, parentId = null) {
    const id = nextId++;
    const name = component.constructor.name || "Anonymous";
    const record = { id, name, component, parentId };
    components.set(id, record);
    if (component.rootElement) byElement.set(component.rootElement, id);
    publish();
    return id;
  }
  function unregisterComponent(component) {
    for (const [id, record] of components) {
      if (record.component === component) {
        components.delete(id);
        if (component.rootElement) byElement.delete(component.rootElement);
        publish();
        return;
      }
    }
  }
  function getComponentForElement(element) {
    const id = byElement.get(element);
    if (id === void 0) return null;
    return components.get(id) ?? null;
  }
  function getTemplateSource(_component) {
    return null;
  }
  function publish() {
    if (typeof window === "undefined") return;
    window.__SWC_DEVTOOLS__ = {
      apps: [],
      components: [...components.values()],
      getComponentForElement,
      getTemplateSource
    };
  }
  function initDevtoolsBridge() {
    publish();
  }

  // src/runtime/portals.ts
  var originalParent = /* @__PURE__ */ new WeakMap();
  var movedFromRoot = /* @__PURE__ */ new WeakMap();
  function portalNodesUnder(root) {
    const nodes = [...root.querySelectorAll("[data-portal]")];
    if (root.matches("[data-portal]")) nodes.unshift(root);
    return nodes;
  }
  function applyPortals(root) {
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
  function restorePortals(root) {
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

  // src/runtime/component.ts
  var SwcComponent = class {
    props;
    env;
    rootElement = null;
    mounted = false;
    hookState = [];
    hookIndex = 0;
    willStart = [];
    willPatch = [];
    patched = [];
    mountEffects = [];
    unmountEffects = [];
    constructor(props, env) {
      this.props = props;
      this.env = env;
    }
    consumeHookSlot() {
      const index = this.hookIndex;
      this.hookIndex += 1;
      return index;
    }
    /** Run `setup` with this instance as the active hook host. */
    callSetup() {
      this.hookIndex = 0;
      withActiveHost(this, () => {
        this.setup?.();
      });
    }
    /** Called when props are updated on an existing instance (SPA navigation). */
    onPropsChanged(_props) {
    }
    /** Replace props and re-render without recreating the component instance. */
    updateProps(next) {
      this.props = next;
      this.onPropsChanged(next);
      this.patch();
    }
    /** Queue a patch if this component is still in the document. */
    rerender() {
      if (this.rootElement?.isConnected) scheduleRender(this);
    }
    /** Patch in place when a root already exists; otherwise produce a new root. */
    renderOrPatch() {
      if (this.rootElement) {
        this.patch();
        return this.rootElement;
      }
      return this.render();
    }
    render() {
      const result = this.template();
      const root = result.render();
      this.rootElement = root;
      if (!this.mounted) {
        this.mounted = true;
        registerComponent(this);
        applyPortals(root);
        runMountEffects(this);
        this.onMount?.();
      }
      return root;
    }
    patch() {
      if (!this.rootElement) return;
      runWillPatch(this);
      const previousRoot = this.rootElement;
      const result = this.template();
      const next = result.patch(previousRoot);
      if (next !== previousRoot && previousRoot.parentNode) {
        previousRoot.replaceWith(next);
      }
      this.rootElement = next;
      applyPortals(next);
      runPatched(this);
      this.afterPatch?.();
    }
    destroy() {
      runUnmountEffects(this);
      this.onWillUnmount?.();
      restorePortals(this.rootElement);
      unregisterComponent(this);
      this.rootElement?.remove();
      this.rootElement = null;
      this.mounted = false;
    }
  };

  // src/runtime/patch/keyed.ts
  function collectKeyedChildren(container) {
    const map = /* @__PURE__ */ new Map();
    for (const child of container.children) {
      if (!(child instanceof HTMLElement)) continue;
      const key = child.dataset.swcKey;
      if (key) map.set(key, child);
    }
    return map;
  }
  function patchKeyedChildren(container, items) {
    const prev = collectKeyedChildren(container);
    const nextKeys = /* @__PURE__ */ new Set();
    const ordered = [];
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
      const current = container.children[index];
      if (current !== element) {
        container.insertBefore(element, current ?? null);
      }
    }
    while (container.children.length > ordered.length) {
      container.lastElementChild?.remove();
    }
  }

  // src/template/html.ts
  var ALLOWED_ATTRS = /* @__PURE__ */ new Set([
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
    "rel"
  ]);
  var elementHandlers = /* @__PURE__ */ new WeakMap();
  function isVNode(node) {
    return typeof node === "object" && node !== null && !(node instanceof HTMLElement) && !isTemplateResult(node) && "tag" in node;
  }
  var VOID_ELEMENTS = /* @__PURE__ */ new Set([
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
    "wbr"
  ]);
  function isTemplateResult(value) {
    return typeof value === "object" && value !== null && "render" in value && typeof value.render === "function" && typeof value.patch === "function";
  }
  function flattenValues(values) {
    const out = [];
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
        out.push(...flattenValues(v));
        continue;
      }
      if (typeof v === "function") continue;
      out.push(String(v));
    }
    return out;
  }
  function awaitingAttrName(partial) {
    const m = partial.match(/([@\.\w-]+)=\s*$/);
    return m ? m[1] : null;
  }
  function stripAwaitingAttr(partial) {
    const attr = awaitingAttrName(partial);
    if (!attr) return partial;
    return partial.slice(0, partial.length - attr.length - 1);
  }
  function isAttrWhitespace(ch) {
    return ch === " " || ch === "\n" || ch === "\r" || ch === "	";
  }
  function parseTagAttributes(raw) {
    const s = raw.trim();
    const tagMatch = s.match(/^([^\s/]+)/);
    const tag = tagMatch?.[1] || "div";
    const rest = s.slice(tag.length).trim();
    const attrs = {};
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
        let val2 = "";
        while (i < rest.length && rest[i] !== quote) val2 += rest[i++];
        if (i < rest.length) i++;
        attrs[key] = val2;
        continue;
      }
      let val = "";
      while (i < rest.length && !isAttrWhitespace(rest[i])) val += rest[i++];
      attrs[key] = val;
    }
    return { tag, attrs };
  }
  function buildTree(strings, values) {
    const root = [];
    const stack = [];
    let partial = null;
    let partialHandlers = {};
    const parentChildren = () => stack.length > 0 ? stack[stack.length - 1].children : root;
    const appendChild = (child) => {
      parentChildren().push(child);
    };
    const openTag = (raw, handlers = {}) => {
      const trimmed = raw.trim();
      const selfClosing = trimmed.endsWith("/");
      const inner = selfClosing ? trimmed.slice(0, -1).trim() : trimmed;
      const parsed = parseTagAttributes(inner);
      const node = {
        tag: parsed.tag,
        attrs: parsed.attrs,
        handlers: { ...handlers },
        children: [],
        key: parsed.attrs.key
      };
      appendChild(node);
      if (!selfClosing && !VOID_ELEMENTS.has(parsed.tag.toLowerCase())) {
        stack.push(node);
      }
    };
    const closeTag = (raw) => {
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
    const completePartial = (rest) => {
      if (partial === null) return;
      let raw = partial + rest;
      if (raw.startsWith("<")) raw = raw.slice(1);
      openTag(raw, partialHandlers);
      partial = null;
      partialHandlers = {};
    };
    const processText = (text) => {
      let i = 0;
      while (i < text.length) {
        if (partial !== null) {
          const gt2 = text.indexOf(">", i);
          if (gt2 === -1) {
            partial += text.slice(i);
            return;
          }
          completePartial(text.slice(i, gt2));
          i = gt2 + 1;
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
    const processValue = (value) => {
      if (partial !== null) {
        const attr = awaitingAttrName(partial);
        if (value == null || value === false) {
          if (attr) partial = stripAwaitingAttr(partial);
          return;
        }
        if (typeof value === "function") {
          if (attr) partialHandlers[attr.startsWith("@") ? attr.slice(1) : attr] = value;
          return;
        }
        const text = String(value);
        const needsQuotes = attr && !attr.startsWith("@") && (text === "" || text.includes(" ") || text.includes("\n") || text.includes("	") || text.includes(".") || text.includes("=") || text.includes("<") || text.includes(">"));
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
  function html(strings, ...values) {
    const vnodes = buildTree(strings, values);
    return {
      render() {
        return materialize(vnodes);
      },
      patch(existing) {
        return patchRoot(existing, vnodes);
      }
    };
  }
  function applyStyle(el, raw) {
    el.style.cssText = "";
    for (const part of raw.split(";")) {
      const idx = part.indexOf(":");
      if (idx === -1) continue;
      const prop = part.slice(0, idx).trim();
      const val = part.slice(idx + 1).trim();
      if (prop) el.style.setProperty(prop, val);
    }
  }
  function applyAttrs(el, attrs, key) {
    if (key) el.dataset.swcKey = key;
    const nextNames = /* @__PURE__ */ new Set();
    const classes = [];
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
  function syncHandlers(el, next) {
    const previous = elementHandlers.get(el) ?? {};
    for (const [event, handler] of Object.entries(previous)) {
      if (next[event] !== handler) el.removeEventListener(event, handler);
    }
    for (const [event, handler] of Object.entries(next)) {
      if (previous[event] !== handler) el.addEventListener(event, handler);
    }
    elementHandlers.set(el, { ...next });
  }
  function renderChild(node) {
    if (typeof node === "string") return document.createTextNode(node);
    if (node instanceof HTMLElement) return node;
    if (isTemplateResult(node)) return node.render();
    return renderVNode(node);
  }
  function materialize(vnodes) {
    const root = document.createElement("div");
    root.style.display = "contents";
    for (const node of vnodes) {
      root.appendChild(renderChild(node));
    }
    if (root.childNodes.length === 1 && root.firstElementChild) {
      return root.firstElementChild;
    }
    return root;
  }
  function childKey(node) {
    if (typeof node === "string") return void 0;
    if (node instanceof HTMLElement) return node.dataset.swcKey;
    if (isTemplateResult(node)) return node.key;
    return node.key;
  }
  function patchRoot(existing, vnodes) {
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
  function patchVNode(el, vn) {
    applyAttrs(el, vn.attrs, vn.key);
    syncHandlers(el, vn.handlers);
    patchChildren(el, vn.children);
  }
  function patchChildren(container, children) {
    const meaningful = children.filter((c) => c !== "" && c != null);
    const keys = meaningful.map(childKey);
    if (meaningful.length > 0 && keys.every((k) => k)) {
      patchKeyedChildren(
        container,
        meaningful.map((child, index2) => ({
          key: keys[index2],
          render: () => {
            const node = renderChild(child);
            return node instanceof HTMLElement ? node : wrapNode(node);
          },
          patch: (element) => patchChildElement(element, child)
        }))
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
  function wrapNode(node) {
    if (node instanceof HTMLElement) return node;
    const span = document.createElement("span");
    span.style.display = "contents";
    span.appendChild(node);
    return span;
  }
  function patchChildElement(existing, child) {
    if (isTemplateResult(child)) return child.patch(existing);
    if (child instanceof HTMLElement) return child;
    if (isVNode(child) && existing.tagName.toLowerCase() === child.tag.toLowerCase()) {
      patchVNode(existing, child);
      return existing;
    }
    const rendered = renderChild(child);
    return rendered instanceof HTMLElement ? rendered : wrapNode(rendered);
  }
  function patchOrCreate(current, child) {
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
  function renderVNode(vn) {
    const el = document.createElement(vn.tag);
    applyAttrs(el, vn.attrs, vn.key);
    syncHandlers(el, vn.handlers);
    for (const child of vn.children) {
      el.appendChild(renderChild(child));
    }
    return el;
  }

  // src/runtime/error.ts
  var SwcError = class extends Error {
    code;
    details;
    constructor(message, code = "swc_error", details) {
      super(message);
      this.name = "SwcError";
      this.code = code;
      this.details = details;
    }
  };

  // src/runtime/app.ts
  var ErrorBoundary = class extends SwcComponent {
    template() {
      const { error, retry } = this.props;
      return html`
      <div class="sum-flash sum-flash--error">
        <strong>SWC error</strong>
        <p>${error.message}</p>
        <button type="button" class="sum-btn sum-btn--secondary" @click=${() => retry()}>Retry</button>
      </div>
    `;
    }
  };
  function renderErrorFallback(error, retry) {
    const wrap = document.createElement("div");
    wrap.className = "sum-flash sum-flash--error";
    const title = document.createElement("strong");
    title.textContent = "SWC error";
    wrap.appendChild(title);
    const message = document.createElement("p");
    message.textContent = error.message;
    wrap.appendChild(message);
    const button = document.createElement("button");
    button.type = "button";
    button.className = "sum-btn sum-btn--secondary";
    button.textContent = "Retry";
    button.addEventListener("click", retry);
    wrap.appendChild(button);
    return wrap;
  }
  function showError(rootEl, env, error, retry) {
    try {
      const boundary = new ErrorBoundary({ error, retry }, env);
      rootEl.replaceChildren(boundary.render());
    } catch {
      rootEl.replaceChildren(renderErrorFallback(error, retry));
    }
  }
  var SwcApp = class _SwcApp {
    env;
    Root;
    rootEl = null;
    component = null;
    constructor(env, Root) {
      this.env = env;
      this.Root = Root;
    }
    static start(mountEl, env, Root) {
      const app = new _SwcApp(env, Root);
      void app.mount(mountEl);
      return app;
    }
    async mount(element) {
      this.rootEl = element;
      await this.renderRoot();
    }
    async renderRoot() {
      if (!this.rootEl) return;
      try {
        if (!this.component) {
          this.component = new this.Root({}, this.env);
          this.component.callSetup();
          await runWillStart(this.component);
          this.rootEl.replaceChildren(this.component.render());
        } else {
          this.component.patch();
        }
      } catch (err) {
        const swcErr = err instanceof SwcError ? err : new SwcError(String(err));
        showError(this.rootEl, this.env, swcErr, () => this.retry());
      }
    }
    retry() {
      this.component?.destroy();
      this.component = null;
      void this.renderRoot();
    }
    destroy() {
      this.component?.destroy();
      this.component = null;
      this.rootEl = null;
    }
  };

  // src/runtime/env.ts
  var SwcEnv = class {
    bootstrap;
    services;
    constructor(bootstrap2, services) {
      this.bootstrap = bootstrap2;
      this.services = services;
    }
  };

  // src/types/bootstrap.ts
  function readBootstrap() {
    const boot = window.__SWC_BOOTSTRAP__;
    if (!boot) {
      throw new Error("SWC bootstrap missing on window.__SWC_BOOTSTRAP__");
    }
    return boot;
  }

  // src/services/rpc.ts
  var RpcService = class {
    url;
    csrfToken;
    searchReadCache = /* @__PURE__ */ new Map();
    constructor(url, csrfToken) {
      this.url = url;
      this.csrfToken = csrfToken;
    }
    searchReadKey(model, domain, fields, limit) {
      return JSON.stringify({ model, domain, fields, limit });
    }
    /** Clears cached search_read results (e.g. after writes). */
    invalidateSearchReadCache() {
      this.searchReadCache.clear();
    }
    async dispatch(model, method, args = [], kwargs = {}) {
      const body = { model, method, args };
      if (Object.keys(kwargs).length > 0) {
        body.kwargs = kwargs;
      }
      const res = await fetch(this.url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": this.csrfToken
        },
        credentials: "same-origin",
        body: JSON.stringify(body)
      });
      if (!res.ok) {
        throw new SwcError(`RPC HTTP ${res.status}`, "rpc_http");
      }
      const data = await res.json();
      if (data.ok === false || data.error) {
        throw new SwcError(data.error?.message ?? "RPC failed", "rpc_error", data.error);
      }
      return data.result;
    }
    searchRead(model, domain = [], fields = [], limit = 80) {
      const key = this.searchReadKey(model, domain, fields, limit);
      let pending2 = this.searchReadCache.get(key);
      if (!pending2) {
        pending2 = this.dispatch(model, "search_read", [domain, fields], {
          limit
        });
        this.searchReadCache.set(key, pending2);
        void pending2.catch(() => {
          this.searchReadCache.delete(key);
        });
      }
      return pending2;
    }
    read(model, ids, fields = []) {
      return this.dispatch(model, "read", [ids, fields]);
    }
    write(model, ids, values) {
      this.invalidateSearchReadCache();
      return this.dispatch(model, "write", [ids, values]);
    }
    create(model, values) {
      this.invalidateSearchReadCache();
      return this.dispatch(model, "create", [values]);
    }
    unlink(model, ids) {
      this.invalidateSearchReadCache();
      return this.dispatch(model, "unlink", [ids]);
    }
    callMethod(model, method, recordId, vals) {
      const args = vals ? [recordId, method, vals] : [recordId, method];
      return this.dispatch(model, "call", args);
    }
    readGroup(model, domain, fields, groupBy, limit = 80) {
      const spec = {
        domain,
        groupby: groupBy,
        fields: fields.map((name) => ({ name, field: name, measure: name === "id" ? "count" : "sum" }))
      };
      return this.dispatch(model, "read_group", [spec], { limit });
    }
    onchange(model, values, field) {
      return this.dispatch(model, "onchange", [values, field]);
    }
  };

  // src/services/http.ts
  var HttpService = class {
    csrfToken;
    constructor(csrfToken) {
      this.csrfToken = csrfToken;
    }
    get csrf() {
      return this.csrfToken;
    }
    async getJSON(url) {
      const res = await fetch(url, {
        credentials: "same-origin",
        headers: { Accept: "application/json" }
      });
      if (!res.ok) {
        throw new SwcError(`GET ${url} failed: ${res.status}`, "http_get");
      }
      return await res.json();
    }
    async postForm(url, data) {
      const body = new URLSearchParams({ ...data, csrf_token: this.csrfToken });
      return fetch(url, {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body
      });
    }
    async postJSON(url, body) {
      const payload = { ...body, csrf_token: this.csrfToken };
      const res = await fetch(url, {
        method: "POST",
        credentials: "same-origin",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
          "X-CSRF-Token": this.csrfToken
        },
        body: JSON.stringify(payload)
      });
      if (!res.ok) {
        throw new SwcError(`POST ${url} failed: ${res.status}`, "http_post");
      }
      return await res.json();
    }
  };

  // src/services/notification.ts
  var MAX_TOASTS = 5;
  var EXIT_MS = 250;
  var NotificationService = class {
    stack;
    constructor(stackEl) {
      this.stack = stackEl ?? document.getElementById("sum-toast-stack") ?? this.createStack();
    }
    createStack() {
      const el = document.createElement("div");
      el.id = "sum-toast-stack";
      el.className = "sum-toast-stack";
      el.setAttribute("aria-live", "polite");
      document.body.appendChild(el);
      return el;
    }
    success(title, body, details) {
      this.show({ kind: "success", title, body, details });
    }
    error(title, body, details) {
      this.show({ kind: "error", title, body, details });
    }
    warning(title, body, details) {
      this.show({ kind: "warning", title, body, details });
    }
    show(msg, timeoutMs = 6e3) {
      this.capStack();
      const toast = this.buildToast(msg);
      this.stack.appendChild(toast);
      this.armTimer(toast, timeoutMs);
    }
    bootstrap(messages) {
      for (const m of messages ?? []) {
        this.show(m);
      }
    }
    liveToasts() {
      return [...this.stack.children].filter(
        (el) => el instanceof HTMLElement && !el.classList.contains("sum-toast-out")
      );
    }
    capStack() {
      const live = this.liveToasts();
      while (live.length >= MAX_TOASTS) {
        const oldest = live.shift();
        if (oldest) this.dismiss(oldest);
      }
    }
    buildToast(msg) {
      const toast = document.createElement("div");
      toast.className = `sum-toast sum-toast--${msg.kind || "info"}`;
      toast.setAttribute("role", "status");
      const title = document.createElement("span");
      title.className = "sum-toast-title";
      title.textContent = msg.title;
      const body = document.createElement("p");
      body.className = "sum-toast-body";
      body.textContent = msg.body;
      toast.append(title, body);
      if (msg.details) {
        const details = document.createElement("pre");
        details.className = "sum-toast-details";
        details.textContent = msg.details;
        toast.append(details);
      }
      const close = document.createElement("button");
      close.type = "button";
      close.className = "sum-toast-close";
      close.setAttribute("aria-label", "Close");
      close.textContent = "\xD7";
      close.addEventListener("click", () => this.dismiss(toast));
      toast.append(close);
      return toast;
    }
    armTimer(toast, timeoutMs) {
      let remaining = timeoutMs;
      let started = Date.now();
      let timer = window.setTimeout(() => this.dismiss(toast), remaining);
      toast.addEventListener("mouseenter", () => {
        window.clearTimeout(timer);
        remaining -= Date.now() - started;
      });
      toast.addEventListener("mouseleave", () => {
        started = Date.now();
        timer = window.setTimeout(() => this.dismiss(toast), Math.max(0, remaining));
      });
    }
    dismiss(toast) {
      if (!toast.isConnected || toast.classList.contains("sum-toast-out")) return;
      toast.classList.add("sum-toast-out");
      let removed = false;
      const remove = () => {
        if (removed) return;
        removed = true;
        toast.remove();
      };
      toast.addEventListener("animationend", remove, { once: true });
      window.setTimeout(remove, EXIT_MS);
    }
  };

  // src/constants/routes.ts
  var WEB_ROUTE = "/web";
  var SWC_API_BASE = "/web/swc";
  var Q_ACTION = "action";
  var Q_MENU_ID = "menu_id";
  var Q_VIEW_TYPE = "view_type";
  var Q_RECORD_ID = "id";
  var Q_EDIT = "edit";
  var Q_SEARCH = "q";
  var Q_SHELL = "shell";
  var Q_MODEL = "model";
  var Q_FILTER = "filter";
  var Q_SORT = "sort";
  var Q_OFFSET = "offset";
  var Q_GROUPBY = "groupby";
  var EDIT_ENABLED = "1";
  var VIEW_LIST = "list";
  var VIEW_FORM = "form";
  var VIEW_KANBAN = "kanban";
  var RECORD_UPDATED = "record.updated";
  var ACTION_CLOSED = "action.closed";
  var EXPORT_CSV_ROUTE = "/web/export/csv";
  var EXPORT_PDF_ROUTE = "/web/export/pdf";
  var BULK_TEMPLATE_ROUTE = "/web/bulk/template";
  var BULK_UPLOAD_ROUTE = "/web/bulk/upload";

  // src/model/record.ts
  var SwcRecord = class {
    model;
    id;
    data;
    dirty = /* @__PURE__ */ new Set();
    /** Client-side field domains from onchange (field name → domain). */
    fieldDomains = /* @__PURE__ */ new Map();
    /** Dynamic modifier overrides from onchange or eval. */
    modifierOverrides = /* @__PURE__ */ new Map();
    /** Optional callback after a field value changes (onchange RPC). */
    onFieldChange;
    constructor(model, id, data) {
      this.model = model;
      this.id = id;
      this.data = { ...data };
    }
    get(field) {
      return this.data[field];
    }
    set(field, value) {
      this.data[field] = value;
      this.dirty.add(field);
    }
    /** Notify listeners that a field finished editing (triggers onchange RPC). */
    notifyFieldChange(field) {
      this.onFieldChange?.(field);
    }
    isDirty() {
      return this.dirty.size > 0;
    }
    dirtyValues() {
      const out = {};
      for (const k of this.dirty) {
        out[k] = this.data[k];
      }
      return out;
    }
    clearDirty() {
      this.dirty.clear();
    }
    values() {
      return { ...this.data };
    }
  };
  var RecordStore = class {
    rpc;
    constructor(rpc) {
      this.rpc = rpc;
    }
    fromPayload(model, id, data) {
      return new SwcRecord(model, id, data);
    }
    async save(rec) {
      if (rec.id <= 0) {
        const newId = await this.rpc.create(rec.model, rec.data);
        rec.clearDirty();
        return newId;
      }
      if (!rec.isDirty()) return rec.id;
      await this.rpc.write(rec.model, [rec.id], rec.dirtyValues());
      rec.clearDirty();
      return rec.id;
    }
    async unlink(rec) {
      if (rec.id <= 0) return;
      await this.rpc.unlink(rec.model, [rec.id]);
    }
    async duplicate(rec, omit = ["id"]) {
      const values = {};
      for (const [k, v] of Object.entries(rec.data)) {
        if (omit.includes(k)) continue;
        values[k] = v;
      }
      return this.rpc.create(rec.model, values);
    }
    async applyOnchange(rec, field) {
      try {
        const result = await this.rpc.onchange(rec.model, rec.values(), field);
        if (result.value) {
          for (const [k, v] of Object.entries(result.value)) {
            rec.set(k, v);
          }
        }
        if (result.domain) {
          for (const [k, domain] of Object.entries(result.domain)) {
            rec.fieldDomains.set(k, domain);
          }
        }
        return result;
      } catch (err) {
        if (err instanceof SwcError && err.code === "rpc_error") return null;
        throw err;
      }
    }
    validate(rec, requiredFields) {
      for (const f of requiredFields) {
        const v = rec.get(f);
        if (v == null || v === "") {
          throw new SwcError(`Field ${f} is required`, "validation");
        }
      }
    }
  };

  // src/widgets/field-events.ts
  function inputValueFromEvent(event) {
    const target = event.target;
    if (target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || target instanceof HTMLSelectElement) {
      return target.value;
    }
    return "";
  }
  function checkboxCheckedFromEvent(event) {
    const target = event.target;
    return target instanceof HTMLInputElement ? target.checked : false;
  }

  // src/services/router.ts
  var RouterService = class _RouterService {
    static searchParams(route) {
      const params = new URLSearchParams();
      if (route.actionId) params.set(Q_ACTION, String(route.actionId));
      if (route.menuId) params.set(Q_MENU_ID, route.menuId);
      if (route.viewType) params.set(Q_VIEW_TYPE, route.viewType);
      if (route.recordId) params.set(Q_RECORD_ID, String(route.recordId));
      if (route.formEdit) params.set(Q_EDIT, EDIT_ENABLED);
      if (route.listSearch) params.set(Q_SEARCH, route.listSearch);
      if (route.model) params.set(Q_MODEL, route.model);
      if (route.listFilter) params.set(Q_FILTER, route.listFilter);
      if (route.listSort) params.set(Q_SORT, route.listSort);
      if (route.listOffset) params.set(Q_OFFSET, String(route.listOffset));
      if (route.listGroupBy) params.set(Q_GROUPBY, route.listGroupBy);
      if (route.shell) params.set(Q_SHELL, route.shell);
      return params;
    }
    static buildUrl(route) {
      return `${WEB_ROUTE}?${_RouterService.searchParams(route).toString()}`;
    }
    parse(location2 = window.location) {
      const q = new URLSearchParams(location2.search);
      return {
        actionId: Number(q.get(Q_ACTION) ?? "0"),
        menuId: q.get(Q_MENU_ID) ?? "",
        viewType: q.get(Q_VIEW_TYPE) ?? "",
        recordId: Number(q.get(Q_RECORD_ID) ?? "0"),
        formEdit: q.get(Q_EDIT) === EDIT_ENABLED,
        listSearch: q.get(Q_SEARCH) ?? "",
        model: q.get(Q_MODEL) ?? "",
        listFilter: q.get(Q_FILTER) ?? "",
        listSort: q.get(Q_SORT) ?? "",
        listOffset: Number(q.get(Q_OFFSET) ?? "0"),
        listGroupBy: q.get(Q_GROUPBY) ?? "",
        shell: q.get(Q_SHELL) ?? ""
      };
    }
    workspaceUrl(route) {
      return _RouterService.buildUrl({ ...this.parse(), ...route });
    }
    push(route) {
      const url = this.workspaceUrl(route);
      window.history.pushState({}, "", url);
      window.dispatchEvent(new PopStateEvent("popstate"));
    }
    /** Replace the workspace query with an absolute route (no merge with current). */
    assign(route) {
      const url = _RouterService.buildUrl(route);
      window.history.pushState({}, "", url);
      window.dispatchEvent(new PopStateEvent("popstate"));
    }
  };

  // src/views/shared/view-toolbar.ts
  function linkButton(href, label, className = "sum-btn sum-btn--secondary") {
    const a = document.createElement("a");
    a.className = className;
    a.href = href;
    a.textContent = label;
    return a;
  }
  function exportFieldNamesCsv(fields) {
    return fields.map((f) => f.name).filter(Boolean).join(",");
  }
  function newRecordUrl(payload) {
    return RouterService.buildUrl({
      actionId: payload.actionId,
      menuId: payload.menuId,
      viewType: VIEW_FORM
    });
  }
  function exportQuery(payload, fields, recordId = 0) {
    const params = new URLSearchParams();
    params.set("model", payload.model);
    if (payload.actionId > 0) params.set("action", String(payload.actionId));
    if (fields) params.set("fields", fields);
    if (recordId > 0) params.set("id", String(recordId));
    return params;
  }
  function renderSearchField(value, onSearch, onInput) {
    return html`
    <div class="sum-list-search-wrap">
      <span class="sum-list-search-icon" aria-hidden="true">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="7" />
          <path d="M20 20l-3-3" />
        </svg>
      </span>
      <input
        type="search"
        class="sum-list-search"
        placeholder="Search…"
        value=${value}
        @keydown=${(event) => event.key === "Enter" && onSearch()}
        @input=${(event) => onInput(inputValueFromEvent(event))}
      />
    </div>
  `;
  }
  function renderNewButton(payload) {
    return linkButton(newRecordUrl(payload), "New", "sum-btn sum-list-btn-new");
  }
  function renderCollectionToolbar(options) {
    const fields = exportFieldNamesCsv((options.payload.arch.fields ?? []).filter((f) => !f.invisible));
    const reportActions = renderReportActions(options.payload, fields);
    const toolbarClass = options.viewType === VIEW_KANBAN ? "sum-kanban-report-bar" : "sum-list-toolbar";
    return html`
    <div class="sum-view-toolbar ${toolbarClass}">
      <div class="sum-view-toolbar-primary">
        ${renderNewButton(options.payload)}
        ${renderSearchField(options.search, options.onSearch, options.onInput)}
        ${options.extraPrimary ?? ""}
      </div>
      ${reportActions ?? ""}
    </div>
  `;
  }
  function toolbarButton(label, className, onClick, disabled = false) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = className;
    button.textContent = label;
    button.disabled = disabled;
    button.addEventListener("click", onClick);
    return button;
  }
  function resolveHeaderButtonClass(archClass) {
    const base = "sum-header-btn";
    if (archClass?.includes("sum_highlight")) {
      return `${base} sum-header-btn--primary`;
    }
    return `${base} sum-header-btn--secondary`;
  }
  function headerButton(label, archClass, onClick, disabled = false) {
    const className = disabled ? `${resolveHeaderButtonClass(archClass)} sum-header-btn--disabled` : resolveHeaderButtonClass(archClass);
    return toolbarButton(label, className, onClick, disabled);
  }
  function renderReportActions(payload, fields, recordId = 0) {
    const report = payload.arch.report;
    if (!report?.download && !report?.upload) return null;
    const exportParams = exportQuery(payload, fields, recordId);
    const items = [];
    if (report.download) {
      items.push(linkButton(`${EXPORT_CSV_ROUTE}?${exportParams.toString()}`, "Export CSV"));
      items.push(linkButton(`${EXPORT_PDF_ROUTE}?${exportParams.toString()}`, "Export PDF"));
    }
    if (report.upload && fields) {
      const templateParams = new URLSearchParams(exportParams);
      items.push(linkButton(`${BULK_TEMPLATE_ROUTE}?${templateParams.toString()}`, "Download template"));
      items.push(
        html`<form class="sum-list-upload-form" method="post" enctype="multipart/form-data" action=${BULK_UPLOAD_ROUTE}>
        <input type="hidden" name="csrf_token" value=${payload.csrfToken} />
        <input type="hidden" name="model" value=${payload.model} />
        ${payload.actionId > 0 ? html`<input type="hidden" name="action" value=${String(payload.actionId)} />` : ""}
        <input type="hidden" name="fields" value=${fields} />
        <label class="sum-btn sum-btn--secondary sum-list-upload-label">
          Import CSV
          <input type="file" name="file" accept=".csv,text/csv" class="sum-list-upload-input" @change=${(event) => event.target.form?.requestSubmit()} />
        </label>
      </form>`
      );
    }
    if (items.length === 0) return null;
    const wrap = document.createElement("div");
    wrap.className = "sum-view-toolbar-actions";
    for (const item of items) {
      wrap.appendChild(item instanceof HTMLElement ? item : item.render());
    }
    return wrap;
  }

  // src/runtime/registry.ts
  function debugDuplicatesEnabled() {
    return typeof location !== "undefined" && /(?:^|[?&])debug=1(?:&|$)/.test(location.search);
  }
  var CategoryRegistry = class {
    constructor(map) {
      this.map = map;
    }
    map;
    add(key, value) {
      if (this.map.has(key) && debugDuplicatesEnabled()) {
        throw new Error(`Registry already has "${key}"`);
      }
      this.map.set(key, value);
    }
    get(key) {
      return this.map.get(key);
    }
    keys() {
      return [...this.map.keys()];
    }
  };
  var Registry = class {
    maps = {
      fields: /* @__PURE__ */ new Map(),
      views: /* @__PURE__ */ new Map(),
      services: /* @__PURE__ */ new Map(),
      main_components: /* @__PURE__ */ new Map()
    };
    category(name) {
      const store = this.maps[name];
      if (!store) throw new Error(`Unknown registry category: ${String(name)}`);
      return new CategoryRegistry(store);
    }
    get(category, key) {
      const store = this.maps[category];
      if (!store) throw new Error(`Unknown registry category: ${String(category)}`);
      return store.get(key);
    }
  };
  var registry = new Registry();

  // src/widgets/field-shell.ts
  function fieldInputId(field) {
    return `f-${field.name}`;
  }
  function fieldLabelId(field) {
    return `${fieldInputId(field)}-label`;
  }
  function fieldAutocomplete(field) {
    const explicit = field.options?.autocomplete;
    if (typeof explicit === "string" && explicit.trim()) {
      return explicit.trim();
    }
    const name = field.name.toLowerCase();
    const widget = (field.widget ?? "").toLowerCase();
    if (widget === "email" || name === "email") return "email";
    if (name === "phone" || name === "mobile" || name === "phone_number") return "tel";
    if (name === "name" || name === "display_name") return "organization";
    if (name === "street" || name === "street2") return "street-address";
    if (name === "city") return "address-level2";
    if (name === "zip" || name === "postal_code") return "postal-code";
    if (name === "website" || name === "url") return "url";
    if (name === "firstname" || name === "first_name") return "given-name";
    if (name === "lastname" || name === "last_name") return "family-name";
    return "off";
  }
  function isFullWidthField(field) {
    if (field.type === "text" || field.widget === "text") return true;
    if (field.type === "one2many" || field.widget === "one2many") return true;
    if (field.widget === "image") return true;
    return false;
  }
  function fieldWidgetClass(field, extra = []) {
    const parts = ["sum-field-widget"];
    if (isFullWidthField(field)) {
      parts.push("sum-field-widget--full");
    }
    if (field.type === "many2one" || field.widget === "many2one") {
      parts.push("sum-field-widget--many2one");
    }
    for (const mod of extra) {
      if (mod) parts.push(mod);
    }
    return parts.join(" ");
  }
  function fieldLabel(field, forId, row = false, labelId = fieldLabelId(field)) {
    const label = field.string ?? field.name;
    const cls = row ? "sum-field-label sum-field-label--row" : "sum-field-label";
    if (forId) {
      return html`<label class=${cls} id=${labelId} for=${forId}>${label}</label>`;
    }
    return html`<span class=${cls} id=${labelId}>${label}</span>`;
  }
  function fieldControl(body, compact = false, ariaLabelledBy) {
    const cls = compact ? "sum-field-control sum-field-control--compact" : "sum-field-control";
    if (ariaLabelledBy) {
      return html`<div class=${cls} aria-labelledby=${ariaLabelledBy}>${body}</div>`;
    }
    return html`<div class=${cls}>${body}</div>`;
  }
  function fieldPlaceholder(field) {
    return field.placeholder ?? field.string ?? field.name;
  }
  function fieldReadonlyValue(val, placeholder = "") {
    const hasValue = val.trim() !== "";
    const text = hasValue ? val : placeholder;
    const cls = hasValue ? "sum-field-value" : "sum-field-value sum-field-value--placeholder";
    return html`<div class=${cls}>${text}</div>`;
  }
  function fieldReadonlyInput(field, val, inputType = "text") {
    const placeholder = fieldPlaceholder(field);
    return html`<input
    type=${inputType}
    id=${fieldInputId(field)}
    class="sum-field-input"
    name=${field.name}
    value=${val}
    placeholder=${placeholder}
    autocomplete=${fieldAutocomplete(field)}
    readonly
    tabindex="-1"
  />`;
  }
  function renderFieldShell(field, body, options = {}) {
    const showLabel = options.showLabel !== false;
    const labelId = fieldLabelId(field);
    const labelFor = options.labelFor === false ? void 0 : options.labelFor;
    const useRow = options.layout === "row" || options.layout !== "stack" && !isFullWidthField(field) && !options.compact;
    const modifiers = [...options.modifiers ?? []];
    if (useRow) modifiers.push("sum-field-widget--row");
    const wrappedBody = fieldControl(body, options.compact === true, labelFor ? void 0 : labelId);
    if (useRow) {
      return html`<div class=${fieldWidgetClass(field, modifiers)}>
      ${showLabel ? fieldLabel(field, labelFor, true, labelId) : ""}
      ${wrappedBody}
    </div>`;
    }
    return html`<div class=${fieldWidgetClass(field, modifiers)}>
    ${showLabel ? fieldLabel(field, labelFor, false, labelId) : ""}
    ${wrappedBody}
  </div>`;
  }

  // src/widgets/field-value.ts
  function booleanFromUnknown(value) {
    return value === true || value === 1 || value === "1" || value === "true";
  }
  function stringFromUnknown(value) {
    if (value == null || value === false) return "";
    return String(value);
  }
  function recordDisplayName(record, fieldName) {
    const named = record.get(`${fieldName}_name`);
    if (named != null && named !== "") return String(named);
    const raw = record.get(fieldName);
    if (raw == null || raw === "") return "";
    return `#${raw}`;
  }

  // src/model/modifiers.ts
  function evalModifierExpr(expr, record) {
    if (!expr || !record) return void 0;
    const trimmed = expr.trim();
    if (!trimmed) return void 0;
    try {
      const ctx = { ...record.data, record: record.data };
      const fn = new Function("ctx", `with (ctx) { return !!(${trimmed}); }`);
      return Boolean(fn(ctx));
    } catch {
      return void 0;
    }
  }
  function fieldModifiers(field, record) {
    const override = record?.modifierOverrides.get(field.name);
    const dynamicInvisible = evalModifierExpr(field.invisible_expr, record);
    const dynamicReadonly = evalModifierExpr(field.readonly_expr, record);
    const dynamicRequired = evalModifierExpr(field.required_expr, record);
    return {
      invisible: override?.invisible ?? dynamicInvisible ?? field.invisible ?? false,
      readonly: override?.readonly ?? dynamicReadonly ?? field.readonly ?? false,
      required: override?.required ?? dynamicRequired ?? field.required ?? false
    };
  }
  function isFieldVisible(field, record) {
    return !fieldModifiers(field, record).invisible;
  }
  function isFieldReadonly(field, record, viewReadonly) {
    return viewReadonly || fieldModifiers(field, record).readonly;
  }
  function fieldDomain(field, record) {
    const fromRecord = record?.fieldDomains.get(field.name);
    if (fromRecord) return fromRecord;
    const raw = field.options?.domain;
    if (!raw) return void 0;
    try {
      const parsed = JSON.parse(raw);
      if (!record) return parsed;
      return evalDomainPlaceholders(parsed, record);
    } catch {
      return void 0;
    }
  }
  function evalDomainPlaceholders(domain, record) {
    return domain.map((clause) => {
      if (!Array.isArray(clause)) return clause;
      return clause.map((part) => {
        if (typeof part !== "string") return part;
        if (part.startsWith("$") && part.endsWith("$")) {
          const key = part.slice(1, -1);
          return record.get(key);
        }
        return part;
      });
    });
  }

  // src/widgets/DefaultField.ts
  function inputTypeForField(field) {
    if (field.widget === "email") return "email";
    if (field.type === "integer" || field.type === "float" || field.type === "numeric") return "number";
    if (field.type === "date") return "date";
    if (field.type === "datetime") return "datetime-local";
    return "text";
  }
  function stepForField(field) {
    if (field.type === "integer") return "1";
    if (field.type === "float" || field.type === "numeric") return "any";
    return void 0;
  }
  function parseNumericValue(field, raw) {
    if (raw === "") return null;
    if (field.type === "integer") return Number.parseInt(raw, 10);
    if (field.type === "float" || field.type === "numeric") return Number.parseFloat(raw);
    return raw;
  }
  var DefaultField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const fieldValue = stringFromUnknown(record.get(field.name));
      const placeholder = fieldPlaceholder(field);
      const inputType = inputTypeForField(field);
      const step = stepForField(field);
      const id = fieldInputId(field);
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(
          field,
          field.type === "integer" || field.type === "float" || field.type === "numeric" ? fieldReadonlyInput(field, fieldValue, "text") : fieldReadonlyInput(field, fieldValue, inputType === "text" ? "text" : inputType),
          { labelFor: id }
        );
      }
      return renderFieldShell(
        field,
        html`<input
        id=${id}
        type=${inputType}
        class="sum-field-input"
        name=${field.name}
        placeholder=${placeholder}
        value=${fieldValue}
        autocomplete=${fieldAutocomplete(field)}
        ${step ? html`step=${step}` : ""}
        @input=${(event) => record.set(field.name, parseNumericValue(field, inputValueFromEvent(event)))}
        @change=${() => record.notifyFieldChange(field.name)}
      />`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/field-async.ts
  var AsyncFieldController = class {
    constructor(component) {
      this.component = component;
    }
    component;
    generation = 0;
    begin() {
      this.generation += 1;
      return this.generation;
    }
    cancel() {
      this.generation += 1;
    }
    refresh() {
      if (this.component.rootElement?.isConnected) {
        this.component.patch();
      }
    }
    commitIfCurrent(generation) {
      if (generation !== this.generation) return;
      this.refresh();
    }
  };

  // src/widgets/Many2OneField.ts
  var Many2OneField = class extends SwcComponent {
    suggestions = [];
    open = false;
    highlightIndex = 0;
    asyncCtrl = new AsyncFieldController(this);
    onWillUnmount() {
      this.asyncCtrl.cancel();
    }
    async search(query) {
      const gen = this.asyncCtrl.begin();
      const comodel = this.props.field.relation ?? this.props.field.options?.relation ?? "";
      if (!comodel) return;
      const baseDomain = fieldDomain(this.props.field, this.props.record) ?? [];
      const domain = query ? [...baseDomain, ["name", "ilike", query]] : baseDomain;
      this.suggestions = await this.env.services.rpc.searchRead(comodel, domain, ["id", "name"], 20);
      this.open = true;
      this.highlightIndex = 0;
      this.asyncCtrl.commitIfCurrent(gen);
    }
    pick(row) {
      const { field, record } = this.props;
      record.set(field.name, row.id);
      record.set(`${field.name}_name`, row.name);
      record.notifyFieldChange(field.name);
      this.open = false;
      this.asyncCtrl.refresh();
    }
    onKeydown(event) {
      if (!this.open || this.suggestions.length === 0) return;
      if (event.key === "ArrowDown") {
        event.preventDefault();
        this.highlightIndex = (this.highlightIndex + 1) % this.suggestions.length;
        this.asyncCtrl.refresh();
      } else if (event.key === "ArrowUp") {
        event.preventDefault();
        this.highlightIndex = (this.highlightIndex - 1 + this.suggestions.length) % this.suggestions.length;
        this.asyncCtrl.refresh();
      } else if (event.key === "Enter") {
        event.preventDefault();
        const row = this.suggestions[this.highlightIndex];
        if (row) this.pick(row);
      } else if (event.key === "Escape") {
        this.open = false;
        this.asyncCtrl.refresh();
      }
    }
    template() {
      const { field, record, readonly } = this.props;
      const display = recordDisplayName(record, field.name);
      const id = fieldInputId(field);
      const placeholder = fieldPlaceholder(field);
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(field, fieldReadonlyValue(display, placeholder), { labelFor: false });
      }
      return renderFieldShell(
        field,
        html`<div class="sum-m2o-wrap">
        <input
          id=${id}
          class="sum-field-input"
          name=${field.name}
          placeholder=${placeholder}
          value=${display}
          autocomplete="off"
          @input=${(event) => void this.search(inputValueFromEvent(event))}
          @keydown=${(event) => this.onKeydown(event)}
        />
        ${this.open ? html`<ul class="sum-m2o-suggest">
              ${this.suggestions.map(
          (row, index) => html`<li>
                  <button
                    type="button"
                    class=${index === this.highlightIndex ? "sum-m2o-option sum-m2o-option--active" : "sum-m2o-option"}
                    @click=${() => this.pick(row)}
                  >
                    ${String(row.name ?? row.id)}
                  </button>
                </li>`
        )}
            </ul>` : ""}
      </div>`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/StatusbarField.ts
  function isClickable(field) {
    const opt = field.options?.clickable;
    return opt !== "0" && opt !== "false";
  }
  var StatusbarField = class extends SwcComponent {
    stages = [];
    loaded = false;
    asyncCtrl = new AsyncFieldController(this);
    setup() {
      void this.loadStages();
    }
    onWillUnmount() {
      this.asyncCtrl.cancel();
    }
    async loadStages() {
      const gen = this.asyncCtrl.begin();
      const { field } = this.props;
      if (field.selection?.length) {
        this.stages = field.selection.map(([value, label]) => ({ id: value, label }));
        this.loaded = true;
        this.asyncCtrl.commitIfCurrent(gen);
        return;
      }
      const comodel = field.relation ?? field.options?.relation ?? "";
      if (!comodel) {
        const fallback = (field.options?.states ?? "draft,done").split(",").map((s) => s.trim()).filter(Boolean);
        this.stages = fallback.map((s) => ({ id: s, label: s }));
        this.loaded = true;
        this.asyncCtrl.commitIfCurrent(gen);
        return;
      }
      const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name", "sequence"], 200);
      rows.sort((a, b) => Number(a.sequence ?? 0) - Number(b.sequence ?? 0));
      this.stages = rows.map((row) => ({
        id: Number(row.id),
        label: String(row.name ?? row.id)
      }));
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
    }
    currentId() {
      const { field, record } = this.props;
      const rawValue = record.get(field.name);
      if (rawValue == null || rawValue === "") return "";
      return field.type === "many2one" || field.relation ? Number(rawValue) : String(rawValue);
    }
    template() {
      const { field, record, readonly } = this.props;
      const current = this.currentId();
      const clickable = isClickable(field) && !isFieldReadonly(field, record, readonly);
      return html`<div class="sum-statusbar-stages" role="group" aria-label=${field.string ?? field.name}>
      ${this.stages.map((stage) => {
        const active = stage.id === current || String(stage.id) === String(current);
        const stageClass = active ? "sum-statusbar-stage sum-statusbar-stage--current" : "sum-statusbar-stage";
        return html`<button type="button" class=${stageClass} disabled=${!clickable ? "disabled" : void 0} @click=${() => {
          if (!clickable) return;
          record.set(field.name, stage.id);
          if (field.type === "many2one" || field.relation) {
            record.set(`${field.name}_name`, stage.label);
          }
          this.asyncCtrl.refresh();
        }}>${stage.label}</button>`;
      })}
      ${!this.loaded ? html`<span class="sum-statusbar-chip">Loading…</span>` : ""}
    </div>`;
    }
  };

  // src/widgets/PriorityField.ts
  function priorityMode(field) {
    const mode = (field.options?.mode ?? field.options?.display ?? "stars").toLowerCase();
    return mode === "select" || mode === "dropdown" ? "select" : "stars";
  }
  function selectionOptions(field) {
    if (!field.selection?.length) {
      return [
        { value: "0", label: "Low" },
        { value: "1", label: "Medium" },
        { value: "2", label: "High" }
      ];
    }
    return field.selection.map(([value, label]) => ({ value, label }));
  }
  function currentValue(field, record) {
    const rawValue = record.get(field.name);
    if (rawValue == null || rawValue === "") return selectionOptions(field)[0]?.value ?? "0";
    return String(rawValue);
  }
  function numericLevel(value) {
    const parsed = Number.parseInt(value, 10);
    return Number.isNaN(parsed) ? 0 : Math.max(0, parsed);
  }
  function starCount(field) {
    const fromOpt = Number(field.options?.stars ?? field.options?.max ?? 0);
    if (fromOpt > 0) return Math.min(Math.max(fromOpt, 1), 5);
    const maxLevel = selectionOptions(field).length - 1;
    return Math.max(Math.min(maxLevel, 4), 3);
  }
  var PriorityField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const options = selectionOptions(field);
      const value = currentValue(field, record);
      const mode = priorityMode(field);
      if (isFieldReadonly(field, record, readonly)) {
        const label = options.find((option) => option.value === value)?.label ?? value;
        if (mode === "select") {
          return renderFieldShell(field, fieldReadonlyValue(label), { labelFor: false });
        }
        return renderFieldShell(field, this.renderStars(numericLevel(value), true), { labelFor: false });
      }
      if (mode === "select") {
        const id = fieldInputId(field);
        return renderFieldShell(
          field,
          html`<select
          id=${id}
          class="sum-field-select sum-priority-select"
          name=${field.name}
          @change=${(event) => record.set(field.name, inputValueFromEvent(event))}
        >
          ${options.map(
            (option) => html`<option value=${option.value} selected=${value === option.value ? "selected" : ""}>
                ${option.label}
              </option>`
          )}
        </select>`,
          { labelFor: id }
        );
      }
      return renderFieldShell(
        field,
        this.renderStars(numericLevel(value), false, (level) => {
          record.set(field.name, String(level));
        }),
        { labelFor: false }
      );
    }
    starButtons(level, disabled, onPick) {
      const { field } = this.props;
      const options = selectionOptions(field);
      const count = starCount(field);
      const capped = Math.min(level, count);
      const out = [];
      for (let index = 0; index < count; index += 1) {
        const starIndex = index + 1;
        const filled = starIndex <= capped;
        const option = options[Math.min(starIndex, options.length - 1)];
        const click = () => {
          if (disabled) return;
          const next = capped === starIndex ? starIndex - 1 : starIndex;
          onPick?.(Math.max(0, next));
        };
        if (filled) {
          out.push(html`<button type="button" class="sum-priority-star sum-priority-star--on" disabled=${disabled ? "disabled" : void 0} title=${option?.label ?? `Level ${starIndex}`} aria-label=${option?.label ?? `Priority ${starIndex}`} @click=${click}>★</button>`);
        } else {
          out.push(html`<button type="button" class="sum-priority-star" disabled=${disabled ? "disabled" : void 0} title=${option?.label ?? `Level ${starIndex}`} aria-label=${option?.label ?? `Priority ${starIndex}`} @click=${click}>★</button>`);
        }
      }
      return out;
    }
    renderStars(level, disabled, onPick) {
      const { field } = this.props;
      return html`<div class="sum-priority-stars" role="group" aria-labelledby=${fieldLabelId(field)}>
      ${this.starButtons(level, disabled, onPick)}
    </div>`;
    }
  };

  // src/widgets/BooleanField.ts
  var BooleanField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const checked = booleanFromUnknown(record.get(field.name));
      const id = fieldInputId(field);
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(field, fieldReadonlyValue(checked ? "Yes" : "No"), { labelFor: false });
      }
      return renderFieldShell(
        field,
        html`<input
        id=${id}
        type="checkbox"
        class="sum-field-input"
        name=${field.name}
        autocomplete="off"
        checked=${checked ? "checked" : ""}
        @change=${(event) => record.set(field.name, checkboxCheckedFromEvent(event))}
      />`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/TextareaField.ts
  var TextareaField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const fieldValue = stringFromUnknown(record.get(field.name));
      const placeholder = fieldPlaceholder(field);
      const id = fieldInputId(field);
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(field, fieldReadonlyValue(fieldValue, placeholder), { labelFor: false });
      }
      return renderFieldShell(
        field,
        html`<textarea
        id=${id}
        class="sum-field-textarea"
        name=${field.name}
        placeholder=${placeholder}
        autocomplete=${fieldAutocomplete(field)}
        rows="5"
        @input=${(event) => record.set(field.name, inputValueFromEvent(event))}
      >${fieldValue}</textarea>`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/SelectionField.ts
  var SelectionField = class extends SwcComponent {
    options = [];
    loaded = false;
    asyncCtrl = new AsyncFieldController(this);
    setup() {
      void this.loadOptions();
    }
    onWillUnmount() {
      this.asyncCtrl.cancel();
    }
    async loadOptions() {
      const gen = this.asyncCtrl.begin();
      const { field, record, readonly } = this.props;
      if (field.selection?.length) {
        this.options = field.selection.map(([value, label]) => ({ value, label }));
        this.loaded = true;
        this.asyncCtrl.commitIfCurrent(gen);
        return;
      }
      if (isFieldReadonly(field, record, readonly)) {
        this.loaded = true;
        this.asyncCtrl.commitIfCurrent(gen);
        return;
      }
      const comodel = field.relation ?? field.options?.relation ?? "";
      if (!comodel) {
        this.loaded = true;
        this.asyncCtrl.commitIfCurrent(gen);
        return;
      }
      const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name"], 200);
      this.options = (rows ?? []).map((row) => ({
        value: String(row.id ?? ""),
        label: String(row.name ?? row.id ?? "")
      }));
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
    }
    displayValue() {
      const { field, record } = this.props;
      const rawValue = record.get(field.name);
      const id = rawValue == null || rawValue === "" ? "" : String(rawValue);
      if (!id) return "";
      const named = record.get(`${field.name}_name`);
      if (named) return String(named);
      const match = this.options.find((option) => option.value === id);
      return match?.label ?? recordDisplayName(record, field.name);
    }
    template() {
      const { field, record, readonly } = this.props;
      const current = record.get(field.name);
      const currentVal = current == null || current === "" ? "" : String(current);
      const id = fieldInputId(field);
      const placeholder = fieldPlaceholder(field);
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(field, fieldReadonlyValue(this.displayValue(), placeholder), { labelFor: false });
      }
      return renderFieldShell(
        field,
        html`<select
        id=${id}
        class="sum-field-input sum-field-select"
        name=${field.name}
        autocomplete="off"
        @change=${(event) => {
          const fieldValue = inputValueFromEvent(event);
          const option = this.options.find((o) => o.value === fieldValue);
          record.set(field.name, fieldValue ? Number(fieldValue) || fieldValue : null);
          if (option) record.set(`${field.name}_name`, option.label);
          this.asyncCtrl.refresh();
        }}
      >
        <option value="" disabled=${currentVal !== "" ? "disabled" : false} selected=${currentVal === "" ? "selected" : false}>${placeholder}</option>
        ${this.options.map(
          (option) => html`<option value=${option.value} selected=${option.value === currentVal ? "selected" : ""}>
              ${option.label}
            </option>`
        )}
      </select>
      ${!this.loaded ? html`<span class="sum-field-hint">Loading…</span>` : ""}`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/PhoneField.ts
  var PhoneField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const fieldValue = stringFromUnknown(record.get(field.name));
      const placeholder = fieldPlaceholder(field);
      const id = fieldInputId(field);
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(field, fieldReadonlyValue(fieldValue, placeholder), { labelFor: false });
      }
      return renderFieldShell(
        field,
        html`<input
        id=${id}
        type="tel"
        class="sum-field-input sum-field-phone"
        name=${field.name}
        placeholder=${placeholder}
        value=${fieldValue}
        autocomplete=${fieldAutocomplete(field)}
        @input=${(event) => record.set(field.name, inputValueFromEvent(event))}
      />`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/BooleanRadioField.ts
  var BooleanRadioField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const checked = booleanFromUnknown(record.get(field.name));
      const name = field.name;
      const fieldReadonly = isFieldReadonly(field, record, readonly);
      return renderFieldShell(
        field,
        html`<div class="sum-field-radio-group" role="radiogroup" aria-labelledby=${fieldLabelId(field)}>
        <label class="sum-field-radio">
          <input
            type="radio"
            name=${name}
            value="1"
            checked=${checked ? "checked" : ""}
            disabled=${fieldReadonly ? "disabled" : void 0}
            @change=${() => !fieldReadonly && record.set(field.name, true)}
          />
          Yes
        </label>
        <label class="sum-field-radio">
          <input
            type="radio"
            name=${name}
            value="0"
            checked=${!checked ? "checked" : ""}
            disabled=${fieldReadonly ? "disabled" : void 0}
            @change=${() => !fieldReadonly && record.set(field.name, false)}
          />
          No
        </label>
      </div>`,
        { labelFor: false }
      );
    }
  };

  // src/widgets/BooleanToggleField.ts
  var BooleanToggleField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const checked = booleanFromUnknown(record.get(field.name));
      const id = fieldInputId(field);
      return renderFieldShell(
        field,
        html`<label class="sum-field-toggle" for=${id}>
        <span class="sum-field-toggle-name">${field.string ?? field.name}</span>
        <input
          id=${id}
          type="checkbox"
          class="sum-field-input"
          name=${field.name}
          autocomplete="off"
          checked=${checked ? "checked" : ""}
          disabled=${isFieldReadonly(field, record, readonly) ? "disabled" : void 0}
          @change=${(event) => record.set(field.name, checkboxCheckedFromEvent(event))}
        />
        <span>${checked ? "On" : "Off"}</span>
      </label>`,
        { showLabel: false }
      );
    }
  };

  // src/widgets/Many2ManyTagsField.ts
  function tagIds(record, fieldName) {
    const rawValue = record.get(fieldName);
    if (!Array.isArray(rawValue)) return [];
    return rawValue.map((v) => Number(v)).filter((n) => !Number.isNaN(n));
  }
  function tagNamesFromRecord(record, fieldName) {
    const out = /* @__PURE__ */ new Map();
    const rawValue = record.get(`${fieldName}_names`);
    if (Array.isArray(rawValue)) {
      for (const item of rawValue) {
        if (item && typeof item === "object") {
          const row = item;
          const id = Number(row.id);
          if (!Number.isNaN(id)) out.set(id, String(row.name ?? id));
        }
      }
    }
    return out;
  }
  var Many2ManyTagsField = class extends SwcComponent {
    catalog = [];
    loaded = false;
    asyncCtrl = new AsyncFieldController(this);
    setup() {
      void this.loadCatalog();
    }
    onWillUnmount() {
      this.asyncCtrl.cancel();
    }
    async loadCatalog() {
      const gen = this.asyncCtrl.begin();
      const { field, record, readonly } = this.props;
      if (isFieldReadonly(field, record, readonly)) {
        this.loaded = true;
        this.asyncCtrl.commitIfCurrent(gen);
        return;
      }
      const comodel = field.relation ?? field.options?.relation ?? "";
      if (!comodel) {
        this.loaded = true;
        this.asyncCtrl.commitIfCurrent(gen);
        return;
      }
      const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name"], 500);
      this.catalog = rows.map((row) => ({ id: Number(row.id), name: String(row.name ?? row.id) }));
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
    }
    selectedTags() {
      const { field, record } = this.props;
      const ids = tagIds(record, field.name);
      const names = tagNamesFromRecord(record, field.name);
      return ids.map((id) => {
        const fromCatalog = this.catalog.find((tag) => tag.id === id);
        if (fromCatalog) return fromCatalog;
        const fromRecord = names.get(id);
        if (fromRecord) return { id, name: fromRecord };
        return { id, name: `#${id}` };
      });
    }
    setIds(ids) {
      this.props.record.set(this.props.field.name, ids);
      this.asyncCtrl.refresh();
    }
    template() {
      const { field, record, readonly } = this.props;
      const selected = this.selectedTags();
      const selectedSet = new Set(selected.map((tag) => tag.id));
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(
          field,
          html`<div class="sum-multi-select-tags sum-multi-select-tags--readonly sum-field-tags">
          ${selected.map((tag) => html`<span class="sum-multi-select-tag"><span class="sum-multi-select-tag-label">${tag.name}</span></span>`)}
        </div>`,
          { labelFor: false }
        );
      }
      const id = fieldInputId(field);
      return renderFieldShell(
        field,
        html`<div class="sum-multi-select-box">
        <div class="sum-multi-select-tags sum-field-tags">
          ${selected.map(
          (tag) => html`<span class="sum-multi-select-tag">
                <span class="sum-multi-select-tag-label">${tag.name}</span>
                <button type="button" class="sum-multi-select-remove" aria-label="Remove" @click=${() => this.setIds(selected.filter((t) => t.id !== tag.id).map((t) => t.id))}>×</button>
              </span>`
        )}
        </div>
        <select
          id=${id}
          class="sum-multi-select-add sum-field-select"
          @change=${(event) => {
          const fieldValue = Number(inputValueFromEvent(event));
          const select = event.target;
          select.value = "";
          if (!fieldValue || selectedSet.has(fieldValue)) return;
          this.setIds([...selected.map((tag) => tag.id), fieldValue]);
        }}
        >
          <option value="">Add tag…</option>
          ${this.catalog.filter((tag) => !selectedSet.has(tag.id)).map((tag) => html`<option value=${String(tag.id)}>${tag.name}</option>`)}
        </select>
        ${!this.loaded ? html`<span class="sum-field-hint">Loading…</span>` : ""}
      </div>`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/One2ManyField.ts
  var tempLineId = -1;
  function nextTempId() {
    tempLineId -= 1;
    return tempLineId;
  }
  function inverseFieldName(parentModel) {
    const part = parentModel.split(".").pop() ?? "parent";
    return `${part}_id`;
  }
  function columnsForField(field) {
    return field.subview?.fields ?? [];
  }
  function columnNames(cols) {
    return cols.map((c) => c.name);
  }
  function parseCellValue(col, raw) {
    if (raw === "") return null;
    if (col.type === "integer") return Number.parseInt(raw, 10);
    if (col.type === "float" || col.type === "numeric") return Number.parseFloat(raw);
    if (col.type === "boolean") return raw === "true" || raw === "1";
    return raw;
  }
  function displayCellValue(col, line) {
    const raw = line[col.name];
    if (raw == null) return "";
    const named = line[`${col.name}_name`];
    if (named != null && String(named) !== "") return String(named);
    if (col.type === "boolean") {
      return booleanFromUnknown(raw) ? "Yes" : "No";
    }
    return String(raw);
  }
  var One2ManyField = class extends SwcComponent {
    lines = [];
    loaded = false;
    saving = false;
    asyncCtrl = new AsyncFieldController(this);
    writeTimers = /* @__PURE__ */ new Map();
    setup() {
      void this.loadLines();
    }
    onWillUnmount() {
      this.asyncCtrl.cancel();
      for (const t of this.writeTimers.values()) clearTimeout(t);
      this.writeTimers.clear();
    }
    comodel() {
      const { field } = this.props;
      return field.relation ?? field.options?.relation ?? "";
    }
    inverse() {
      const { field, record } = this.props;
      return field.options?.inverse ?? inverseFieldName(record.model);
    }
    editable() {
      const { field, record, readonly } = this.props;
      if (isFieldReadonly(field, record, readonly)) return false;
      if (record.id <= 0) return false;
      const mode = field.subview?.editable ?? "bottom";
      return mode === "bottom" || mode === "top";
    }
    async loadLines() {
      const gen = this.asyncCtrl.begin();
      const { field, record } = this.props;
      const comodel = this.comodel();
      const cols = columnsForField(field);
      if (!comodel || record.id <= 0 || cols.length === 0) {
        this.loaded = true;
        this.asyncCtrl.commitIfCurrent(gen);
        return;
      }
      const inv = this.inverse();
      const names = ["id", ...columnNames(cols)];
      const rows = await this.env.services.rpc.searchRead(
        comodel,
        [[inv, "=", record.id]],
        names,
        200
      );
      this.lines = (rows ?? []).map((row) => ({
        id: Number(row.id ?? 0),
        data: { ...row }
      }));
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
    }
    lineById(id) {
      return this.lines.find((l) => l.id === id);
    }
    scheduleWrite(lineId, col, value) {
      const key = `${lineId}:${col.name}`;
      const prev = this.writeTimers.get(key);
      if (prev) clearTimeout(prev);
      this.writeTimers.set(
        key,
        setTimeout(() => {
          this.writeTimers.delete(key);
          void this.persistCell(lineId, col, value);
        }, 350)
      );
    }
    async persistCell(lineId, col, value) {
      if (lineId <= 0) return;
      const comodel = this.comodel();
      if (!comodel) return;
      this.saving = true;
      this.asyncCtrl.refresh();
      try {
        await this.env.services.rpc.write(comodel, [lineId], { [col.name]: value });
        const line = this.lineById(lineId);
        if (line) line.data[col.name] = value;
      } finally {
        this.saving = false;
        this.asyncCtrl.refresh();
      }
    }
    async createLine(lineId, col, value) {
      const { record } = this.props;
      const comodel = this.comodel();
      const line = this.lineById(lineId);
      if (!comodel || !line || line.id > 0) return;
      line.data[col.name] = value;
      this.saving = true;
      this.asyncCtrl.refresh();
      try {
        const vals = { ...line.data, [this.inverse()]: record.id };
        delete vals.id;
        const newId = await this.env.services.rpc.create(comodel, vals);
        line.id = newId;
        line.data.id = newId;
      } finally {
        this.saving = false;
        this.asyncCtrl.refresh();
      }
    }
    onCellInput(lineId, col, raw) {
      const value = typeof raw === "boolean" ? raw : parseCellValue(col, String(raw ?? ""));
      const line = this.lineById(lineId);
      if (!line) return;
      line.data[col.name] = value;
      if (line.id <= 0) {
        void this.createLine(lineId, col, value);
        return;
      }
      this.scheduleWrite(line.id, col, value);
    }
    async addRowViaDialog() {
      const cols = columnsForField(this.props.field);
      if (cols.length === 0) return;
      const dialog = this.env.services.dialog;
      if (dialog) {
        const ok = await dialog.confirm(
          "Add line",
          `Add a new line to ${this.props.field.string ?? this.props.field.name}?`
        );
        if (!ok) return;
      }
      this.addRow();
    }
    addRow() {
      const id = nextTempId();
      this.lines = [...this.lines, { id, data: {} }];
      this.asyncCtrl.refresh();
    }
    async deleteRow(lineId) {
      const comodel = this.comodel();
      const line = this.lineById(lineId);
      if (!line) return;
      if (line.id > 0 && comodel) {
        this.saving = true;
        this.asyncCtrl.refresh();
        try {
          await this.env.services.rpc.unlink(comodel, [line.id]);
        } finally {
          this.saving = false;
        }
      }
      this.lines = this.lines.filter((l) => l.id !== lineId);
      this.asyncCtrl.refresh();
    }
    renderCellEditor(col, line) {
      const fieldValue = String(line.data[col.name] ?? "");
      const readonly = !this.editable();
      if (readonly) {
        return html`<span>${displayCellValue(col, line.data)}</span>`;
      }
      if (col.type === "boolean") {
        const checked = booleanFromUnknown(line.data[col.name]);
        return fieldControl(
          html`<input
          type="checkbox"
          class="sum-field-input"
          checked=${checked ? "checked" : ""}
          @change=${(event) => this.onCellInput(line.id, col, checkboxCheckedFromEvent(event))}
        />`,
          true
        );
      }
      if (col.selection?.length) {
        return fieldControl(
          html`<select
          class="sum-field-select"
          @change=${(event) => this.onCellInput(line.id, col, inputValueFromEvent(event))}
        >
          <option value="">—</option>
          ${col.selection.map(
            ([v, label]) => html`<option value=${v} selected=${fieldValue === v ? "selected" : ""}>${label}</option>`
          )}
        </select>`,
          true
        );
      }
      const inputType = col.type === "integer" || col.type === "float" || col.type === "numeric" ? "number" : col.type === "date" ? "date" : "text";
      return fieldControl(
        html`<input
        type=${inputType}
        class="sum-field-input"
        value=${fieldValue}
        @input=${(event) => this.onCellInput(line.id, col, inputValueFromEvent(event))}
      />`,
        true
      );
    }
    renderLineRow(line, cols, canEdit) {
      const cells = cols.map(
        (col) => html`<td>${this.renderCellEditor(col, line)}</td>`
      );
      if (canEdit) {
        cells.push(html`<td class="sum-o2m-col-actions"><button type="button" .sum-o2m-delete-btn data-line-id=${String(line.id)} title="Remove line">×</button></td>`);
      }
      return html`<tr class="sum-o2m-row">${cells}</tr>`;
    }
    onTableClick(event) {
      const deleteButton = event.target.closest(".sum-o2m-delete-btn");
      if (!deleteButton) return;
      const id = Number(deleteButton.getAttribute("data-line-id"));
      if (!Number.isFinite(id)) return;
      void this.deleteRow(id);
    }
    template() {
      const { field, record } = this.props;
      const label = field.string ?? field.name;
      const cols = columnsForField(field);
      const canEdit = this.editable();
      const emptyMsg = !this.loaded ? "Loading\u2026" : record.id <= 0 ? "Save the record before adding lines." : cols.length === 0 ? "No columns configured." : "No lines";
      return renderFieldShell(
        field,
        html`<div class="sum-o2m-table-wrap">
        <div class="sum-o2m-title">${label}${this.saving ? " (saving\u2026)" : ""}</div>
        <table class="sum-o2m-table">
          <thead>
            <tr>
              ${cols.map((col) => html`<th>${col.string ?? col.name}</th>`)}
              ${canEdit ? html`<th class="sum-o2m-col-actions"></th>` : ""}
            </tr>
          </thead>
          <tbody @click=${(event) => this.onTableClick(event)}>
            ${this.lines.length === 0 ? html`<tr>
                  <td colspan=${String(cols.length + (canEdit ? 1 : 0))}>${emptyMsg}</td>
                </tr>` : this.lines.map((line) => this.renderLineRow(line, cols, canEdit))}
          </tbody>
        </table>
        ${canEdit && cols.length > 0 ? html`<button type="button" class="sum-o2m-add-row" @click=${() => void this.addRowViaDialog()}>
              + Add a line
            </button>` : ""}
        ${!canEdit && record.id <= 0 && !this.props.readonly ? html`<p class="sum-o2m-hint">Save the parent record before editing lines.</p>` : ""}
      </div>`,
        { layout: "stack", showLabel: false }
      );
    }
  };

  // src/widgets/DateField.ts
  function isDateTime(field) {
    return field.type === "datetime" || field.widget === "datetime";
  }
  function toNativeValue(field, raw) {
    const text = String(raw ?? "").trim();
    if (!text) return "";
    if (isDateTime(field)) {
      const d = new Date(text);
      if (Number.isNaN(d.getTime())) return text.slice(0, 16);
      const pad = (n) => String(n).padStart(2, "0");
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
    }
    return text.slice(0, 10);
  }
  function formatDisplay(field, raw) {
    const native = toNativeValue(field, raw);
    if (!native) return "";
    if (isDateTime(field)) {
      const d2 = new Date(native);
      if (Number.isNaN(d2.getTime())) return native;
      return d2.toLocaleString(void 0, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit"
      });
    }
    const d = /* @__PURE__ */ new Date(`${native}T00:00:00`);
    if (Number.isNaN(d.getTime())) return native;
    return d.toLocaleDateString(void 0, { year: "numeric", month: "short", day: "numeric" });
  }
  function todayNative(field) {
    const d = /* @__PURE__ */ new Date();
    const pad = (n) => String(n).padStart(2, "0");
    if (isDateTime(field)) {
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
    }
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  }
  var DateField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const rawValue = record.get(field.name);
      const native = toNativeValue(field, rawValue);
      const display = formatDisplay(field, rawValue);
      const placeholder = fieldPlaceholder(field);
      const id = fieldInputId(field);
      const inputType = isDateTime(field) ? "datetime-local" : "date";
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(field, fieldReadonlyInput(field, display, "text"), { labelFor: id });
      }
      return renderFieldShell(
        field,
        html`<div class="sum-date-field-inline">
        <input
          id=${id}
          type=${inputType}
          class="sum-field-input sum-date-input"
          name=${field.name}
          value=${native}
          placeholder=${placeholder}
          autocomplete="off"
          @input=${(event) => {
          record.set(field.name, inputValueFromEvent(event) || null);
          this.patch();
        }}
          @change=${(event) => {
          record.set(field.name, inputValueFromEvent(event) || null);
          record.notifyFieldChange(field.name);
        }}
        />
        <div class="sum-date-field-actions">
          <button
            type="button"
            class="sum-date-action-btn"
            @click=${() => {
          record.set(field.name, todayNative(field));
          record.notifyFieldChange(field.name);
          this.patch();
        }}
          >
            Today
          </button>
          <button
            type="button"
            class="sum-date-action-btn"
            @click=${() => {
          record.set(field.name, null);
          record.notifyFieldChange(field.name);
          this.patch();
        }}
          >
            Clear
          </button>
        </div>
      </div>`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/ImageField.ts
  var ImageField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const image = stringFromUnknown(record.get(field.name));
      const hasImage = image.length > 0;
      const id = fieldInputId(field);
      const fieldReadonly = isFieldReadonly(field, record, readonly);
      return renderFieldShell(
        field,
        html`<div data-sum-avatar>
        ${hasImage ? html`<div class="sum-image-thumb"><img class="sum-image-thumb-img" src=${image} alt="" /></div>` : html`<div class="sum-image-thumb sum-image-thumb--empty">No image</div>`}
        ${fieldReadonly ? html`<input type="hidden" data-sum-image-value name=${field.name} value=${image} />` : html`<label class="sum-form-avatar-upload">
              Upload
              <input id=${id} type="file" accept="image/*" />
              <input
                type="hidden"
                data-sum-image-value
                name=${field.name}
                value=${image}
                @input=${(event) => record.set(field.name, inputValueFromEvent(event))}
              />
            </label>`}
      </div>`,
        { modifiers: ["sum-field-widget--image"], labelFor: fieldReadonly ? false : id }
      );
    }
  };

  // src/widgets/extra-fields.ts
  var MonetaryField = class extends DefaultField {
    template() {
      const { field, record, readonly } = this.props;
      const symbol = field.options?.currency_symbol ?? "\xA4";
      const fieldValue = stringFromUnknown(record.get(field.name));
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(field, fieldReadonlyValue(fieldValue ? `${symbol} ${fieldValue}` : ""), {
          labelFor: false
        });
      }
      return super.template();
    }
  };
  var HtmlField = class extends DefaultField {
    template() {
      const { field, record, readonly } = this.props;
      const raw = stringFromUnknown(record.get(field.name));
      if (isFieldReadonly(field, record, readonly)) {
        const text = raw.replace(/<[^>]+>/g, " ").trim();
        return renderFieldShell(field, fieldReadonlyValue(text), { labelFor: false });
      }
      return super.template();
    }
  };
  var BinaryField = class extends SwcComponent {
    template() {
      const { field, record } = this.props;
      const name = stringFromUnknown(record.get(`${field.name}_name`) ?? record.get(field.name) ?? "Download");
      return renderFieldShell(
        field,
        html`<a class="sum-field-link" href="/web/content/${field.name}/${record.id}" download>${name}</a>`,
        { labelFor: false }
      );
    }
  };
  var ColorField = class extends DefaultField {
    template() {
      const { field, record, readonly } = this.props;
      const fieldValue = Number(record.get(field.name) ?? 0);
      const swatch = `hsl(${fieldValue * 47 % 360} 70% 45%)`;
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(
          field,
          html`<span class="sum-color-swatch" style=${`background:${swatch}`}></span>`,
          { labelFor: false }
        );
      }
      return super.template();
    }
  };
  var UrlField = class extends DefaultField {
    template() {
      const { field, record, readonly } = this.props;
      const fieldValue = stringFromUnknown(record.get(field.name));
      if (isFieldReadonly(field, record, readonly) && fieldValue) {
        return renderFieldShell(
          field,
          html`<a class="sum-field-link" href=${fieldValue} target="_blank" rel="noopener">${fieldValue}</a>`,
          { labelFor: false }
        );
      }
      return super.template();
    }
  };
  var ProgressField = class extends DefaultField {
    template() {
      const { field, record, readonly } = this.props;
      const fieldValue = Math.min(100, Math.max(0, Number(record.get(field.name) ?? 0)));
      if (isFieldReadonly(field, record, readonly)) {
        return renderFieldShell(
          field,
          html`<div class="sum-progress">
          <div class="sum-progress-bar" style=${`width:${fieldValue}%`}></div>
          <span>${fieldValue}%</span>
        </div>`,
          { labelFor: false }
        );
      }
      return super.template();
    }
  };
  var HandleField = class extends SwcComponent {
    template() {
      const { field } = this.props;
      return renderFieldShell(
        field,
        html`<span class="sum-handle-grip" title="Reorder" aria-hidden="true">⋮⋮</span>`,
        { labelFor: false }
      );
    }
  };

  // src/widgets/registry.ts
  var FIELD_CONSTRUCTORS = {
    default: DefaultField,
    char: DefaultField,
    email: DefaultField,
    integer: DefaultField,
    float: DefaultField,
    numeric: DefaultField,
    date: DateField,
    datetime: DateField,
    json: TextareaField,
    many2one: Many2OneField,
    one2many: One2ManyField,
    many2many: Many2ManyTagsField,
    selection: SelectionField,
    boolean: BooleanField,
    text: TextareaField,
    statusbar: StatusbarField,
    priority: PriorityField,
    phone: PhoneField,
    radio: BooleanRadioField,
    boolean_toggle: BooleanToggleField,
    many2many_tags: Many2ManyTagsField,
    image: ImageField,
    monetary: MonetaryField,
    html: HtmlField,
    binary: BinaryField,
    reference: DefaultField,
    color: ColorField,
    url: UrlField,
    progress: ProgressField,
    handle: HandleField
  };
  function registerDefaultWidgets() {
    const fields = registry.category("fields");
    for (const [key, WidgetConstructor] of Object.entries(FIELD_CONSTRUCTORS)) {
      fields.add(key, WidgetConstructor);
    }
  }
  var WIDGET_ALIASES = {
    progressbar: "progress"
  };
  var TYPE_MAP = {
    boolean: "boolean",
    text: "text",
    many2one: "many2one",
    one2many: "one2many",
    many2many: "many2many_tags",
    selection: "selection",
    date: "date",
    datetime: "datetime",
    integer: "integer",
    float: "float",
    numeric: "numeric"
  };
  function resolveFieldWidget(field) {
    const widget = field.widget ?? "";
    if (widget && widget in WIDGET_ALIASES) {
      return WIDGET_ALIASES[widget];
    }
    if (widget) return widget;
    if (field.type && field.type in TYPE_MAP) {
      return TYPE_MAP[field.type];
    }
    return field.type ?? "default";
  }
  function instantiateFieldWidget(env, field, record, readonly) {
    const key = resolveFieldWidget(field);
    const WidgetConstructor = registry.get("fields", key) ?? registry.get("fields", "default") ?? DefaultField;
    const widget = new WidgetConstructor({ field, record, readonly }, env);
    widget.callSetup();
    return widget;
  }
  function renderField(env, field, record, readonly) {
    return instantiateFieldWidget(env, field, record, readonly).render();
  }

  // src/views/form/form-sheet.ts
  function renderFields(rf, fields, record, readonly) {
    return visibleFields(fields).map((f) => rf(f, record, readonly));
  }
  function collectDivFields(div) {
    const out = [...div.fields ?? [], ...div.h1Fields ?? []];
    for (const nested of div.divs ?? []) {
      out.push(...collectDivFields(nested));
    }
    return out;
  }
  function collectFormFields(sheet, headerFields = []) {
    const out = [...headerFields];
    if (!sheet) return out.filter((f) => !f.invisible);
    out.push(...sheet.fields ?? []);
    for (const div of sheet.divs ?? []) {
      out.push(...collectDivFields(div));
    }
    for (const g of sheet.groups ?? []) {
      out.push(...collectGroupFields(g));
    }
    for (const nb of sheet.notebook ?? []) {
      for (const pg of nb.pages ?? []) {
        out.push(...pg.fields ?? []);
        for (const g of pg.groups ?? []) {
          out.push(...collectGroupFields(g));
        }
      }
    }
    return out.filter((f) => !f.invisible);
  }
  function collectGroupFields(group) {
    const out = [...group.fields ?? []];
    for (const nested of group.groups ?? []) {
      out.push(...collectGroupFields(nested));
    }
    return out;
  }
  function visibleFields(fields) {
    return fields.filter((f) => !f.invisible);
  }
  function renderSeparators(separators = []) {
    if (separators.length === 0) return html``;
    return html`${separators.map(
      (sep) => sep.string ? html`<div class="sum-separator--title">${sep.string}</div>` : html`<hr class="sum-separator--rule" />`
    )}`;
  }
  function renderLabels(labels = []) {
    if (labels.length === 0) return html``;
    return html`${labels.map((lab) => {
      const text = lab.string ?? "";
      if (lab.for) {
        return html`<label class="sum-form-label sum-form-label--hint" for=${`f-${lab.for}`}>${text}</label>`;
      }
      return html`<div class="sum-form-label sum-form-label--hint">${text}</div>`;
    })}`;
  }
  function initialsFromName(name) {
    const parts = name.trim().split(/\s+/).filter(Boolean);
    if (parts.length === 0) return "?";
    if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }
  function renderHeroField(field, record, readonly) {
    const val = String(record.get(field.name) ?? "");
    const placeholder = fieldPlaceholder(field);
    const hasValue = val.trim() !== "";
    if (isFieldReadonly(field, record, readonly)) {
      const text = hasValue ? val : placeholder;
      const cls = hasValue ? "sum-form-hero-input sum-form-hero-input--bold" : "sum-form-hero-input sum-form-hero-input--bold sum-form-hero-input--placeholder";
      return html`<h1><div class=${cls}>${text}</div></h1>`;
    }
    return html`<h1>
    <input
      id=${fieldInputId(field)}
      class="sum-form-hero-input sum-form-hero-input--bold"
      name=${field.name}
      placeholder=${placeholder}
      value=${val}
      autocomplete=${fieldAutocomplete(field)}
      aria-label=${placeholder}
      @input=${(event) => record.set(field.name, inputValueFromEvent(event))}
    />
  </h1>`;
  }
  function renderContactItem(field, record, readonly) {
    const val = String(record.get(field.name) ?? "");
    const label = field.string ?? field.name;
    const placeholder = fieldPlaceholder(field);
    const inputType = field.widget === "email" ? "email" : "text";
    if (isFieldReadonly(field, record, readonly)) {
      const text = val.trim() !== "" ? val : placeholder;
      const cls = val.trim() !== "" ? "sum-form-inline-input" : "sum-form-inline-input sum-form-inline-input--placeholder";
      return html`<div class="sum-form-contact-item">
      <label class="sum-field-label">${label}</label>
      <div class=${cls}>${text}</div>
    </div>`;
    }
    return html`<div class="sum-form-contact-item">
    <label class="sum-field-label">${label}</label>
    <input
      type=${inputType}
      class="sum-form-inline-input"
      name=${field.name}
      placeholder=${placeholder}
      value=${val}
      @input=${(event) => record.set(field.name, inputValueFromEvent(event))}
    />
  </div>`;
  }
  function renderAvatar(record, readonly) {
    const image = String(record.get("image") ?? "");
    const name = String(record.get("name") ?? "");
    const hasImage = image.length > 0;
    const initials2 = initialsFromName(name);
    return html`<div class="sum-form-avatar sum-form-avatar--compact" data-sum-avatar>
    <div class="sum-form-avatar-box sum-form-avatar-box--circle">
      ${hasImage ? html`<img
            .sum-form-avatar-img
            .sum-form-avatar-img--visible
            class=${image.includes("data:") ? "sum-form-avatar-img--cropped" : ""}
            src=${image}
            alt=""
          />` : html`<span class="sum-form-avatar-initials">${initials2}</span>`}
    </div>
    ${readonly ? "" : html`<div class="sum-form-avatar-actions">
          <input
            type="hidden"
            name="image"
            data-sum-avatar-value
            value=${image}
            @input=${(event) => record.set("image", inputValueFromEvent(event))}
          />
          <label class="sum-form-avatar-upload">
            Upload
            <input type="file" accept="image/*" />
          </label>
        </div>`}
  </div>`;
  }
  function renderTitleBody(rf, div, record, readonly) {
    const h1Fields = visibleFields(div.h1Fields ?? []);
    const contactDiv = (div.divs ?? []).find((d) => (d.class ?? "").includes("sum-title-contact-row"));
    const contactFields = visibleFields(contactDiv?.fields ?? []);
    return html`<div class="sum-form-title-body sum-form-title-body--main">
    ${h1Fields.length > 0 ? renderHeroField(h1Fields[0], record, readonly) : ""}
    ${contactFields.length > 0 ? html`<div class="sum-title-contact-row">
          ${contactFields.map((f) => renderContactItem(f, record, readonly))}
        </div>` : ""}
    ${h1Fields.length === 0 && contactFields.length === 0 ? renderFields(rf, div.fields ?? [], record, readonly) : ""}
  </div>`;
  }
  function renderTitleDiv(rf, div, record, readonly, hasImageField, onStatButton) {
    const cls = div.class ?? "";
    if (cls.includes("sum_button_box") || cls.includes("button_box")) {
      const buttons = div.buttons ?? [];
      return html`<div class="sum-form-button-box ${cls}">
      ${buttons.map(
        (archButton) => html`<button type="button" class="sum-stat-button ${archButton.class ?? ""}" data-action=${archButton.name} @click=${() => onStatButton?.(archButton.name)}>
          ${archButton.string || archButton.name}
        </button>`
      )}
    </div>`;
    }
    const isTitle = cls.includes("sum_title");
    if (!isTitle) {
      return html`<div class=${cls}>${renderFields(rf, div.fields ?? [], record, readonly)}</div>`;
    }
    if (hasImageField) {
      return html`<div class="sum-form-split-layout sum-form-split-layout--compact" data-sum-form-split>
      <aside class="sum-form-split-left sum-form-split-left--avatar">${renderAvatar(record, readonly)}</aside>
      <div class="sum-form-split-main">${renderTitleBody(rf, div, record, readonly)}</div>
    </div>`;
    }
    return html`<div class="sum-form-title-row sum-form-title-row--sheet">
    ${renderTitleBody(rf, div, record, readonly)}
  </div>`;
  }
  function outerGroupMaxCols(group, childCount) {
    if (group.col && group.col > 0) return group.col;
    return Math.max(childCount, 1);
  }
  function childGroupColspan(group) {
    return group.colspan && group.colspan > 0 ? group.colspan : 1;
  }
  function gridSpan12(maxCols, colspan) {
    const cols = Math.max(maxCols, 1);
    const span = Math.max(colspan, 1);
    return Math.min(12, Math.max(1, Math.round(span * 12 / cols)));
  }
  function packGroupRows(parent, nested) {
    const maxCols = outerGroupMaxCols(parent, nested.length);
    const rows = [];
    let current = [];
    let used = 0;
    for (const child of nested) {
      const colspan = childGroupColspan(child);
      if (used + colspan > maxCols && current.length > 0) {
        rows.push(current);
        current = [];
        used = 0;
      }
      current.push({ group: child, gridSpan: gridSpan12(maxCols, colspan) });
      used += colspan;
    }
    if (current.length > 0) rows.push(current);
    return rows;
  }
  function groupClassNames(group, ctx, plain) {
    const parts = ["sum-form-group"];
    if (plain || !group.string) {
      parts.push("sum-form-group--plain");
    } else if (ctx === "row" || ctx === "inner") {
      parts.push("sum-form-group--col");
    } else {
      parts.push("sum-form-group--full");
    }
    if ((group.fields ?? []).length > 0) {
      parts.push("sum-form-group--row-layout");
    }
    return parts.join(" ");
  }
  function renderGroup(rf, group, record, readonly, ctx = "sheet", plain = false) {
    const nested = group.groups ?? [];
    const fields = group.fields ?? [];
    const hasNested = nested.length > 0;
    if (hasNested && fields.length === 0) {
      const rows = packGroupRows(group, nested);
      return html`<div class="sum-form-group-outer sum-field-region--sheet">
      ${rows.map(
        (row) => html`<div class="sum-form-group-row">
          ${row.map(
          (item) => html`<div class="sum-form-group-span" style=${`--sum-group-span:${item.gridSpan}`}>
              ${renderGroup(rf, item.group, record, readonly, "row")}
            </div>`
        )}
        </div>`
      )}
    </div>`;
    }
    const innerCols = group.col && group.col > 0 ? group.col : 0;
    const innerColsClass = innerCols > 0 ? " sum-form-group--inner-cols" : "";
    return html`<div
    class=${groupClassNames(group, ctx, plain) + innerColsClass}
    style=${innerCols ? `--sum-inner-cols:${innerCols}` : false}
  >
    ${group.string ? html`<div class="sum-form-group-title">${group.string}</div>` : ""}
    <div class="sum-form-group-grid">
      ${renderFields(rf, fields, record, readonly)}
      ${renderSeparators(group.separators)}
      ${renderLabels(group.labels)}
      ${nested.map((g) => renderGroup(rf, g, record, readonly, "inner", true))}
    </div>
  </div>`;
  }
  function renderNotebook(rf, notebook, record, readonly, notebookIndex, activePage, onTab) {
    const pages = notebook.pages ?? [];
    if (pages.length === 0) return html``;
    const idx = Math.min(Math.max(activePage, 0), pages.length - 1);
    const page = pages[idx];
    return html`<div class="sum-notebook sum-notebook--sheet">
    <div class="sum-notebook-tabs" role="tablist">
      ${pages.map((pg, i) => {
      const tabClass = i === idx ? "sum-notebook-tab sum-notebook-tab--active" : "sum-notebook-tab";
      return html`<button type="button" class=${tabClass} role="tab" aria-selected=${i === idx ? "true" : "false"} @click=${() => onTab(notebookIndex, i)}>${pg.title}</button>`;
    })}
    </div>
    <div class="sum-notebook-page sum-notebook-page--sheet" role="tabpanel">
      <div class="sum-form-sheet-stack sum-notebook-page-body">
        ${renderFields(rf, page.fields ?? [], record, readonly)}
        ${(page.groups ?? []).map((g) => renderGroup(rf, g, record, readonly))}
        ${renderSeparators(page.separators)}
        ${renderLabels(page.labels)}
      </div>
    </div>
  </div>`;
  }
  function renderFormSheet(options) {
    const {
      env,
      sheet,
      record,
      readonly,
      hasImageField = false,
      activeNotebookPages,
      onNotebookTab,
      renderField: renderFieldOpt,
      onStatButton
    } = options;
    const rf = renderFieldOpt ?? ((f, r, ro) => renderField(env, f, r, ro));
    if (!sheet) {
      return html`<div class="sum-form-sheet"></div>`;
    }
    const parts = [];
    for (const div of sheet.divs ?? []) {
      parts.push(renderTitleDiv(rf, div, record, readonly, hasImageField, onStatButton));
    }
    const topFields = visibleFields(sheet.fields ?? []);
    const groups = sheet.groups ?? [];
    if (topFields.length > 0 || groups.length > 0) {
      parts.push(
        html`<div class="sum-form-sheet-stack sum-field-region--sheet">
        ${renderFields(rf, topFields, record, readonly)}
        ${groups.map((g) => renderGroup(rf, g, record, readonly))}
      </div>`
      );
    }
    (sheet.notebook ?? []).forEach((nb, notebookIndex) => {
      const activePage = activeNotebookPages[notebookIndex] ?? 0;
      parts.push(renderNotebook(rf, nb, record, readonly, notebookIndex, activePage, onNotebookTab));
    });
    const sheetSeparators = sheet.separators ?? [];
    const sheetLabels = sheet.labels ?? [];
    if (sheetSeparators.length > 0 || sheetLabels.length > 0) {
      parts.push(
        html`<div class="sum-form-sheet-meta">${renderSeparators(sheetSeparators)}${renderLabels(sheetLabels)}</div>`
      );
    }
    return html`<div class="sum-form-sheet">${parts}</div>`;
  }

  // src/widgets/password-match.ts
  var DEFAULT_MESSAGE = "Passwords do not match.";
  function resolveHint(confirm, hint) {
    if (hint) {
      return hint;
    }
    const el = document.createElement("p");
    el.className = "sum-field-hint";
    el.setAttribute("role", "alert");
    el.setAttribute("aria-live", "polite");
    el.hidden = true;
    const field = confirm.closest(".field, .sum-field-widget");
    if (field) {
      field.appendChild(el);
    } else {
      confirm.insertAdjacentElement("afterend", el);
    }
    return el;
  }
  function showMismatch(password, confirm, hint, message) {
    hint.textContent = message;
    hint.hidden = false;
    hint.classList.add("sum-field-hint--error");
    password.classList.add("sum-input-invalid");
    confirm.classList.add("sum-input-invalid");
    confirm.setAttribute("aria-invalid", "true");
  }
  function clearMismatch(password, confirm, hint) {
    hint.textContent = "";
    hint.hidden = true;
    hint.classList.remove("sum-field-hint--error");
    password.classList.remove("sum-input-invalid");
    confirm.classList.remove("sum-input-invalid");
    confirm.removeAttribute("aria-invalid");
  }
  function bindPasswordMatch(options) {
    const { password, confirm, message = DEFAULT_MESSAGE } = options;
    const hint = resolveHint(confirm, options.hint);
    const evaluate = () => {
      if (!confirm.value && !password.value) {
        clearMismatch(password, confirm, hint);
        return true;
      }
      if (password.value !== confirm.value) {
        showMismatch(password, confirm, hint, message);
        return false;
      }
      clearMismatch(password, confirm, hint);
      return true;
    };
    const onInput = () => {
      evaluate();
    };
    password.addEventListener("input", onInput);
    confirm.addEventListener("input", onInput);
    password.addEventListener("blur", onInput);
    confirm.addEventListener("blur", onInput);
    return {
      isValid: evaluate,
      destroy: () => {
        password.removeEventListener("input", onInput);
        confirm.removeEventListener("input", onInput);
        password.removeEventListener("blur", onInput);
        confirm.removeEventListener("blur", onInput);
      }
    };
  }
  function validatePasswordMatchGroups(root = document) {
    let ok = true;
    root.querySelectorAll("[data-password-match]").forEach((group) => {
      const password = group.querySelector("[data-password-primary]");
      const confirm = group.querySelector("[data-password-confirm]");
      if (!password || !confirm) {
        return;
      }
      if (!password.value && !confirm.value) {
        return;
      }
      if (password.value !== confirm.value) {
        const hint = group.querySelector("[data-password-match-hint]");
        if (hint) {
          showMismatch(password, confirm, hint, DEFAULT_MESSAGE);
        }
        ok = false;
      }
    });
    return ok;
  }
  function initPasswordMatchGroups(root = document) {
    root.querySelectorAll("[data-password-match]").forEach((group) => {
      if (group.dataset.passwordMatchBound === "1") {
        return;
      }
      const password = group.querySelector("[data-password-primary]");
      const confirm = group.querySelector("[data-password-confirm]");
      if (!password || !confirm) {
        return;
      }
      const hint = group.querySelector("[data-password-match-hint]");
      bindPasswordMatch({ password, confirm, hint });
      group.dataset.passwordMatchBound = "1";
    });
  }

  // src/views/form/form-interactions.ts
  function onNotebookKeydown(ev) {
    if (!(ev instanceof KeyboardEvent)) return;
    if (ev.key !== "ArrowLeft" && ev.key !== "ArrowRight") return;
    const target = ev.target;
    if (!(target instanceof HTMLButtonElement) || target.getAttribute("role") !== "tab") return;
    const tabs = target.parentElement;
    if (!tabs) return;
    const buttons = Array.from(tabs.querySelectorAll('button[role="tab"]'));
    const idx = buttons.indexOf(target);
    if (idx < 0) return;
    ev.preventDefault();
    const next = ev.key === "ArrowRight" ? Math.min(idx + 1, buttons.length - 1) : Math.max(idx - 1, 0);
    buttons[next]?.focus();
    buttons[next]?.click();
  }
  function bindMany2OneDismiss(root) {
    const onDocClick = (ev) => {
      const target = ev.target;
      if (!(target instanceof Node)) return;
      for (const widget of root.querySelectorAll(".sum-field-widget--many2one")) {
        if (widget.contains(target)) continue;
        const list = widget.querySelector(".sum-m2o-suggest");
        list?.remove();
      }
    };
    const onKey = (ev) => {
      if (ev.key !== "Escape") return;
      for (const list of root.querySelectorAll(".sum-m2o-suggest")) {
        list.remove();
      }
    };
    document.addEventListener("click", onDocClick, true);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("click", onDocClick, true);
      document.removeEventListener("keydown", onKey);
    };
  }
  function normalizeCrop(c) {
    return {
      x: Math.min(100, Math.max(0, c.x)),
      y: Math.min(100, Math.max(0, c.y)),
      zoom: Math.min(4, Math.max(1, c.zoom))
    };
  }
  function applyCropStyle(img, crop) {
    const c = normalizeCrop(crop);
    img.style.objectPosition = `${c.x}% ${c.y}%`;
    img.style.transform = `scale(${c.zoom})`;
    img.style.transformOrigin = `${c.x}% ${c.y}%`;
  }
  function openAvatarCropModal(file, onSave) {
    const modal = document.createElement("div");
    modal.className = "sum-avatar-crop-modal";
    modal.innerHTML = `
    <div class="sum-avatar-crop-modal-inner">
      <h3 class="sum-avatar-crop-title">Crop image</h3>
      <p class="sum-avatar-crop-hint">Drag to reposition \xB7 use zoom slider</p>
      <div class="sum-avatar-crop-stage">
        <div class="sum-avatar-crop-viewport">
          <img class="sum-avatar-crop-img" alt="" />
          <div class="sum-avatar-crop-ring"></div>
        </div>
      </div>
      <label class="sum-avatar-crop-zoom-label">Zoom
        <input type="range" class="sum-avatar-crop-zoom" min="1" max="4" step="0.05" value="1" />
      </label>
      <div class="sum-avatar-crop-modal-actions">
        <button type="button" class="sum-avatar-crop-save">Save</button>
        <button type="button" class="sum-avatar-crop-cancel">Cancel</button>
      </div>
    </div>`;
    const img = modal.querySelector(".sum-avatar-crop-img");
    const zoom = modal.querySelector(".sum-avatar-crop-zoom");
    const stage = modal.querySelector(".sum-avatar-crop-stage");
    let crop = { x: 50, y: 50, zoom: 1 };
    let dragging = false;
    const close = () => modal.remove();
    const reader = new FileReader();
    reader.onload = () => {
      img.src = String(reader.result ?? "");
      applyCropStyle(img, crop);
    };
    reader.readAsDataURL(file);
    stage.addEventListener("pointerdown", (ev) => {
      dragging = true;
      stage.setPointerCapture(ev.pointerId);
    });
    stage.addEventListener("pointermove", (ev) => {
      if (!dragging) return;
      const rect = stage.getBoundingClientRect();
      crop.x = (ev.clientX - rect.left) / rect.width * 100;
      crop.y = (ev.clientY - rect.top) / rect.height * 100;
      applyCropStyle(img, crop);
    });
    stage.addEventListener("pointerup", () => {
      dragging = false;
    });
    zoom.addEventListener("input", () => {
      crop.zoom = Number(zoom.value);
      applyCropStyle(img, crop);
    });
    modal.querySelector(".sum-avatar-crop-cancel")?.addEventListener("click", close);
    modal.querySelector(".sum-avatar-crop-save")?.addEventListener("click", () => {
      const canvas = document.createElement("canvas");
      const size = 256;
      canvas.width = size;
      canvas.height = size;
      const ctx = canvas.getContext("2d");
      if (ctx && img.complete) {
        const c = normalizeCrop(crop);
        const sw = img.naturalWidth / c.zoom;
        const sh = img.naturalHeight / c.zoom;
        const sx = c.x / 100 * img.naturalWidth - sw / 2;
        const sy = c.y / 100 * img.naturalHeight - sh / 2;
        ctx.drawImage(img, sx, sy, sw, sh, 0, 0, size, size);
        onSave(canvas.toDataURL("image/png"), c);
      } else {
        onSave(img.src, normalizeCrop(crop));
      }
      close();
    });
    document.body.appendChild(modal);
  }
  function bindAvatarUpload(root) {
    const onChange = (ev) => {
      const input = ev.target;
      if (!(input instanceof HTMLInputElement) || input.type !== "file") return;
      const file = input.files?.[0];
      if (!file || !file.type.startsWith("image/")) return;
      const host = input.closest("[data-sum-avatar]") ?? input.closest(".sum-field-widget--image");
      if (!host) return;
      const hidden = host.querySelector("[data-sum-avatar-value], [data-sum-image-value]");
      openAvatarCropModal(file, (dataUrl) => {
        if (hidden) hidden.value = dataUrl;
        hidden?.dispatchEvent(new Event("input", { bubbles: true }));
        const img = host.querySelector(".sum-form-avatar-img, .sum-image-thumb-img");
        if (img) {
          img.src = dataUrl;
          img.classList.add("sum-form-avatar-img--visible", "sum-form-avatar-img--cropped");
        }
        const initials2 = host.querySelector(".sum-form-avatar-initials");
        initials2?.remove();
      });
      input.value = "";
    };
    root.addEventListener("change", onChange);
    return () => root.removeEventListener("change", onChange);
  }
  function bindDateDismiss(root) {
    const onDocClick = (ev) => {
      const target = ev.target;
      if (!(target instanceof Node)) return;
      for (const details of root.querySelectorAll("details.sum-date-field[open]")) {
        if (details.contains(target)) continue;
        details.open = false;
      }
    };
    document.addEventListener("click", onDocClick, true);
    return () => document.removeEventListener("click", onDocClick, true);
  }
  function initFormInteractions(root) {
    const cleanups = [];
    for (const tabs of root.querySelectorAll(".sum-notebook-tabs")) {
      tabs.addEventListener("keydown", onNotebookKeydown);
      cleanups.push(() => tabs.removeEventListener("keydown", onNotebookKeydown));
    }
    cleanups.push(bindMany2OneDismiss(root));
    cleanups.push(bindAvatarUpload(root));
    cleanups.push(bindDateDismiss(root));
    initPasswordMatchGroups(root);
    return () => {
      for (const fn of cleanups) fn();
    };
  }

  // src/widgets/field-host.ts
  var FieldHost = class {
    env;
    entries = /* @__PURE__ */ new Map();
    constructor(env) {
      this.env = env;
    }
    render(field, record, readonly) {
      const widgetName = resolveFieldWidget(field);
      const key = `${record.id}:${field.name}`;
      const prev = this.entries.get(key);
      if (prev && prev.readonly === readonly && prev.widgetName === widgetName) {
        return prev.widget.renderOrPatch();
      }
      prev?.widget.destroy();
      const widget = instantiateFieldWidget(this.env, field, record, readonly);
      this.entries.set(key, { widget, readonly, widgetName });
      return widget.render();
    }
    /** Drop one field widget after onchange, or all widgets when `fieldName` is omitted. */
    invalidate(fieldName) {
      if (!fieldName) {
        this.clear();
        return;
      }
      const suffix = `:${fieldName}`;
      for (const [key, entry] of [...this.entries]) {
        if (key === fieldName || key.endsWith(suffix)) {
          entry.widget.destroy();
          this.entries.delete(key);
        }
      }
    }
    clear() {
      for (const { widget } of this.entries.values()) {
        widget.destroy();
      }
      this.entries.clear();
    }
  };

  // src/views/chatter/ChatterPanel.ts
  var ChatterPanel = class extends SwcComponent {
    messages = [];
    attachments = [];
    draft = "";
    loading = true;
    posting = false;
    enabled = true;
    tab = "messages";
    setup() {
      void this.load();
    }
    async load() {
      const { model, recordId } = this.props;
      if (recordId <= 0) {
        this.loading = false;
        this.rerender();
        return;
      }
      this.loading = true;
      this.rerender();
      try {
        const base = this.env.bootstrap.swcApiBase || SWC_API_BASE;
        const data = await this.env.services.http.getJSON(
          `${base}/chatter?model=${encodeURIComponent(model)}&id=${recordId}`
        );
        this.messages = data.messages ?? [];
        this.attachments = data.attachments ?? [];
        this.enabled = data.enabled !== false;
      } finally {
        this.loading = false;
        this.rerender();
      }
    }
    async post() {
      const body = this.draft.trim();
      if (!body || this.props.recordId <= 0) return;
      this.posting = true;
      this.rerender();
      try {
        await this.env.services.http.postForm("/web/chatter/post", {
          model: this.props.model,
          res_id: String(this.props.recordId),
          body,
          next: window.location.pathname + window.location.search
        });
        this.draft = "";
        await this.load();
        this.env.services.bus.emit(RECORD_UPDATED, {
          model: this.props.model,
          id: this.props.recordId
        });
      } finally {
        this.posting = false;
        this.rerender();
      }
    }
    template() {
      if (this.props.recordId <= 0) {
        return html`<aside class="sum-chatter sum-chatter--empty">Save the record to post messages.</aside>`;
      }
      if (!this.enabled) return html``;
      if (this.loading) {
        return html`<aside class="sum-chatter sum-chatter--loading">Loading messages…</aside>`;
      }
      return html`
      <aside class="sum-chatter">
        <div class="sum-chatter-tabs">
          <button type="button" class="sum-chatter-tab${this.tab === "messages" ? " sum-chatter-tab--active" : ""}" @click=${() => {
        this.tab = "messages";
        this.rerender();
      }}>Messages</button>
          <button type="button" class="sum-chatter-tab${this.tab === "attachments" ? " sum-chatter-tab--active" : ""}" @click=${() => {
        this.tab = "attachments";
        this.rerender();
      }}>Attachments (${this.attachments.length})</button>
        </div>
        ${this.tab === "attachments" ? html`<ul class="sum-chatter-attachments">
              ${this.attachments.length === 0 ? html`<li class="sum-chatter-empty">No attachments.</li>` : this.attachments.map(
        (a) => html`<li><a href=${a.url} target="_blank" rel="noopener">${a.name}</a></li>`
      )}
            </ul>` : html`
        <div class="sum-chatter-composer">
          <textarea
            class="sum-chatter-input"
            placeholder="Write a message…"
            rows="3"
            value=${this.draft}
            @input=${(event) => {
        this.draft = inputValueFromEvent(event);
        this.rerender();
      }}
          ></textarea>
          <button
            type="button"
            class="sum-btn sum-btn--primary sum-chatter-send"
            disabled=${this.posting ? "disabled" : void 0}
            @click=${() => void this.post()}
          >
            Post
          </button>
        </div>
        <ul class="sum-chatter-messages">
          ${this.messages.length === 0 ? html`<li class="sum-chatter-empty">No messages yet.</li>` : this.messages.map(
        (m) => html`<li class="sum-chatter-message">
                  <div class="sum-chatter-meta">${m.author} · ${m.createDate}</div>
                  <div class="sum-chatter-body">${m.body}</div>
                </li>`
      )}
        </ul>`}
      </aside>
    `;
    }
  };

  // src/views/shared/object-action.ts
  async function runObjectAction(env, options) {
    try {
      const result = await env.services.rpc.callMethod(
        options.model,
        options.methodName,
        options.recordId,
        options.extraArgs
      );
      if (await env.services.action.applyCallResult(result)) {
        return true;
      }
      env.services.notification.success(options.buttonLabel, "Action completed.");
      await options.onSuccess?.();
      return false;
    } catch (error) {
      env.services.notification.error(
        options.buttonLabel,
        error instanceof SwcError ? error.message : String(error)
      );
      return false;
    }
  }

  // src/views/form/FormView.ts
  var FormView = class extends SwcComponent {
    recordStore;
    record;
    snapshot = {};
    editing = false;
    saving = false;
    acting = false;
    error = "";
    activeNotebookPages = {};
    teardownInteractions = null;
    fieldHost;
    chatterPanel;
    setup() {
      this.recordStore = new RecordStore(this.env.services.rpc);
      this.fieldHost = new FieldHost(this.env);
      this.initRecordState(this.props.payload);
      this.chatterPanel = new ChatterPanel(
        {
          model: this.props.payload.model,
          recordId: this.props.payload.recordId,
          csrfToken: this.props.payload.csrfToken
        },
        this.env
      );
      this.chatterPanel.callSetup();
    }
    onPropsChanged(props) {
      this.initRecordState(props.payload);
      this.chatterPanel.updateProps({
        model: props.payload.model,
        recordId: props.payload.recordId,
        csrfToken: props.payload.csrfToken
      });
      this.fieldHost.clear();
    }
    initRecordState(payload) {
      this.editing = payload.formEdit || payload.recordId <= 0;
      this.snapshot = { ...payload.record ?? {} };
      this.bindRecord(this.recordStore.fromPayload(payload.model, payload.recordId, this.snapshot));
    }
    onMount() {
      this.bindFormInteractions();
    }
    onWillUnmount() {
      this.teardownInteractions?.();
      this.teardownInteractions = null;
      this.fieldHost.clear();
      this.chatterPanel.destroy();
    }
    patch() {
      this.teardownInteractions?.();
      super.patch();
    }
    afterPatch() {
      this.bindFormInteractions();
    }
    bindFormInteractions() {
      if (this.rootElement) {
        this.teardownInteractions = initFormInteractions(this.rootElement);
      }
    }
    bindRecord(record) {
      this.record = record;
      this.record.onFieldChange = (field) => void this.handleFieldChange(field);
    }
    async handleFieldChange(field) {
      if (this.isReadonly()) return;
      const result = await this.recordStore.applyOnchange(this.record, field);
      if (result?.warning) {
        this.env.services.notification.warning(result.warning.title, result.warning.message);
      }
      this.fieldHost.invalidate(field);
      this.rerender();
    }
    renderFieldCached = (field, record, readonly) => {
      if (!isFieldVisible(field, record)) {
        const element = document.createElement("div");
        element.hidden = true;
        return element;
      }
      return this.fieldHost.render(field, record, readonly);
    };
    isReadonly() {
      return !this.editing;
    }
    toolbarBusy() {
      return this.saving || this.acting;
    }
    fields() {
      const arch = this.props.payload.arch;
      return collectFormFields(arch.sheet, arch.header?.fields ?? []);
    }
    headerButtons() {
      return this.props.payload.arch.header?.buttons ?? [];
    }
    startEdit() {
      this.editing = true;
      this.error = "";
      this.rerender();
    }
    cancelEdit() {
      const payload = this.props.payload;
      if (payload.recordId <= 0) {
        const url = this.env.services.router.workspaceUrl({
          actionId: payload.actionId,
          menuId: payload.menuId,
          viewType: VIEW_LIST,
          recordId: 0,
          formEdit: false
        });
        this.env.services.action.navigate(url);
        return;
      }
      this.bindRecord(this.recordStore.fromPayload(payload.model, payload.recordId, { ...this.snapshot }));
      this.editing = false;
      this.error = "";
      this.rerender();
    }
    async reloadRecord() {
      const payload = this.props.payload;
      if (payload.recordId <= 0) return;
      const fieldNames = this.fields().map((f) => f.name);
      const rows = await this.env.services.rpc.read(payload.model, [payload.recordId], fieldNames);
      if (!rows[0]) return;
      this.snapshot = { ...rows[0] };
      this.bindRecord(this.recordStore.fromPayload(payload.model, payload.recordId, this.snapshot));
      this.rerender();
    }
    async save() {
      if (this.rootElement && !validatePasswordMatchGroups(this.rootElement)) {
        this.error = "Passwords do not match.";
        this.rerender();
        return;
      }
      this.saving = true;
      this.error = "";
      this.rerender();
      try {
        const required = this.fields().filter((f) => f.required).map((f) => f.name);
        this.recordStore.validate(this.record, required);
        const id = await this.recordStore.save(this.record);
        this.env.services.notification.success("Saved", "Record saved successfully.");
        const payload = this.props.payload;
        if (payload.recordId <= 0 && id > 0) {
          this.env.services.action.openRecord({
            actionId: payload.actionId,
            menuId: payload.menuId,
            recordId: id,
            viewType: VIEW_FORM
          });
          return;
        }
        this.snapshot = { ...this.record.data };
        this.editing = false;
        this.rerender();
      } catch (err) {
        const message = err instanceof SwcError ? err.message : String(err);
        if (err instanceof SwcError && err.code === "validation") {
          this.error = message;
        } else {
          this.env.services.notification.error("Save failed", message);
        }
      } finally {
        this.saving = false;
        this.rerender();
      }
    }
    async deleteRecord() {
      const payload = this.props.payload;
      if (payload.recordId <= 0) return;
      const ok = await this.env.services.dialog.confirm("Delete record", "This cannot be undone.");
      if (!ok) return;
      try {
        await this.recordStore.unlink(this.record);
        this.env.services.notification.success("Deleted", "Record deleted.");
        this.env.services.action.navigate(
          this.env.services.router.workspaceUrl({
            actionId: payload.actionId,
            menuId: payload.menuId,
            viewType: VIEW_LIST,
            recordId: 0
          })
        );
      } catch (err) {
        this.env.services.notification.error(
          "Delete failed",
          err instanceof SwcError ? err.message : String(err)
        );
      }
    }
    async duplicateRecord() {
      const payload = this.props.payload;
      if (payload.recordId <= 0) return;
      try {
        const newId = await this.recordStore.duplicate(this.record);
        this.env.services.notification.success("Duplicated", "Record duplicated.");
        this.env.services.action.openRecord({
          actionId: payload.actionId,
          menuId: payload.menuId,
          recordId: newId,
          viewType: VIEW_FORM
        });
      } catch (err) {
        this.env.services.notification.error(
          "Duplicate failed",
          err instanceof SwcError ? err.message : String(err)
        );
      }
    }
    async runObjectButton(archButton) {
      const payload = this.props.payload;
      if (archButton.type !== "object" || payload.recordId <= 0) return;
      this.acting = true;
      this.error = "";
      this.rerender();
      const navigated = await runObjectAction(this.env, {
        model: payload.model,
        methodName: archButton.name,
        recordId: payload.recordId,
        buttonLabel: archButton.string || archButton.name,
        onSuccess: () => this.reloadRecord()
      });
      this.acting = false;
      if (!navigated) this.rerender();
    }
    renderToolbarPrimary() {
      const payload = this.props.payload;
      const busy = this.toolbarBusy();
      const items = [];
      if (payload.recordId > 0 && this.isReadonly()) {
        if (!this.props.inDialog) {
          items.push(renderNewButton(payload));
          items.push(headerButton("Edit", void 0, () => this.startEdit(), busy));
          items.push(headerButton("Duplicate", void 0, () => void this.duplicateRecord(), busy));
          items.push(
            headerButton("Delete", "sum-btn--danger", () => void this.deleteRecord(), busy)
          );
        } else {
          items.push(headerButton("Edit", void 0, () => this.startEdit(), busy));
        }
      } else {
        items.push(headerButton("Save", "sum_highlight", () => void this.save(), busy));
        items.push(headerButton("Cancel", void 0, () => this.cancelEdit(), busy || this.saving));
      }
      for (const archButton of this.headerButtons()) {
        if (archButton.type !== "object") continue;
        items.push(
          headerButton(
            archButton.string || archButton.name,
            archButton.class,
            () => void this.runObjectButton(archButton),
            busy
          )
        );
      }
      return items;
    }
    template() {
      const payload = this.props.payload;
      const readonly = this.isReadonly();
      const headerFields = payload.arch.header?.fields ?? [];
      const exportFields = exportFieldNamesCsv(this.fields());
      const reportActions = payload.recordId > 0 ? renderReportActions(payload, exportFields, payload.recordId) : null;
      const toolbarItems = this.renderToolbarPrimary();
      const busy = this.toolbarBusy();
      const sheet = renderFormSheet({
        env: this.env,
        sheet: payload.arch.sheet,
        record: this.record,
        readonly,
        hasImageField: payload.arch.formMeta?.hasImageField ?? false,
        activeNotebookPages: this.activeNotebookPages,
        onNotebookTab: (notebookIndex, pageIndex) => {
          this.activeNotebookPages = { ...this.activeNotebookPages, [notebookIndex]: pageIndex };
          this.rerender();
        },
        renderField: this.renderFieldCached,
        onStatButton: (name) => void this.runObjectButton({ name, string: name, type: "object" })
      });
      const footerButtons = payload.arch.footer?.buttons ?? [];
      const showChatter = payload.arch.hasChatter && payload.recordId > 0;
      return html`
      <div class="sum-form-view sum-form-view--workspace-chrome${readonly ? " sum-form-view--readonly" : ""}">
        <div class="sum-ws-record-toolbar sum-view-toolbar sum-form-toolbar">
          <div class="sum-statusbar-buttons sum-view-toolbar-primary">${toolbarItems}</div>
          ${headerFields.length > 0 ? html`<div class="sum-statusbar-status sum-ws-toolbar-right">
                ${headerFields.map((f) => this.renderFieldCached(f, this.record, readonly))}
              </div>` : ""}
          ${reportActions ?? ""}
        </div>
        ${this.error ? html`<div class="sum-flash sum-flash--error">${this.error}</div>` : ""}
        <div class="sum-form-layout${showChatter ? " sum-form-layout--with-chatter" : ""}">
          <div class="sum-form-sheet-bg">
            ${sheet}
            ${footerButtons.length > 0 ? html`<div class="sum-form-footer">
                  ${footerButtons.map(
        (archButton) => headerButton(
          archButton.string || archButton.name,
          archButton.class,
          () => void this.runObjectButton(archButton),
          busy
        )
      )}
                </div>` : ""}
          </div>
          ${showChatter ? this.chatterPanel.render() : ""}
        </div>
      </div>
    `;
    }
  };

  // src/services/action.ts
  var ActionService = class {
    constructor(router) {
      this.router = router;
    }
    router;
    env;
    dialogView = null;
    setEnv(env) {
      this.env = env;
    }
    navigate(url) {
      if ((url.startsWith(`${WEB_ROUTE}?`) || url === WEB_ROUTE) && this.router) {
        const parsed = new URL(url, window.location.origin);
        this.router.assign(this.router.parse({ search: parsed.search }));
        return;
      }
      window.location.assign(url);
    }
    openWindowAction(actionId, menuId, extra) {
      const params = new URLSearchParams({ [Q_ACTION]: String(actionId) });
      if (menuId) params.set(Q_MENU_ID, menuId);
      for (const [k, v] of Object.entries(extra ?? {})) {
        if (v) params.set(k, v);
      }
      this.navigate(`${WEB_ROUTE}?${params.toString()}`);
    }
    openRecord({ actionId, menuId, recordId, viewType = VIEW_FORM }) {
      const route = {
        actionId,
        menuId,
        recordId,
        viewType,
        formEdit: false,
        listSearch: ""
      };
      if (this.router) {
        this.router.push(route);
        return;
      }
      this.navigate(RouterService.buildUrl(route));
    }
    /** Apply an object-action RPC result. Returns true when navigation or a dialog handled it. */
    async applyCallResult(result) {
      if (result === true || result === false || result == null) return false;
      const body = result;
      if (body.close) {
        this.closeDialog();
        this.env?.services.bus.emit(ACTION_CLOSED, {});
        return true;
      }
      if (body.open) {
        const target = body.open.target || "dialog";
        if (target === "dialog") {
          await this.openFormDialog(body.open);
          return true;
        }
        this.navigateCurrent(body.open);
        return true;
      }
      if (body.redirect) {
        const parsed = this.parseRedirectOpen(body.redirect);
        if (parsed) {
          return this.applyCallResult({ open: parsed });
        }
        this.navigate(body.redirect);
        return true;
      }
      return false;
    }
    navigateCurrent(open) {
      this.navigate(
        RouterService.buildUrl({
          actionId: open.actionId ?? 0,
          viewType: open.viewType || VIEW_FORM,
          recordId: open.recordId ?? 0,
          model: open.model,
          formEdit: false,
          listSearch: "",
          menuId: ""
        })
      );
    }
    parseRedirectOpen(redirect) {
      if (!redirect.startsWith(`${WEB_ROUTE}?`) && !redirect.startsWith("?")) return null;
      const q = new URLSearchParams(redirect.slice(redirect.indexOf("?") + 1));
      const model = q.get("model") ?? "";
      const recordId = Number(q.get("id") ?? "0");
      if (!model || recordId <= 0) return null;
      return {
        model,
        actionId: Number(q.get(Q_ACTION) ?? "0") || void 0,
        viewType: q.get(Q_VIEW_TYPE) || VIEW_FORM,
        recordId,
        target: q.get("target") || "dialog"
      };
    }
    async openFormDialog(open) {
      const env = this.env;
      if (!env) {
        this.navigateCurrent({ ...open, target: "current" });
        return;
      }
      const params = new URLSearchParams();
      if (open.model) params.set("model", open.model);
      if (open.actionId) params.set(Q_ACTION, String(open.actionId));
      if (open.recordId) params.set("id", String(open.recordId));
      params.set(Q_VIEW_TYPE, open.viewType || VIEW_FORM);
      params.set(Q_EDIT, EDIT_ENABLED);
      const base = env.bootstrap.swcApiBase || SWC_API_BASE;
      const payload = await env.services.http.getJSON(
        `${base}/workspace?${params.toString()}`
      );
      this.closeDialog();
      const view = new FormView({ payload, inDialog: true }, env);
      view.callSetup();
      this.dialogView = view;
      const title = payload.arch.title || payload.arch.model || "Wizard";
      void env.services.dialog.openHost(title, view.render()).then(() => {
        view.destroy();
        if (this.dialogView === view) this.dialogView = null;
      });
    }
    closeDialog() {
      this.env?.services.dialog.close();
    }
  };

  // src/services/bus.ts
  var BusService = class {
    handlers = /* @__PURE__ */ new Map();
    ws = null;
    subscribe(channel, handler) {
      if (!this.handlers.has(channel)) {
        this.handlers.set(channel, /* @__PURE__ */ new Set());
      }
      this.handlers.get(channel).add(handler);
      return () => this.handlers.get(channel)?.delete(handler);
    }
    emit(channel, payload) {
      for (const fn of this.handlers.get(channel) ?? []) {
        fn(payload);
      }
    }
    /** Connect to /web/swc/bus when bootstrap.busEnabled is true. */
    connect(url = `${SWC_API_BASE}/bus`) {
      if (this.ws) return;
      try {
        const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
        this.ws = new WebSocket(`${proto}//${window.location.host}${url}`);
        this.ws.addEventListener("message", (ev) => {
          try {
            const msg = JSON.parse(String(ev.data));
            if (msg.channel) this.emit(msg.channel, msg.payload);
          } catch {
          }
        });
        this.ws.addEventListener("close", () => {
          this.ws = null;
        });
      } catch {
      }
    }
    disconnect() {
      this.ws?.close();
      this.ws = null;
    }
  };

  // src/services/dialog.ts
  var DialogService = class {
    layer = null;
    pendingResolve = null;
    confirm(title, body) {
      return this.open({
        title,
        body,
        buttons: [
          { label: "Cancel", value: false },
          { label: "OK", primary: true, value: true }
        ]
      });
    }
    alert(title, body) {
      return this.open({
        title,
        body,
        buttons: [{ label: "OK", primary: true, value: true }]
      }).then(() => void 0);
    }
    openHost(title, content) {
      this.close();
      return new Promise((resolve) => {
        this.pendingResolve = resolve;
        const layer = document.createElement("div");
        layer.className = "sum-dialog-layer";
        layer.setAttribute("role", "presentation");
        const dialog = document.createElement("div");
        dialog.className = "sum-dialog sum-dialog--form";
        dialog.setAttribute("role", "dialog");
        dialog.setAttribute("aria-modal", "true");
        dialog.setAttribute("aria-labelledby", "sum-dialog-title");
        const heading = document.createElement("h2");
        heading.id = "sum-dialog-title";
        heading.className = "sum-dialog-title";
        heading.textContent = title;
        const body = document.createElement("div");
        body.className = "sum-dialog-body sum-dialog-body--host";
        body.appendChild(content);
        const actions = document.createElement("div");
        actions.className = "sum-dialog-actions";
        const closeBtn = document.createElement("button");
        closeBtn.type = "button";
        closeBtn.textContent = "Close";
        closeBtn.className = "sum-dialog-btn";
        closeBtn.addEventListener("click", () => {
          this.close(false);
        });
        actions.appendChild(closeBtn);
        dialog.append(heading, body, actions);
        layer.appendChild(dialog);
        document.body.appendChild(layer);
        this.layer = layer;
        this.bindDismiss(layer);
      });
    }
    open(options) {
      this.close();
      return new Promise((resolve) => {
        this.pendingResolve = resolve;
        const layer = document.createElement("div");
        layer.className = "sum-dialog-layer";
        layer.setAttribute("role", "presentation");
        const dialog = document.createElement("div");
        dialog.className = "sum-dialog";
        dialog.setAttribute("role", "dialog");
        dialog.setAttribute("aria-modal", "true");
        dialog.setAttribute("aria-labelledby", "sum-dialog-title");
        const title = document.createElement("h2");
        title.id = "sum-dialog-title";
        title.className = "sum-dialog-title";
        title.textContent = options.title;
        const body = document.createElement("p");
        body.className = "sum-dialog-body";
        body.textContent = options.body;
        const actions = document.createElement("div");
        actions.className = "sum-dialog-actions";
        const buttons = options.buttons ?? [{ label: "Close", primary: true, value: true }];
        for (const archButton of buttons) {
          const el = document.createElement("button");
          el.type = "button";
          el.textContent = archButton.label;
          el.className = "sum-dialog-btn";
          if (archButton.primary) el.classList.add("sum-dialog-btn--primary");
          if (archButton.danger) el.classList.add("sum-dialog-btn--danger");
          el.addEventListener("click", () => {
            this.close(archButton.value ?? true);
          });
          actions.appendChild(el);
        }
        dialog.append(title, body, actions);
        layer.appendChild(dialog);
        document.body.appendChild(layer);
        this.layer = layer;
        this.bindDismiss(layer);
        actions.querySelector("button")?.focus();
      });
    }
    bindDismiss(layer) {
      const onKey = (event) => {
        if (event.key === "Escape") {
          this.close(false);
        }
      };
      document.addEventListener("keydown", onKey, true);
      layer.addEventListener(
        "click",
        (event) => {
          if (event.target === layer) {
            this.close(false);
          }
        },
        true
      );
      layer.addEventListener(
        "remove",
        () => document.removeEventListener("keydown", onKey, true),
        { once: true }
      );
    }
    close(value = false) {
      const resolve = this.pendingResolve;
      this.pendingResolve = null;
      this.layer?.remove();
      this.layer = null;
      resolve?.(value);
    }
  };

  // src/services/service-registry.ts
  function registerCoreServices(services) {
    const cat = registry.category("services");
    for (const [key, instance] of Object.entries(services)) {
      cat.add(key, instance);
    }
  }

  // src/views/list/control-panel.ts
  function parseFilterCSV(raw) {
    return (raw ?? "").split(",").map((s) => s.trim()).filter(Boolean);
  }
  function toggleFilterName(active, name) {
    if (active.includes(name)) return active.filter((n) => n !== name);
    return [...active, name];
  }
  function renderControlPanel(options) {
    const { payload, state, onPage } = options;
    const rows = payload.records ?? [];
    const total = payload.listTotal ?? rows.length;
    const page = Math.floor(state.offset / state.limit) + 1;
    const pageCount = Math.max(1, Math.ceil(total / state.limit));
    const showPager = pageCount > 1 || state.offset > 0;
    if (!showPager) return html``;
    return html`
    <div class="sum-list-control sum-list-control--secondary">
      <div class="sum-list-pager">
        <button
          type="button"
          class="sum-btn sum-btn--ghost"
          disabled=${state.offset <= 0 ? "disabled" : void 0}
          @click=${() => onPage(Math.max(0, state.offset - state.limit))}
        >
          Prev
        </button>
        <span>${page} / ${pageCount}</span>
        <button
          type="button"
          class="sum-btn sum-btn--ghost"
          disabled=${state.offset + state.limit >= total ? "disabled" : void 0}
          @click=${() => onPage(state.offset + state.limit)}
        >
          Next
        </button>
      </div>
    </div>
  `;
  }
  function renderSearchFilters(options) {
    const domainFilters = options.filters.filter((f) => f.domain || !f.groupBy);
    const groupFilters = options.filters.filter((f) => f.groupBy);
    if (domainFilters.length === 0 && groupFilters.length === 0) return html``;
    return html`
    <div class="sum-search-filters">
      ${domainFilters.map((f) => {
      const on = options.active.includes(f.name);
      return html`<button
          type="button"
          class=${on ? "sum-search-chip sum-search-chip--active" : "sum-search-chip"}
          @click=${() => options.onToggle(f.name)}
        >
          ${f.string || f.name}
        </button>`;
    })}
      ${groupFilters.length ? html`<span class="sum-search-filters-label">Group</span>${groupFilters.map((f) => {
      const on = options.active.includes(f.name);
      return html`<button
              type="button"
              class=${on ? "sum-search-chip sum-search-chip--active" : "sum-search-chip"}
              @click=${() => options.onToggle(f.name)}
            >
              ${f.string || f.name}
            </button>`;
    })}` : ""}
    </div>
  `;
  }
  function renderSortHeader(field, currentSort, onSort) {
    const name = field.name;
    const active = currentSort === name || currentSort === `-${name}`;
    const desc = currentSort === `-${name}`;
    const marker = active ? desc ? " \u2193" : " \u2191" : "";
    return html`<th
    class=${active ? "sum-list-th sum-list-th--sort" : "sum-list-th"}
    @click=${() => onSort(name)}
  >
    ${field.string ?? field.name}${marker}
  </th>`;
  }
  function renderRowCheckbox(id, selected, onToggle) {
    return html`<td class="sum-list-select-cell" @click=${(event) => event.stopPropagation()}>
    <input
      type="checkbox"
      checked=${selected ? "checked" : void 0}
      @change=${(event) => onToggle(id, checkboxCheckedFromEvent(event))}
    />
  </td>`;
  }
  function renderSelectAllHeader(allSelected, onToggleAll) {
    return html`<th class="sum-list-select-head">
    <input
      type="checkbox"
      title="Select all"
      checked=${allSelected ? "checked" : void 0}
      @change=${(event) => onToggleAll(checkboxCheckedFromEvent(event))}
    />
  </th>`;
  }

  // src/template/helpers.ts
  function keyedResult(key, result) {
    return {
      key,
      render() {
        const element = result.render();
        element.dataset.swcKey = key;
        return element;
      },
      patch(existing) {
        const element = result.patch(existing);
        element.dataset.swcKey = key;
        return element;
      }
    };
  }
  function forEach(items, keyFn, renderFn) {
    return items.map((item, index) => keyedResult(String(keyFn(item, index)), renderFn(item, index)));
  }

  // src/views/list/ListView.ts
  var ListView = class extends SwcComponent {
    panelState = {
      search: "",
      offset: 0,
      limit: 40,
      selectedIds: /* @__PURE__ */ new Set(),
      filters: []
    };
    deleting = false;
    acting = false;
    fieldHost;
    setup() {
      this.fieldHost = new FieldHost(this.env);
      this.syncFromPayload(this.props.payload);
    }
    onPropsChanged(props) {
      this.syncFromPayload(props.payload);
      this.panelState.selectedIds = /* @__PURE__ */ new Set();
      this.fieldHost.clear();
    }
    onWillUnmount() {
      this.fieldHost.clear();
    }
    syncFromPayload(payload) {
      this.panelState.search = payload.listSearch ?? "";
      this.panelState.offset = payload.listOffset ?? 0;
      this.panelState.order = payload.listSort ?? "";
      this.panelState.filters = parseFilterCSV(payload.listFilter);
    }
    columns() {
      return this.props.payload.arch.fields.filter((f) => !f.invisible);
    }
    pageRows() {
      return [...this.props.payload.records ?? []];
    }
    navigateList(patch) {
      const payload = this.props.payload;
      const url = this.env.services.router.workspaceUrl({
        actionId: payload.actionId,
        menuId: payload.menuId,
        viewType: VIEW_LIST,
        listSearch: patch.listSearch ?? this.panelState.search,
        listOffset: patch.listOffset ?? 0,
        listSort: patch.listSort ?? this.panelState.order ?? "",
        listFilter: patch.listFilter ?? this.panelState.filters.join(","),
        model: payload.actionId ? "" : payload.model
      });
      this.env.services.action.navigate(url);
    }
    reloadCollection() {
      this.navigateList({ listSearch: this.panelState.search, listOffset: 0 });
    }
    applyPage(offset) {
      this.navigateList({ listOffset: offset });
    }
    applySort(fieldName) {
      const current = this.panelState.order ?? "";
      let next = fieldName;
      if (current === fieldName) next = `-${fieldName}`;
      else if (current === `-${fieldName}`) next = "";
      this.navigateList({ listSort: next, listOffset: 0 });
    }
    applyFilter(name) {
      const next = toggleFilterName(this.panelState.filters, name);
      this.navigateList({ listFilter: next.join(","), listOffset: 0 });
    }
    openRow(row) {
      const id = Number(row.id ?? 0);
      if (id <= 0) return;
      this.env.services.action.openRecord({
        actionId: this.props.payload.actionId,
        menuId: this.props.payload.menuId,
        recordId: id,
        viewType: VIEW_FORM
      });
    }
    toggleRow(id, checked) {
      if (checked) this.panelState.selectedIds.add(id);
      else this.panelState.selectedIds.delete(id);
      this.rerender();
    }
    toggleAll(checked, ids) {
      this.panelState.selectedIds = checked ? new Set(ids) : /* @__PURE__ */ new Set();
      this.rerender();
    }
    toolbarBusy() {
      return this.deleting || this.acting;
    }
    headerObjectButtons() {
      return (this.props.payload.arch.header?.buttons ?? []).filter(
        (archButton) => archButton.type === "object"
      );
    }
    async bulkDelete() {
      const ids = [...this.panelState.selectedIds];
      if (ids.length === 0 || this.toolbarBusy()) return;
      const ok = await this.env.services.dialog.confirm(
        "Delete records",
        `Delete ${ids.length} selected record(s)?`
      );
      if (!ok) return;
      this.deleting = true;
      this.rerender();
      try {
        await this.env.services.rpc.unlink(this.props.payload.model, ids);
        this.panelState.selectedIds = /* @__PURE__ */ new Set();
        this.env.services.notification.success("Deleted", `${ids.length} record(s) removed.`);
        this.reloadCollection();
      } catch (err) {
        this.env.services.notification.error(
          "Delete failed",
          err instanceof SwcError ? err.message : String(err)
        );
      } finally {
        this.deleting = false;
        this.rerender();
      }
    }
    async runHeaderObject(archButton) {
      const ids = [...this.panelState.selectedIds];
      if (ids.length === 0 || this.toolbarBusy()) return;
      this.acting = true;
      this.rerender();
      const navigated = await runObjectAction(this.env, {
        model: this.props.payload.model,
        methodName: archButton.name,
        recordId: ids[0],
        extraArgs: { active_ids: ids.join(",") },
        buttonLabel: archButton.string || archButton.name,
        onSuccess: () => this.reloadCollection()
      });
      this.acting = false;
      if (!navigated) this.rerender();
    }
    /** List cells use readonly field widgets via FieldHost. */
    renderRow(row) {
      const id = Number(row.id ?? 0);
      const cols = this.columns();
      const record = new SwcRecord(this.props.payload.model, id, row);
      return html`<tr class="sum-list-row sum-list-row--click" @click=${() => this.openRow(row)}>
      ${renderRowCheckbox(
        id,
        this.panelState.selectedIds.has(id),
        (rid, checked) => this.toggleRow(rid, checked)
      )}
      ${cols.map((c) => {
        return html`<td class="sum-list-td">${this.fieldHost.render(c, record, true)}</td>`;
      })}
    </tr>`;
    }
    template() {
      const payload = this.props.payload;
      const cols = this.columns();
      const rows = this.pageRows();
      const ids = rows.map((r) => Number(r.id ?? 0)).filter((id) => id > 0);
      const allSelected = ids.length > 0 && ids.every((id) => this.panelState.selectedIds.has(id));
      const filters = payload.arch.search?.filters ?? [];
      return html`
      <div class="sum-list-view">
        ${renderCollectionToolbar({
        payload,
        viewType: VIEW_LIST,
        search: this.panelState.search,
        onSearch: () => this.reloadCollection(),
        onInput: (next) => {
          this.panelState.search = next;
        },
        extraPrimary: [
          this.panelState.selectedIds.size > 0 ? html`<button
                  type="button"
                  class="sum-btn sum-btn--danger"
                  disabled=${this.toolbarBusy() ? "disabled" : void 0}
                  @click=${() => void this.bulkDelete()}
                >
                  Delete (${this.panelState.selectedIds.size})
                </button>` : "",
          this.panelState.selectedIds.size >= 2 ? this.headerObjectButtons().map(
            (archButton) => headerButton(
              archButton.string || archButton.name,
              archButton.class,
              () => void this.runHeaderObject(archButton),
              this.toolbarBusy()
            )
          ) : ""
        ]
      })}
        ${renderSearchFilters({
        filters,
        active: this.panelState.filters,
        onToggle: (name) => this.applyFilter(name)
      })}
        ${renderControlPanel({
        payload,
        state: this.panelState,
        onPage: (o) => this.applyPage(o)
      })}
        <div class="sum-list-table-wrap">
          <table class="sum-list-table">
            <thead>
              <tr>
                ${renderSelectAllHeader(allSelected, (checked) => this.toggleAll(checked, ids))}
                ${cols.map(
        (c) => renderSortHeader(c, this.panelState.order ?? "", (name) => this.applySort(name))
      )}
              </tr>
            </thead>
            <tbody>
              ${forEach(rows, (row) => Number(row.id ?? 0), (row) => this.renderRow(row))}
            </tbody>
          </table>
        </div>
      </div>
    `;
    }
  };

  // src/devtools/profiler.ts
  var events = [];
  var MAX = 500;
  function logRenderEvent(kind, component, durationMs) {
    events.push({ ts: Date.now(), kind, component, durationMs });
    if (events.length > MAX) events.shift();
  }

  // src/devtools/panel.ts
  var panelEl = null;
  var selectedId = null;
  function mountDevtoolsPanel() {
    if (typeof window === "undefined") return;
    if (!window.__SWC_DEVTOOLS__) return;
    if (document.getElementById("swc-vision-panel")) return;
    panelEl = document.createElement("aside");
    panelEl.id = "swc-vision-panel";
    panelEl.className = "sum-devtools-panel";
    panelEl.innerHTML = `
    <header class="sum-devtools-header">
      <strong>SWC Vision</strong>
      <button type="button" id="swc-vision-close">\xD7</button>
    </header>
    <section class="sum-devtools-tree" id="swc-vision-tree"></section>
    <section class="sum-devtools-template" id="swc-vision-template"></section>
  `;
    document.body.appendChild(panelEl);
    panelEl.querySelector("#swc-vision-close")?.addEventListener("click", () => {
      panelEl?.remove();
      panelEl = null;
    });
    refreshTree();
    setInterval(refreshTree, 1e3);
  }
  function refreshTree() {
    if (!panelEl) return;
    const tree = panelEl.querySelector("#swc-vision-tree");
    const templateView = panelEl.querySelector("#swc-vision-template");
    if (!tree) return;
    const comps = window.__SWC_DEVTOOLS__?.components ?? [];
    tree.innerHTML = comps.map(
      (c) => `<button type="button" class="sum-devtools-node${selectedId === c.id ? " sum-devtools-node--active" : ""}" data-id="${c.id}">${c.name} #${c.id}</button>`
    ).join("");
    tree.querySelectorAll("[data-id]").forEach((btn) => {
      btn.addEventListener("click", () => {
        selectedId = Number(btn.dataset.id);
        showTemplate(comps.find((c) => c.id === selectedId) ?? null, templateView);
        refreshTree();
      });
    });
  }
  function showTemplate(comp, el) {
    if (!el || !comp) {
      if (el) el.textContent = "Select a component";
      return;
    }
    const meta = getTemplateSource(comp);
    if (!meta) {
      el.innerHTML = `<p>No template metadata for <code>${comp.name}</code></p>`;
      return;
    }
    el.innerHTML = `
    <h4>${meta.component}</h4>
    <p class="sum-devtools-file">${meta.file}${meta.line ? `:${meta.line}` : ""}</p>
    <pre class="sum-devtools-snippet">${meta.snippet ?? ""}</pre>
  `;
  }
  function enablePicker() {
    document.body.addEventListener(
      "click",
      (ev) => {
        if (!(ev.target instanceof Element)) return;
        if (!ev.altKey) return;
        ev.preventDefault();
        const rec = window.__SWC_DEVTOOLS__?.getComponentForElement(ev.target);
        if (rec) {
          selectedId = rec.id;
          logRenderEvent("pick", rec.name);
          mountDevtoolsPanel();
        }
      },
      true
    );
  }

  // src/devtools/debug.ts
  function isDebugMode() {
    return new URLSearchParams(window.location.search).get("debug") === "1";
  }
  function mountDebugPanel() {
    initDevtoolsBridge();
    if (!isDebugMode()) return;
    if (document.getElementById("sum-debug-panel")) return;
    const el = document.createElement("aside");
    el.id = "sum-debug-panel";
    el.className = "sum-debug-panel";
    el.innerHTML = `<h4>SWC Debug</h4><p>Arch and RPC logging enabled. Alt+click to inspect components.</p>
    <button type="button" id="sum-debug-open-vision">Open SWC Vision</button>`;
    document.body.appendChild(el);
    el.querySelector("#sum-debug-open-vision")?.addEventListener("click", () => mountDevtoolsPanel());
    enablePicker();
  }
  function logWorkspacePayload(label, payload) {
    if (!isDebugMode()) return;
    console.debug(`[SWC ${label}]`, payload);
  }
  function logViewArch(arch) {
    if (!isDebugMode()) return;
    console.debug("[SWC arch]", arch);
  }

  // src/shell/ShellPageView.ts
  function appHref(action) {
    const id = action.replace(/\D/g, "") || action;
    return `/web?action=${encodeURIComponent(id)}`;
  }
  var ShellPageView = class extends SwcComponent {
    template() {
      const { boot, page } = this.props;
      if (page === "apps") {
        return html`<div class="sum-shell-page sum-shell-apps">
        <h1>Applications</h1>
        <div class="sum-shell-app-grid">
          ${boot.apps.map(
          (app) => html`<a class="sum-shell-app-tile" href=${appHref(app.action)}>
              <span class="sum-shell-app-name">${app.name}</span>
            </a>`
        )}
        </div>
      </div>`;
      }
      if (page === "settings") {
        return html`<div class="sum-shell-page sum-shell-settings">
        <h1>Settings</h1>
        <p>Company and user preferences (SPA shell route).</p>
      </div>`;
      }
      return html`<div class="sum-shell-page sum-shell-home">
      <h1>Home</h1>
      <p>Welcome, ${boot.user.name}.</p>
      <div class="sum-shell-app-grid">
        ${boot.apps.slice(0, 6).map(
        (app) => html`<a class="sum-shell-app-tile" href=${appHref(app.action)}>${app.name}</a>`
      )}
      </div>
    </div>`;
    }
  };

  // src/shell/view-tab-sync.ts
  function syncWorkspaceViewTabs(viewTabs) {
    if (viewTabs.length === 0) return;
    const byMode = new Map(viewTabs.map((tab) => [tab.mode, tab]));
    const tabs = document.querySelectorAll(
      ".sum-breadcrumb-right .sum-view-tab[data-view]"
    );
    for (const el of tabs) {
      const tab = byMode.get(el.dataset.view ?? "");
      if (!tab) {
        el.classList.remove("is-active");
        el.removeAttribute("aria-current");
        continue;
      }
      el.href = tab.href;
      el.classList.toggle("is-active", tab.active);
      if (tab.active) el.setAttribute("aria-current", "page");
      else el.removeAttribute("aria-current");
    }
  }
  function initViewTabNavigation() {
    document.addEventListener("click", (ev) => {
      const tab = ev.target.closest(
        ".sum-breadcrumb-right .sum-view-tab[href]"
      );
      if (!tab?.href.includes(`${WEB_ROUTE}?`)) return;
      ev.preventDefault();
      const url = new URL(tab.href, window.location.origin);
      window.history.pushState({}, "", `${url.pathname}${url.search}`);
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
  }

  // src/views/workspace/WorkspaceRouter.ts
  var WorkspaceRouter = class extends SwcComponent {
    payload = null;
    loading = true;
    error = "";
    activeView = null;
    activeViewType = "";
    setup() {
      const load = async () => {
        this.loading = true;
        this.error = "";
        this.rerender();
        try {
          this.payload = await this.fetchWorkspace();
          logWorkspacePayload("workspace", this.payload);
          logViewArch(this.payload.arch);
          syncWorkspaceViewTabs(this.payload.viewTabs);
          this.syncView();
        } catch (err) {
          this.error = err instanceof SwcError ? err.message : String(err);
        } finally {
          this.loading = false;
          this.rerender();
        }
      };
      void load();
      useEffect(() => {
        const onNav = () => void load();
        window.addEventListener("popstate", onNav);
        return () => window.removeEventListener("popstate", onNav);
      });
      useEffect(() => {
        return this.env.services.bus.subscribe(ACTION_CLOSED, () => {
          void load();
        });
      });
      useEffect(() => {
        return this.env.services.bus.subscribe(RECORD_UPDATED, (payload) => {
          const msg = payload;
          if (!this.payload || !msg.model) return;
          if (msg.model !== this.payload.model) return;
          if (msg.id && this.payload.recordId && msg.id !== this.payload.recordId) return;
          void load();
        });
      });
    }
    async fetchWorkspace() {
      const params = RouterService.searchParams(this.env.services.router.parse());
      const base = this.env.bootstrap.swcApiBase || SWC_API_BASE;
      return this.env.services.http.getJSON(`${base}/workspace?${params.toString()}`);
    }
    createView(type, payload) {
      const ViewClass = registry.category("views").get(type) ?? ListView;
      const view = new ViewClass({ payload }, this.env);
      view.callSetup();
      void runWillStart(view).then(() => {
        if (view.rootElement?.isConnected) view.patch();
        else this.rerender();
      });
      return view;
    }
    syncView() {
      if (!this.payload) return;
      const type = this.payload.viewType || this.payload.arch.type;
      if (this.activeView && this.activeViewType === type) {
        this.activeView.updateProps({ payload: this.payload });
        return;
      }
      this.activeView?.destroy();
      this.activeView = this.createView(type, this.payload);
      this.activeViewType = type;
    }
    renderView() {
      if (!this.payload || !this.activeView) return document.createElement("div");
      return this.activeView.renderOrPatch();
    }
    /** Reload workspace payload (e.g. after bus event). */
    reload() {
      void this.fetchWorkspace().then((payload) => {
        this.payload = payload;
        syncWorkspaceViewTabs(payload.viewTabs);
        this.syncView();
        this.patch();
      }).catch((err) => {
        this.error = err instanceof SwcError ? err.message : String(err);
        this.patch();
      });
    }
    template() {
      const route = this.env.services.router.parse();
      if (route.shell === "home" || route.shell === "apps" || route.shell === "settings") {
        const page = route.shell;
        const shellView = new ShellPageView({ boot: this.env.bootstrap, page }, this.env);
        return html`<div class="sum-workspace-root sum-workspace-root--shell">${shellView.render()}</div>`;
      }
      if (this.loading) {
        return html`<div class="sum-workspace-loading">Loading workspace…</div>`;
      }
      if (this.error) {
        return html`<div class="sum-flash sum-flash--error">${this.error}</div>`;
      }
      if (!this.payload) return html`<div></div>`;
      return html`<div class="sum-workspace-root sum-workspace-view">${this.renderView()}</div>`;
    }
  };

  // src/shell/ShellLayout.ts
  var ShellLayout = class extends SwcComponent {
    workspaceRouter;
    setup() {
      this.workspaceRouter = new WorkspaceRouter({}, this.env);
      this.workspaceRouter.callSetup();
      if (this.env.bootstrap.busEnabled) {
        this.env.services.bus.connect();
      }
    }
    workspaceView() {
      return this.workspaceRouter.renderOrPatch();
    }
    template() {
      return html`
      <div id="swc-root-inner">
        <main class="sum-workspace-inner">
          ${this.workspaceView()}
        </main>
      </div>
    `;
    }
  };

  // src/util/shell-storage.ts
  var KEY_SIDEBAR = "sum.shell.sidebarCollapsed";
  var KEY_ACTIVITY_WIDTH = "sum.shell.activityWidthPx";
  var KEY_ACTIVITY_HIDDEN = "sum.shell.activityHidden";
  function readBool(key) {
    try {
      return localStorage.getItem(key) === "1";
    } catch {
      return false;
    }
  }
  function writeBool(key, value) {
    try {
      localStorage.setItem(key, value ? "1" : "0");
    } catch {
    }
  }
  function readActivityWidth() {
    try {
      const n = parseInt(localStorage.getItem(KEY_ACTIVITY_WIDTH) ?? "", 10);
      if (n >= 200 && n <= 520) return n;
    } catch {
    }
    return 300;
  }
  function writeActivityWidth(px) {
    try {
      localStorage.setItem(KEY_ACTIVITY_WIDTH, String(Math.round(px)));
    } catch {
    }
  }
  function readJSON(key, fallback) {
    try {
      const raw = localStorage.getItem(key);
      if (!raw) return fallback;
      const value = JSON.parse(raw);
      return Array.isArray(value) ? value : fallback;
    } catch {
      return fallback;
    }
  }

  // src/shell/activity-panel.ts
  var CHEVRON_LEFT = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M15 18l-6-6 6-6"/></svg>';
  function applyActivityWidth(px) {
    document.documentElement.style.setProperty("--sum-activity-width", `${px}px`);
  }
  function applyActivityHidden(shell, hidden) {
    shell.classList.toggle("sum-shell--activity-hidden", hidden);
    const reveal = document.getElementById("sum-activity-reveal");
    if (reveal) reveal.hidden = !hidden;
    const toggle = document.getElementById("sum-activity-toggle");
    if (toggle) {
      const pressed = !hidden;
      toggle.setAttribute("aria-pressed", pressed ? "true" : "false");
      toggle.classList.toggle("is-pressed", pressed);
    }
  }
  function paintActivityRevealIcon() {
    const reveal = document.getElementById("sum-activity-reveal");
    if (reveal && !reveal.firstChild) reveal.innerHTML = CHEVRON_LEFT;
  }
  function initActivityTabs() {
    document.querySelectorAll("[data-sum-activity-tab]").forEach((tab) => {
      tab.addEventListener("click", () => {
        const name = tab.getAttribute("data-sum-activity-tab");
        const panes = {
          messages: "sum-activity-pane-messages",
          log: "sum-activity-pane-log"
        };
        document.querySelectorAll("[data-sum-activity-tab]").forEach((t) => {
          const on = t.getAttribute("data-sum-activity-tab") === name;
          t.classList.toggle("is-active", on);
          t.setAttribute("aria-selected", on ? "true" : "false");
        });
        for (const [key, id] of Object.entries(panes)) {
          const el = document.getElementById(id);
          if (el) el.hidden = key !== name;
        }
      });
    });
  }
  function initActivityResizer(shell) {
    const resizer = document.getElementById("sum-activity-resizer");
    if (!resizer) return;
    let dragging = false;
    let startX = 0;
    let startW = 300;
    resizer.addEventListener("mousedown", (ev) => {
      if (shell.classList.contains("sum-shell--activity-hidden")) return;
      dragging = true;
      startX = ev.clientX;
      startW = readActivityWidth();
      document.body.style.cursor = "col-resize";
      document.body.style.userSelect = "none";
      ev.preventDefault();
    });
    window.addEventListener("mousemove", (ev) => {
      if (!dragging) return;
      const delta = startX - ev.clientX;
      let width = startW + delta;
      width = Math.min(520, Math.max(200, width));
      applyActivityWidth(width);
    });
    window.addEventListener("mouseup", () => {
      if (!dragging) return;
      dragging = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      const raw = document.documentElement.style.getPropertyValue("--sum-activity-width");
      const px = parseInt(raw, 10);
      if (!Number.isNaN(px)) writeActivityWidth(px);
    });
  }
  function initActivityPanel(shell) {
    applyActivityWidth(readActivityWidth());
    applyActivityHidden(shell, readBool(KEY_ACTIVITY_HIDDEN));
    paintActivityRevealIcon();
    const setActivityHidden = (hidden) => {
      applyActivityHidden(shell, hidden);
      writeBool(KEY_ACTIVITY_HIDDEN, hidden);
    };
    document.getElementById("sum-activity-toggle")?.addEventListener("click", () => {
      setActivityHidden(!shell.classList.contains("sum-shell--activity-hidden"));
    });
    document.getElementById("sum-activity-collapse")?.addEventListener("click", () => {
      setActivityHidden(true);
    });
    document.getElementById("sum-activity-reveal")?.addEventListener("click", () => {
      setActivityHidden(false);
    });
    initActivityTabs();
    initActivityResizer(shell);
  }

  // src/shell/pinned-apps.ts
  var KEY_PINNED_LEGACY = "sumeru:pinned-apps";
  var pinnedCache = [];
  var cacheLoaded = false;
  async function persistPinnedApps(http, modules) {
    const data = await http.postJSON("/web/user/pinned-apps", { modules });
    return Array.isArray(data.modules) ? data.modules.map(String) : modules;
  }
  function loadPinnedApps(initial) {
    if (!cacheLoaded) {
      pinnedCache = initial.slice();
      cacheLoaded = true;
    }
    return pinnedCache.slice();
  }
  function setPinnedCache(modules) {
    pinnedCache = modules.slice();
    cacheLoaded = true;
  }
  function getPinnedApps() {
    return pinnedCache.slice();
  }
  function togglePinnedApp(http, moduleName) {
    const mod = String(moduleName || "").trim();
    if (!mod) return getPinnedApps();
    const previous = getPinnedApps();
    let pins = previous.slice();
    if (pins.includes(mod)) {
      pins = pins.filter((m) => m !== mod);
    } else {
      pins = [mod, ...pins];
    }
    setPinnedCache(pins);
    persistPinnedApps(http, pins).then((saved) => {
      setPinnedCache(saved);
    }).catch(() => {
      setPinnedCache(previous);
    });
    return pins;
  }
  function applyTopNavFilter() {
    const nav = document.querySelector(".sum-top-nav");
    if (!nav) return;
    const moduleItems = [...nav.querySelectorAll(".top-menu-item--module")];
    if (moduleItems.length === 0) return;
    const pins = getPinnedApps();
    const visibleMods = new Set(pins);
    moduleItems.forEach((el) => {
      const mod = el.getAttribute("data-module") ?? "";
      el.classList.toggle("is-topbar-hidden", !visibleMods.has(mod));
    });
    const activeEl = nav.querySelector(".top-menu-item.active");
    activeEl?.scrollIntoView?.({ inline: "nearest", block: "nearest", behavior: "instant" });
  }
  function initPinnedApps(http, initial) {
    loadPinnedApps(initial);
    const legacy = readJSON(KEY_PINNED_LEGACY, []);
    if (getPinnedApps().length === 0 && legacy.length > 0) {
      persistPinnedApps(http, legacy).then((saved) => {
        setPinnedCache(saved);
        try {
          localStorage.removeItem(KEY_PINNED_LEGACY);
        } catch {
        }
        applyTopNavFilter();
      }).catch(() => {
      });
    }
    applyTopNavFilter();
  }

  // src/shell/home-dashboard.ts
  function updatePinButton(btn, displayName, pinned) {
    const name = displayName || btn.getAttribute("data-module") || "App";
    btn.classList.toggle("is-pinned", pinned);
    btn.setAttribute("aria-pressed", pinned ? "true" : "false");
    const label = pinned ? `Unpin ${name} from top bar` : `Pin ${name} to top bar`;
    btn.setAttribute("aria-label", label);
    btn.setAttribute("title", pinned ? "Pinned to top bar \u2014 click to unpin" : "Pin to top bar");
  }
  function syncAllPinButtons() {
    const pins = getPinnedApps();
    document.querySelectorAll(".sum-home-hub-app-pin").forEach((btn) => {
      const mod = btn.getAttribute("data-module") ?? "";
      updatePinButton(btn, btn.getAttribute("data-display-name") ?? mod, pins.includes(mod));
    });
  }
  function tileDisplayName(tile) {
    const nameEl = tile.querySelector(".sum-home-hub-app-name");
    return (nameEl?.textContent ?? tile.getAttribute("data-module") ?? "").trim().toLowerCase();
  }
  function sortTilesAZ(tiles) {
    return [...tiles].sort((a, b) => {
      const displayA = tileDisplayName(a);
      const displayB = tileDisplayName(b);
      if (displayA !== displayB) return displayA.localeCompare(displayB);
      return (a.getAttribute("data-module") ?? "").localeCompare(b.getAttribute("data-module") ?? "");
    });
  }
  function organizePinnedGrid() {
    const pinnedSection = document.getElementById("sum-home-pinned-section");
    const pinnedContainer = document.getElementById("sum-home-pinned-apps");
    const allContainer = document.getElementById("sum-home-all-apps");
    if (!pinnedSection || !pinnedContainer || !allContainer) return;
    const pins = getPinnedApps();
    const allTiles = [
      ...pinnedContainer.querySelectorAll(".sum-home-hub-app"),
      ...allContainer.querySelectorAll(".sum-home-hub-app")
    ];
    const tilesByModule = {};
    allTiles.forEach((tile) => {
      const mod = tile.getAttribute("data-module");
      if (mod) tilesByModule[mod] = tile;
    });
    const pinnedTiles = pins.map((mod) => tilesByModule[mod]).filter(Boolean);
    sortTilesAZ(pinnedTiles).forEach((tile) => {
      pinnedContainer.appendChild(tile);
    });
    sortTilesAZ(allTiles.filter((tile) => !pins.includes(tile.getAttribute("data-module") ?? ""))).forEach(
      (tile) => {
        allContainer.appendChild(tile);
      }
    );
    pinnedSection.hidden = pinnedContainer.children.length === 0;
  }
  function showHomeToast(message) {
    const toast = document.getElementById("sum-home-toast");
    if (!toast) return;
    toast.textContent = message;
    toast.hidden = false;
    window.setTimeout(() => {
      toast.hidden = true;
    }, 3200);
  }
  function initHomeDashboard(http) {
    if (!document.getElementById("sum-home-hub")) return;
    document.addEventListener(
      "click",
      (ev) => {
        const btn = ev.target?.closest(".sum-home-hub-app-pin");
        if (!btn) return;
        ev.preventDefault();
        ev.stopPropagation();
        const mod = btn.getAttribute("data-module") ?? "";
        const displayName = btn.getAttribute("data-display-name") || mod;
        const pins = togglePinnedApp(http, mod);
        const pinned = pins.includes(mod);
        updatePinButton(btn, displayName, pinned);
        syncAllPinButtons();
        organizePinnedGrid();
        applyTopNavFilter();
        showHomeToast(
          pinned ? `${displayName} pinned to top bar` : `${displayName} unpinned from top bar`
        );
      },
      true
    );
    syncAllPinButtons();
    organizePinnedGrid();
  }

  // src/shell/sidebar.ts
  var CHEVRON_RIGHT = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M9 18l6-6-6-6"/></svg>';
  function applySidebar(shell, collapsed) {
    shell.classList.toggle("sum-shell--sidebar-collapsed", collapsed);
    const reveal = document.getElementById("sum-sidebar-reveal");
    if (reveal) reveal.hidden = !collapsed;
  }
  function paintSidebarRevealIcon() {
    const reveal = document.getElementById("sum-sidebar-reveal");
    if (reveal && !reveal.firstChild) reveal.innerHTML = CHEVRON_RIGHT;
  }
  function initSidebar(shell) {
    applySidebar(shell, readBool(KEY_SIDEBAR));
    paintSidebarRevealIcon();
    const toggleSidebar = () => {
      const next = !shell.classList.contains("sum-shell--sidebar-collapsed");
      applySidebar(shell, next);
      writeBool(KEY_SIDEBAR, next);
    };
    for (const id of ["sum-sidebar-toggle", "sum-sidebar-toggle-breadcrumb"]) {
      document.getElementById(id)?.addEventListener("click", toggleSidebar);
    }
    document.getElementById("sum-sidebar-reveal")?.addEventListener("click", () => {
      applySidebar(shell, false);
      writeBool(KEY_SIDEBAR, false);
    });
  }

  // src/shell/company-switcher.ts
  function initCompanySwitcher(boot, http) {
    if (!boot.showCompanySwitcher || boot.companies.length <= 1) return;
    const host = document.getElementById("sum-company-switcher");
    if (!host) return;
    const select = document.createElement("select");
    select.className = "sum-company-switcher-select";
    select.setAttribute("aria-label", "Company");
    for (const company of boot.companies) {
      const opt = document.createElement("option");
      opt.value = String(company.id);
      opt.textContent = company.name;
      if (company.id === boot.activeCompanyId) opt.selected = true;
      select.appendChild(opt);
    }
    select.addEventListener("change", () => {
      void http.postForm("/web/company/switch", {
        company_id: select.value,
        next: window.location.pathname + window.location.search
      }).then(() => {
        document.dispatchEvent(new CustomEvent("swc:company-changed"));
        window.dispatchEvent(new PopStateEvent("popstate"));
      });
    });
    host.replaceChildren(select);
  }

  // src/shell/shell-chrome.ts
  function initShellChrome(boot, http) {
    const shell = document.getElementById("sum-shell");
    if (!shell) return;
    initSidebar(shell);
    initViewTabNavigation();
    if (boot.activityEnabled) {
      initActivityPanel(shell);
    }
    initPinnedApps(http, boot.pinnedApps ?? []);
    initHomeDashboard(http);
    initCompanySwitcher(boot, http);
    new NotificationService().bootstrap(boot.toasts);
  }

  // src/shell/app-launcher.ts
  function buildLauncherIndex(boot) {
    const seen = /* @__PURE__ */ new Set();
    const out = [];
    const add = (item) => {
      const key = item.action.trim();
      if (!key || seen.has(key)) return;
      seen.add(key);
      out.push(item);
    };
    for (const app of boot.apps ?? []) {
      add(appToLauncherItem(app));
    }
    for (const menu of boot.topMenus ?? []) {
      add({
        kind: "menu",
        name: menu.name,
        module: menu.module?.trim() || "Menu",
        action: menu.action
      });
    }
    for (const group of boot.sidebarMenus ?? []) {
      for (const menu of group.subMenus ?? []) {
        add({
          kind: "menu",
          name: menu.name,
          module: menu.module?.trim() || group.name,
          action: menu.action
        });
      }
    }
    return out;
  }
  function appToLauncherItem(app) {
    return {
      kind: app.kind === "menu" ? "menu" : "app",
      name: app.name,
      module: app.module,
      action: app.action,
      description: app.description
    };
  }
  function fuzzyScore(query, text) {
    const q = query.trim().toLowerCase();
    const t = text.trim().toLowerCase();
    if (!q) return 1;
    if (!t) return 0;
    if (t === q) return 100;
    if (t.startsWith(q)) return 80;
    if (t.includes(q)) return 60;
    let qi = 0;
    for (let i = 0; i < t.length && qi < q.length; i++) {
      if (t[i] === q[qi]) qi++;
    }
    return qi === q.length ? 40 : 0;
  }
  function scoreLauncherItem(query, item) {
    const q = query.trim();
    if (!q) return 1;
    const fields = [item.name, item.module, item.description ?? "", item.kind];
    return Math.max(...fields.map((field) => fuzzyScore(q, field)));
  }
  function filterLauncherItems(items, query) {
    const q = query.trim();
    if (!q) return items;
    return items.map((item) => ({ item, score: scoreLauncherItem(q, item) })).filter(({ score }) => score > 0).sort((a, b) => b.score - a.score || a.item.name.localeCompare(b.item.name)).map(({ item }) => item);
  }
  var initialized = false;
  function initAppLauncher(boot, action) {
    if (initialized) return;
    initialized = true;
    const dlg = document.getElementById("sum-app-launcher");
    const input = document.getElementById("sum-app-launcher-input");
    const results = document.getElementById("sum-app-launcher-results");
    const searchBtn = document.getElementById("sum-topbar-search-open");
    const searchField = document.getElementById("sum-topbar-search-field");
    if (!dlg || !input || !results) return;
    const items = buildLauncherIndex(boot);
    let query = "";
    let activeIndex = 0;
    let open = false;
    let pointerNav = false;
    const filtered = () => filterLauncherItems(items, query);
    const scrollActiveIntoView = () => {
      const active = results.querySelector(".sum-app-launcher-result.is-active");
      if (active && typeof active.scrollIntoView === "function") {
        active.scrollIntoView({ block: "nearest" });
      }
    };
    const syncActiveRow = () => {
      const rows = results.querySelectorAll(".sum-app-launcher-result");
      rows.forEach((row, index) => {
        const selected = index === activeIndex;
        row.classList.toggle("is-active", selected);
        row.setAttribute("aria-selected", selected ? "true" : "false");
      });
      scrollActiveIntoView();
    };
    const setActiveIndex = (index) => {
      const list = filtered();
      if (list.length === 0) {
        activeIndex = 0;
        return;
      }
      activeIndex = Math.max(0, Math.min(index, list.length - 1));
      syncActiveRow();
    };
    const close = () => {
      if (!open) return;
      open = false;
      pointerNav = false;
      query = "";
      activeIndex = 0;
      input.value = "";
      results.replaceChildren();
      if (dlg.open) dlg.close();
    };
    const renderResults = () => {
      const list = filtered();
      if (activeIndex >= list.length) activeIndex = Math.max(0, list.length - 1);
      results.replaceChildren();
      list.forEach((item, index) => {
        const row = document.createElement("li");
        row.className = "sum-app-launcher-result";
        if (index === activeIndex) row.classList.add("is-active");
        row.setAttribute("role", "option");
        row.setAttribute("aria-selected", index === activeIndex ? "true" : "false");
        const letter = document.createElement("span");
        letter.className = "sum-app-launcher-result-letter";
        letter.textContent = (item.name.trim()[0] ?? "?").toUpperCase();
        const body = document.createElement("span");
        body.className = "sum-app-launcher-result-body";
        const nameRow = document.createElement("span");
        nameRow.className = "sum-app-launcher-result-name-row";
        const name = document.createElement("span");
        name.className = "sum-app-launcher-result-name";
        name.textContent = item.name;
        const kind = document.createElement("span");
        kind.className = `sum-app-launcher-result-kind sum-app-launcher-result-kind--${item.kind}`;
        kind.textContent = item.kind === "app" ? "App" : "Menu";
        nameRow.append(name, kind);
        const meta = document.createElement("span");
        meta.className = "sum-app-launcher-result-meta";
        meta.textContent = item.description?.trim() || item.module;
        body.append(nameRow, meta);
        row.append(letter, body);
        row.addEventListener("mouseenter", () => {
          if (!pointerNav) return;
          setActiveIndex(index);
        });
        row.addEventListener("click", () => {
          action.navigate(item.action);
          close();
        });
        results.appendChild(row);
      });
    };
    const openLauncher = () => {
      if (open) return;
      open = true;
      pointerNav = false;
      query = "";
      activeIndex = 0;
      input.value = "";
      renderResults();
      if (!dlg.open) dlg.showModal();
      queueMicrotask(() => input.focus());
    };
    const toggle = () => {
      if (open) close();
      else openLauncher();
    };
    const activate = () => {
      const list = filtered();
      const item = list[activeIndex];
      if (!item) return;
      action.navigate(item.action);
      close();
    };
    const onInputKeydown = (ev) => {
      if (!open) return;
      const list = filtered();
      if (ev.key === "ArrowDown") {
        ev.preventDefault();
        pointerNav = false;
        if (list.length === 0) return;
        setActiveIndex(activeIndex + 1);
        return;
      }
      if (ev.key === "ArrowUp") {
        ev.preventDefault();
        pointerNav = false;
        if (list.length === 0) return;
        setActiveIndex(activeIndex - 1);
        return;
      }
      if (ev.key === "Enter") {
        ev.preventDefault();
        activate();
        return;
      }
      if (ev.key === "Escape") {
        ev.preventDefault();
        close();
      }
    };
    const onGlobalKeydown = (ev) => {
      if ((ev.ctrlKey || ev.metaKey) && ev.key.toLowerCase() === "k") {
        ev.preventDefault();
        toggle();
      }
    };
    input.addEventListener("input", () => {
      query = input.value;
      activeIndex = 0;
      pointerNav = false;
      renderResults();
    });
    results.addEventListener("mousemove", () => {
      pointerNav = true;
    });
    dlg.addEventListener("close", () => {
      open = false;
      pointerNav = false;
      query = "";
      activeIndex = 0;
      input.value = "";
      results.replaceChildren();
    });
    input.addEventListener("keydown", onInputKeydown);
    document.addEventListener("keydown", onGlobalKeydown);
    searchBtn?.addEventListener("click", toggle);
    searchField?.addEventListener("click", toggle);
  }

  // src/addon/loader.ts
  var AddonLoader = class _AddonLoader {
    static async loadEntries(urls) {
      for (const url of urls) {
        try {
          await import(
            /* @vite-ignore */
            url
          );
        } catch (err) {
          console.warn("SWC addon entry failed:", url, err);
        }
      }
    }
    static registerFromGlobal() {
      const entries = window.__SWC_ADDON_ENTRIES__;
      if (entries?.length) {
        void _AddonLoader.loadEntries(entries);
      }
    }
  };

  // src/views/kanban/kanban-card.ts
  function isKanbanImageField(field) {
    const name = field.name.toLowerCase();
    return name === "image" || name.startsWith("image_") || field.widget === "image" || field.widget === "circle";
  }
  function isKanbanImageCircle(field) {
    return field.widget === "circle" || field.options?.shape === "circle";
  }
  function isPriorityField(field) {
    return field.name === "priority" || field.widget === "priority";
  }
  function displayValue(row, field) {
    const raw = row[`${field.name}_name`] ?? row[field.name];
    if (raw == null || raw === false) return "";
    return String(raw);
  }
  function imageSrc(row, field) {
    const raw = row[field.name];
    if (typeof raw !== "string" || !raw.trim()) return "";
    const v = raw.trim();
    if (v.startsWith("data:") || v.startsWith("http://") || v.startsWith("https://") || v.startsWith("/")) {
      return v;
    }
    return "";
  }
  function initials(row, fields) {
    const nameField = fields.find((f) => f.name === "name") ?? fields.find((f) => !isKanbanImageField(f));
    const text = nameField ? displayValue(row, nameField) : "";
    const parts = text.trim().split(/\s+/).filter(Boolean);
    if (parts.length === 0) return "?";
    if (parts.length === 1) return parts[0].slice(0, 1).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }
  function titleField(fields) {
    return fields.find((f) => f.name === "name") ?? fields.find((f) => f.name === "display_name") ?? fields.find((f) => !isKanbanImageField(f) && !isPriorityField(f));
  }
  function renderPriority(row, field) {
    const level = Number(row[field.name] ?? 0);
    if (!level) return null;
    const stars = [1, 2, 3].map(
      (n) => html`<span class="sum-kanban-priority-star${n <= level ? " sum-kanban-priority-star--on" : ""}">★</span>`
    );
    return html`<div class="sum-kanban-priority">${stars}</div>`;
  }
  function renderMedia(row, imageField, fields) {
    const src = imageSrc(row, imageField);
    const label = displayValue(row, titleField(fields) ?? imageField);
    if (!src && !label) return null;
    const media = document.createElement("div");
    media.className = `sum-kanban-card-media${isKanbanImageCircle(imageField) ? " sum-kanban-card-media--circle" : " sum-kanban-card-media--square"}`;
    if (src) {
      const img = document.createElement("img");
      img.className = "sum-kanban-card-media-img";
      img.src = src;
      img.alt = "";
      media.appendChild(img);
    } else {
      const initialsEl = document.createElement("span");
      initialsEl.className = "sum-kanban-card-media-initials";
      initialsEl.textContent = initials(row, fields);
      media.appendChild(initialsEl);
    }
    return media;
  }
  function renderKanbanCardInner(row, fields) {
    const imageField = fields.find(isKanbanImageField);
    const priorityField = fields.find(isPriorityField);
    const title = titleField(fields);
    const subs = fields.filter(
      (f) => f !== imageField && f !== title && f !== priorityField && !isKanbanImageField(f) && !isPriorityField(f)
    );
    const media = imageField ? renderMedia(row, imageField, fields) : null;
    const titleEl = title ? html`<div class="sum-kanban-card-title">${displayValue(row, title)}</div>` : null;
    const subEls = subs.map((f) => displayValue(row, f)).filter(Boolean).map((text) => html`<div class="sum-kanban-card-sub">${text}</div>`);
    const priorityEl = priorityField ? renderPriority(row, priorityField) : null;
    if (media) {
      return html`${media}<div class="sum-kanban-card-body">${titleEl}${subEls}${priorityEl}</div>`;
    }
    return html`${titleEl}${subEls}${priorityEl}`;
  }

  // src/views/kanban/KanbanView.ts
  var KanbanView = class extends SwcComponent {
    search = "";
    filters = [];
    drafts = {};
    setup() {
      this.syncFromPayload(this.props.payload);
    }
    onPropsChanged(props) {
      this.syncFromPayload(props.payload);
    }
    syncFromPayload(payload) {
      this.search = payload.listSearch ?? "";
      this.filters = parseFilterCSV(payload.listFilter);
    }
    cardFields() {
      return this.props.payload.arch.fields.filter((f) => !f.invisible);
    }
    navigateKanban(patch) {
      const payload = this.props.payload;
      this.env.services.action.navigate(
        this.env.services.router.workspaceUrl({
          actionId: payload.actionId,
          menuId: payload.menuId,
          viewType: VIEW_KANBAN,
          listSearch: patch.listSearch ?? this.search,
          listFilter: patch.listFilter ?? this.filters.join(",")
        })
      );
    }
    applySearch() {
      this.navigateKanban({ listSearch: this.search });
    }
    applyFilter(name) {
      this.navigateKanban({ listFilter: toggleFilterName(this.filters, name).join(",") });
    }
    openCard(row) {
      const id = Number(row.id ?? 0);
      if (id <= 0) return;
      const payload = this.props.payload;
      this.env.services.action.openRecord({
        actionId: payload.actionId,
        menuId: payload.menuId,
        recordId: id,
        viewType: VIEW_FORM
      });
    }
    async moveCard(recordId, columnValue) {
      const groupField = this.props.payload.arch.kanban?.groupField;
      if (!groupField) return;
      try {
        await this.env.services.rpc.write(this.props.payload.model, [recordId], {
          [groupField]: columnValue || false
        });
        this.env.services.bus.emit(RECORD_UPDATED, {
          model: this.props.payload.model,
          id: recordId
        });
      } catch (err) {
        this.env.services.notification.error(
          "Move failed",
          err instanceof SwcError ? err.message : String(err)
        );
      }
    }
    async quickCreate(columnValue) {
      const key = String(columnValue);
      const name = (this.drafts[key] ?? "").trim();
      const groupField = this.props.payload.arch.kanban?.groupField;
      if (!name || !groupField) return;
      try {
        await this.env.services.rpc.create(this.props.payload.model, {
          ...this.props.payload.defaults ?? {},
          name,
          [groupField]: columnValue || false
        });
        this.drafts[key] = "";
        this.env.services.notification.success("Created", "Record created.");
        this.env.services.bus.emit(RECORD_UPDATED, { model: this.props.payload.model });
      } catch (err) {
        this.env.services.notification.error(
          "Create failed",
          err instanceof SwcError ? err.message : String(err)
        );
      }
    }
    toolbar() {
      return renderCollectionToolbar({
        payload: this.props.payload,
        viewType: VIEW_KANBAN,
        search: this.search,
        onSearch: () => this.applySearch(),
        onInput: (next) => {
          this.search = next;
        }
      });
    }
    renderCard(row, fields, options = {}) {
      const draggable = Boolean(options.draggable);
      const dropValue = options.dropValue;
      return html`<div
      class="sum-kanban-card"
      draggable=${draggable ? "true" : void 0}
      @click=${() => this.openCard(row)}
      @dragstart=${draggable ? (event) => event.dataTransfer?.setData("text/plain", String(row.id)) : void 0}
      @dragover=${dropValue !== void 0 ? (event) => event.preventDefault() : void 0}
      @drop=${dropValue !== void 0 ? (event) => {
        event.preventDefault();
        const id = Number(event.dataTransfer?.getData("text/plain"));
        if (id) void this.moveCard(id, dropValue);
      } : void 0}
    >
      ${renderKanbanCardInner(row, fields)}
    </div>`;
    }
    template() {
      const payload = this.props.payload;
      const kanban = payload.arch.kanban;
      const fields = this.cardFields();
      const filters = payload.arch.search?.filters ?? [];
      if (!kanban?.columns?.length) {
        const rows = payload.records ?? [];
        return html`
        <div class="sum-kanban-view">
          ${this.toolbar()}
          ${renderSearchFilters({
          filters,
          active: this.filters,
          onToggle: (name) => this.applyFilter(name)
        })}
          <div class="sum-kanban-columns">
            ${rows.length === 0 ? html`<div class="sum-kanban-empty">No records</div>` : rows.map((row) => this.renderCard(row, fields))}
          </div>
        </div>
      `;
      }
      return html`
      <div class="sum-kanban-view">
        ${this.toolbar()}
        ${renderSearchFilters({
        filters,
        active: this.filters,
        onToggle: (name) => this.applyFilter(name)
      })}
        <div class="sum-kanban-board sum-kanban-board--grouped">
          <div class="sum-kanban-stage-columns">
            ${kanban.columns.map(
        (col) => html`<div class="sum-kanban-stage-column" data-column=${String(col.value)}>
                <div class="sum-kanban-stage-header">
                  <span>${col.label}</span>
                  <span class="sum-kanban-stage-count">${String(col.records.length)}</span>
                </div>
                <div class="sum-kanban-cards">
                  ${col.records.map(
          (row) => this.renderCard(row, fields, { draggable: kanban.draggable, dropValue: col.value })
        )}
                </div>
                ${kanban.quickCreate ? html`<form
                      class="sum-kanban-quick-create"
                      @submit=${(event) => {
          event.preventDefault();
          void this.quickCreate(col.value);
        }}
                    >
                      <input
                        type="text"
                        class="sum-kanban-quick-input"
                        placeholder="Add…"
                        value=${this.drafts[String(col.value)] ?? ""}
                        @input=${(event) => {
          this.drafts[String(col.value)] = inputValueFromEvent(event);
        }}
                      />
                      <button type="submit" class="sum-btn sum-btn--ghost">Add</button>
                    </form>` : ""}
              </div>`
      )}
          </div>
        </div>
      </div>
    `;
    }
  };

  // src/views/pivot/PivotView.ts
  var PivotView = class extends SwcComponent {
    template() {
      const pivot = this.props.payload.arch.pivot;
      if (!pivot) {
        return html`<div class="sum-pivot-view sum-pivot-view--empty">No pivot data</div>`;
      }
      return html`
      <div class="sum-pivot-view">
        <table class="sum-pivot-table">
          <thead>
            <tr>
              <th></th>
              ${pivot.colLabels.map((c) => html`<th>${c}</th>`)}
            </tr>
          </thead>
          <tbody>
            ${pivot.rowLabels.map(
        (row) => html`<tr>
                <th>${row}</th>
                ${pivot.colLabels.map((col) => {
          const fieldValue = pivot.values[row]?.[col] ?? 0;
          return html`<td>${String(fieldValue)}</td>`;
        })}
              </tr>`
      )}
          </tbody>
        </table>
        <p class="sum-pivot-measure">${pivot.measureLabel}</p>
      </div>
    `;
    }
  };

  // src/views/graph/GraphView.ts
  var GraphView = class extends SwcComponent {
    groups = [];
    measureField = "id";
    groupField = "create_date";
    chart = "bar";
    setup() {
      onWillStart(() => this.load());
    }
    async load() {
      const payload = this.props.payload;
      this.chart = (payload.arch.graph?.chart || "bar").toLowerCase();
      this.groupField = payload.arch.fields.find((f) => f.pivotType === "row")?.name ?? "create_date";
      this.measureField = payload.arch.fields.find((f) => f.pivotType === "measure")?.name ?? "id";
      this.groups = await this.env.services.rpc.readGroup(
        payload.model,
        [],
        [this.measureField],
        [this.groupField],
        40
      );
      this.rerender();
    }
    labelOf(group) {
      const nameKey = `${this.groupField}_name`;
      if (group[nameKey] != null) return String(group[nameKey]);
      if (group[this.groupField] != null) return String(group[this.groupField]);
      return String(group.name ?? "");
    }
    template() {
      const max = Math.max(...this.groups.map((g) => Number(g[this.measureField] ?? 0)), 1);
      if (this.chart === "pie") {
        let accumulatedPercent = 0;
        const total = this.groups.reduce((sum, g) => sum + Number(g[this.measureField] ?? 0), 0) || 1;
        const stops = [];
        const palette = ["#2563eb", "#16a34a", "#f59e0b", "#dc2626", "#7c3aed", "#0891b2"];
        this.groups.forEach((group, index) => {
          const fieldValue = Number(group[this.measureField] ?? 0);
          const start = accumulatedPercent;
          accumulatedPercent += fieldValue / total * 100;
          stops.push(`${palette[index % palette.length]} ${start}% ${accumulatedPercent}%`);
        });
        return html`
        <div class="sum-graph-view">
          <div class="sum-graph-pie" style=${`background:conic-gradient(${stops.join(",")})`}></div>
          <ul class="sum-graph-legend">
            ${this.groups.map(
          (group, index) => html`<li>
                <span class="sum-graph-swatch" style=${`background:${["#2563eb", "#16a34a", "#f59e0b", "#dc2626", "#7c3aed", "#0891b2"][index % 6]}`}></span>
                ${this.labelOf(group)} (${String(group[this.measureField] ?? 0)})
              </li>`
        )}
          </ul>
        </div>
      `;
      }
      return html`
      <div class="sum-graph-view">
        ${this.groups.map((group) => {
        const label = this.labelOf(group);
        const fieldValue = Number(group[this.measureField] ?? 0);
        const pct = Math.round(fieldValue / max * 100);
        return html`<div class="sum-graph-bar-row">
            <span class="sum-graph-label">${label}</span>
            <div class=${this.chart === "line" ? "sum-graph-bar sum-graph-bar--line" : "sum-graph-bar"} style="width:${pct}%"></div>
            <span class="sum-graph-value">${fieldValue}</span>
          </div>`;
      })}
      </div>
    `;
    }
  };

  // src/views/calendar/CalendarView.ts
  var WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
  var CalendarView = class extends SwcComponent {
    dateField = "date_deadline";
    year = 0;
    month = 0;
    setup() {
      const now = /* @__PURE__ */ new Date();
      this.year = now.getFullYear();
      this.month = now.getMonth();
      const archStart = this.props.payload.arch.calendar?.dateStart;
      if (archStart) {
        this.dateField = archStart;
      } else {
        const fields = this.props.payload.arch.fields;
        const dateField = fields.find((f) => f.type === "date" || f.type === "datetime");
        if (dateField) this.dateField = dateField.name;
      }
    }
    eventsByDay() {
      const map = /* @__PURE__ */ new Map();
      for (const row of this.props.payload.records ?? []) {
        const raw = String(row[this.dateField] ?? "").slice(0, 10);
        if (!raw) continue;
        if (!map.has(raw)) map.set(raw, []);
        map.get(raw).push(row);
      }
      return map;
    }
    openRecord(row) {
      const id = Number(row.id ?? 0);
      if (id <= 0) return;
      const payload = this.props.payload;
      this.env.services.action.openRecord({
        actionId: payload.actionId,
        menuId: payload.menuId,
        recordId: id,
        viewType: VIEW_FORM
      });
    }
    shiftMonth(delta) {
      const d = new Date(this.year, this.month + delta, 1);
      this.year = d.getFullYear();
      this.month = d.getMonth();
      this.rerender();
    }
    cells() {
      const first = new Date(this.year, this.month, 1);
      const start = new Date(first);
      start.setDate(1 - first.getDay());
      const out = [];
      for (let index = 0; index < 42; index++) {
        const d = new Date(start);
        d.setDate(start.getDate() + index);
        out.push({ date: d, inMonth: d.getMonth() === this.month });
      }
      return out;
    }
    iso(d) {
      const y = d.getFullYear();
      const m = String(d.getMonth() + 1).padStart(2, "0");
      const day = String(d.getDate()).padStart(2, "0");
      return `${y}-${m}-${day}`;
    }
    template() {
      const events2 = this.eventsByDay();
      const title = new Date(this.year, this.month, 1).toLocaleString(void 0, {
        month: "long",
        year: "numeric"
      });
      return html`
      <div class="sum-calendar-view">
        <div class="sum-calendar-toolbar">
          <button type="button" class="sum-btn sum-btn--ghost" @click=${() => this.shiftMonth(-1)}>Prev</button>
          <h2 class="sum-calendar-title">${this.props.payload.arch.title ?? title}</h2>
          <button type="button" class="sum-btn sum-btn--ghost" @click=${() => this.shiftMonth(1)}>Next</button>
        </div>
        <div class="sum-calendar-grid">
          ${WEEKDAYS.map((d) => html`<div class="sum-calendar-weekday">${d}</div>`)}
          ${this.cells().map((cell) => {
        const key = this.iso(cell.date);
        const rows = events2.get(key) ?? [];
        return html`<section class=${cell.inMonth ? "sum-calendar-cell" : "sum-calendar-cell sum-calendar-cell--muted"}>
              <h3 class="sum-calendar-day-title">${String(cell.date.getDate())}</h3>
              <ul class="sum-calendar-events">
                ${rows.map(
          (row) => html`<li class="sum-calendar-event" @click=${() => this.openRecord(row)}>
                    ${String(row.name ?? row.display_name ?? `#${row.id}`)}
                  </li>`
        )}
              </ul>
            </section>`;
      })}
        </div>
      </div>
    `;
    }
  };

  // src/views/gantt/GanttView.ts
  function parseDate(raw) {
    const text = String(raw ?? "").trim();
    if (!text) return null;
    const date = new Date(text);
    return Number.isNaN(date.getTime()) ? null : date;
  }
  var GanttView = class extends SwcComponent {
    scale;
    setScale;
    setup() {
      const [scale, setScale] = useState("week");
      this.scale = scale;
      this.setScale = setScale;
    }
    dateStartField() {
      return this.props.payload.arch.gantt?.dateStart || this.props.payload.arch.calendar?.dateStart || this.props.payload.arch.fields.find((f) => f.type === "date" || f.type === "datetime")?.name || "date_start";
    }
    dateStopField() {
      return this.props.payload.arch.gantt?.dateStop || this.props.payload.arch.calendar?.dateStop || this.dateStartField();
    }
    openRecord(row) {
      const id = Number(row.id ?? 0);
      if (id <= 0) return;
      const payload = this.props.payload;
      this.env.services.action.openRecord({
        actionId: payload.actionId,
        menuId: payload.menuId,
        recordId: id,
        viewType: VIEW_FORM
      });
    }
    range() {
      const startField = this.dateStartField();
      const stopField = this.dateStopField();
      let min = Infinity;
      let max = -Infinity;
      for (const row of this.props.payload.records ?? []) {
        const start = parseDate(row[startField])?.getTime();
        const stop = parseDate(row[stopField])?.getTime() ?? start;
        if (start == null) continue;
        min = Math.min(min, start);
        max = Math.max(max, stop ?? start);
      }
      if (!Number.isFinite(min)) {
        const now = Date.now();
        return { start: now, end: now + 864e5 * 7 };
      }
      const pad = this.scale.value === "day" ? 864e5 : this.scale.value === "week" ? 864e5 * 7 : 864e5 * 30;
      return { start: min - pad, end: max + pad };
    }
    template() {
      const startField = this.dateStartField();
      const stopField = this.dateStopField();
      const { start, end } = this.range();
      const span = Math.max(end - start, 1);
      const rows = this.props.payload.records ?? [];
      return html`
      <div class="sum-gantt-view">
        <div class="sum-gantt-toolbar">
          <h2>${this.props.payload.arch.title ?? "Gantt"}</h2>
          <div class="sum-gantt-scale">
            ${["day", "week", "month"].map(
        (scale) => html`<button
                type="button"
                class=${this.scale.value === scale ? "sum-btn sum-btn--secondary" : "sum-btn sum-btn--ghost"}
                @click=${() => this.setScale(scale)}
              >${scale}</button>`
      )}
          </div>
        </div>
        <ul class="sum-gantt-rows">
          ${forEach(rows, (row) => Number(row.id ?? 0), (row) => {
        const from = parseDate(row[startField])?.getTime();
        const to = parseDate(row[stopField])?.getTime() ?? from;
        if (from == null || to == null) {
          return html`<li class="sum-gantt-row">
                <span class="sum-gantt-label">${String(row.name ?? row.display_name ?? row.id)}</span>
              </li>`;
        }
        const left = (from - start) / span * 100;
        const width = Math.max((to - from) / span * 100, 0.8);
        return html`<li class="sum-gantt-row" @click=${() => this.openRecord(row)}>
              <span class="sum-gantt-label">${String(row.name ?? row.display_name ?? row.id)}</span>
              <div class="sum-gantt-track">
                <div class="sum-gantt-bar" style=${`left:${left}%;width:${width}%`}></div>
              </div>
            </li>`;
      })}
        </ul>
      </div>
    `;
    }
  };

  // src/views/map/MapView.ts
  function numberField(row, name) {
    const raw = row[name];
    if (raw == null || raw === "") return null;
    const n = Number(raw);
    return Number.isFinite(n) ? n : null;
  }
  var MapView = class extends SwcComponent {
    latField() {
      return this.props.payload.arch.map?.latitude || this.props.payload.arch.fields.find((f) => /lat/i.test(f.name))?.name || "latitude";
    }
    lngField() {
      return this.props.payload.arch.map?.longitude || this.props.payload.arch.fields.find((f) => /lng|lon/i.test(f.name))?.name || "longitude";
    }
    openRecord(row) {
      const id = Number(row.id ?? 0);
      if (id <= 0) return;
      const payload = this.props.payload;
      this.env.services.action.openRecord({
        actionId: payload.actionId,
        menuId: payload.menuId,
        recordId: id,
        viewType: VIEW_FORM
      });
    }
    template() {
      const latName = this.latField();
      const lngName = this.lngField();
      const markers = (this.props.payload.records ?? []).map((row) => {
        const lat = numberField(row, latName);
        const lng = numberField(row, lngName);
        if (lat == null || lng == null) return null;
        return { row, lat, lng };
      }).filter((m) => m != null);
      return html`
      <div class="sum-map-view">
        <h2>${this.props.payload.arch.title ?? "Map"}</h2>
        <p class="sum-map-hint">${markers.length} located record(s).</p>
        <ul class="sum-map-list">
          ${forEach(markers, (marker) => Number(marker.row.id ?? 0), (marker) => html`<li class="sum-map-item">
              <button type="button" class="sum-map-name" @click=${() => this.openRecord(marker.row)}>
                ${String(marker.row.name ?? marker.row.display_name ?? marker.row.id)}
              </button>
              <a
                class="sum-map-link"
                href=${`https://www.openstreetmap.org/?mlat=${marker.lat}&mlon=${marker.lng}#map=16/${marker.lat}/${marker.lng}`}
                target="_blank"
                rel="noopener"
              >${marker.lat.toFixed(4)}, ${marker.lng.toFixed(4)}</a>
            </li>`)}
        </ul>
      </div>
    `;
    }
  };

  // src/views/cohort/CohortView.ts
  function parseDate2(raw) {
    const text = String(raw ?? "").trim();
    if (!text) return null;
    const date = new Date(text);
    return Number.isNaN(date.getTime()) ? null : date;
  }
  function bucketKey(date, interval) {
    if (interval === "month") {
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;
    }
    const tmp = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
    const day = tmp.getUTCDay() || 7;
    tmp.setUTCDate(tmp.getUTCDate() + 4 - day);
    const yearStart = new Date(Date.UTC(tmp.getUTCFullYear(), 0, 1));
    const week = Math.ceil(((tmp.getTime() - yearStart.getTime()) / 864e5 + 1) / 7);
    return `${tmp.getUTCFullYear()}-W${String(week).padStart(2, "0")}`;
  }
  var CohortView = class extends SwcComponent {
    dateField() {
      return this.props.payload.arch.cohort?.dateStart || this.props.payload.arch.calendar?.dateStart || this.props.payload.arch.gantt?.dateStart || this.props.payload.arch.fields.find((f) => f.type === "date" || f.type === "datetime")?.name || "create_date";
    }
    measureField() {
      return this.props.payload.arch.cohort?.measure || this.props.payload.arch.fields.find((f) => f.pivotType === "measure")?.name || "";
    }
    interval() {
      const raw = (this.props.payload.arch.cohort?.interval ?? "month").toLowerCase();
      return raw === "week" ? "week" : "month";
    }
    table() {
      const dateField = this.dateField();
      const measureField = this.measureField();
      const interval = this.interval();
      const groups = /* @__PURE__ */ new Map();
      for (const row of this.props.payload.records ?? []) {
        const date = parseDate2(row[dateField]);
        if (!date) continue;
        const key = bucketKey(date, interval);
        const amount = measureField ? Number(row[measureField] ?? 0) : 1;
        groups.set(key, (groups.get(key) ?? 0) + (Number.isFinite(amount) ? amount : 0));
      }
      const periods = [...groups.keys()].sort();
      const rows = periods.map((cohort, index) => {
        const values = periods.map((_, col) => {
          if (col < index) return 0;
          const later = periods[col];
          return groups.get(later) ?? 0;
        });
        return { cohort, values };
      });
      return { periods, rows };
    }
    template() {
      const { periods, rows } = this.table();
      return html`
      <div class="sum-cohort-view">
        <h2>${this.props.payload.arch.title ?? "Cohort"}</h2>
        <table class="sum-cohort-table">
          <thead>
            <tr>
              <th>Cohort</th>
              ${periods.map((p) => html`<th>${p}</th>`)}
            </tr>
          </thead>
          <tbody>
            ${rows.map(
        (row) => html`<tr>
                <th>${row.cohort}</th>
                ${row.values.map((value) => html`<td>${value === 0 ? "" : String(value)}</td>`)}
              </tr>`
      )}
          </tbody>
        </table>
      </div>
    `;
    }
  };

  // src/i18n/translate.ts
  var translations = /* @__PURE__ */ new Map();
  function loadTranslations(source) {
    translations.clear();
    if (!source) return;
    for (const [k, v] of Object.entries(source)) {
      translations.set(k, v);
    }
  }

  // src/main.ts
  var VIEW_CONSTRUCTORS = {
    list: ListView,
    form: FormView,
    kanban: KanbanView,
    pivot: PivotView,
    graph: GraphView,
    calendar: CalendarView,
    gantt: GanttView,
    map: MapView,
    cohort: CohortView
  };
  function registerCore() {
    registerDefaultWidgets();
    const views = registry.category("views");
    for (const [name, ViewClass] of Object.entries(VIEW_CONSTRUCTORS)) {
      views.add(name, ViewClass);
    }
    const main = registry.category("main_components");
    main.add("shell", ShellLayout);
  }
  function buildEnv(boot) {
    const router = new RouterService();
    const services = {
      rpc: new RpcService(boot.rpcUrl, boot.csrfToken),
      http: new HttpService(boot.csrfToken),
      notification: new NotificationService(),
      action: new ActionService(router),
      router,
      bus: new BusService(),
      dialog: new DialogService()
    };
    registerCoreServices(services);
    return new SwcEnv(boot, services);
  }
  function bootstrap() {
    registerCore();
    AddonLoader.registerFromGlobal();
    let boot;
    try {
      boot = readBootstrap();
    } catch {
      return;
    }
    const env = buildEnv(boot);
    env.services.action.setEnv(env);
    loadTranslations(boot.translations);
    initDevtoolsBridge();
    mountDebugPanel();
    initShellChrome(boot, env.services.http);
    initAppLauncher(boot, env.services.action);
    const mountEl = document.getElementById("swc-workspace");
    if (mountEl) {
      SwcApp.start(mountEl, env, ShellLayout);
    }
  }
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", bootstrap);
  } else {
    bootstrap();
  }
  return __toCommonJS(main_exports);
})();
//# sourceMappingURL=swc.js.map
