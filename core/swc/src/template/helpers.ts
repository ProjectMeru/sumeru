import { type TemplateResult, type TemplateValue } from "./html.js";

function keyedResult(key: string, result: TemplateResult): TemplateResult {
  return {
    render() {
      const element = result.render();
      element.dataset.swcKey = key;
      return element;
    },
  };
}

/** Keyed list of template results (e.g. table rows). */
export function forEach<T>(
  items: T[],
  keyFn: (item: T, index: number) => string | number,
  renderFn: (item: T, index: number) => TemplateResult,
): TemplateResult[] {
  return items.map((item, index) => keyedResult(String(keyFn(item, index)), renderFn(item, index)));
}

/** First matching branch of a condition chain. */
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
