import { html, type TemplateResult, type TemplateValue } from "./html.js";

/** Marker result for keyed list reconciliation in runtime/patch/keyed.ts */
export class KeyedListResult implements TemplateResult {
  readonly __keyedList = true as const;

  constructor(
    readonly items: Array<{ key: string; result: TemplateResult }>,
  ) {}

  render(): HTMLElement {
    const wrap = document.createElement("div");
    wrap.style.display = "contents";
    wrap.dataset.swcKeyedList = "1";
    for (const item of this.items) {
      const el = item.result.render();
      el.dataset.swcKey = item.key;
      wrap.appendChild(el);
    }
    return wrap;
  }
}

function keyedResult(key: string, result: TemplateResult): TemplateResult {
  return {
    render() {
      const el = result.render();
      el.dataset.swcKey = key;
      return el;
    },
  };
}

/** OWL-style t-foreach helper — returns keyed TemplateResults for tbody rows etc. */
export function forEach<T>(
  items: T[],
  keyFn: (item: T, index: number) => string | number,
  renderFn: (item: T, index: number) => TemplateResult,
): TemplateResult[] {
  return items.map((item, index) => keyedResult(String(keyFn(item, index)), renderFn(item, index)));
}

/** Container-based keyed list (non-table contexts). */
export function forEachBlock<T>(
  items: T[],
  keyFn: (item: T, index: number) => string | number,
  renderFn: (item: T, index: number) => TemplateResult,
): KeyedListResult {
  return new KeyedListResult(
    items.map((item, index) => ({
      key: String(keyFn(item, index)),
      result: renderFn(item, index),
    })),
  );
}

/** OWL-style t-if / t-elif / t-else chain. */
export function when(
  condition: unknown,
  renderFn: () => TemplateResult,
  ...elifs: Array<[unknown, () => TemplateResult]>
): TemplateValue {
  if (condition) return renderFn();
  for (const [cond, fn] of elifs) {
    if (cond) return fn();
  }
  return null;
}

export function fragment(children: TemplateValue[]): TemplateResult {
  return html`${children as unknown as TemplateValue}`;
}

export function isKeyedList(value: unknown): value is KeyedListResult {
  return (
    typeof value === "object" &&
    value !== null &&
    (value as KeyedListResult).__keyedList === true
  );
}
