import type { SwcNumberFormat } from "../types/bootstrap.js";

const DEFAULT_NUMBER_FORMAT: SwcNumberFormat = {
  decimalPoint: ".",
  thousandsSep: ",",
  grouping: "[3,0]",
};

let numberFormat = DEFAULT_NUMBER_FORMAT;

export function configureNumberFormat(config?: Partial<SwcNumberFormat>): void {
  numberFormat = {
    decimalPoint: config?.decimalPoint || DEFAULT_NUMBER_FORMAT.decimalPoint,
    thousandsSep: config?.thousandsSep || DEFAULT_NUMBER_FORMAT.thousandsSep,
    grouping: config?.grouping || DEFAULT_NUMBER_FORMAT.grouping,
  };
}

function groupingSizes(grouping: string): number[] {
  try {
    const parsed = JSON.parse(grouping);
    if (Array.isArray(parsed)) {
      const sizes = parsed.filter((size): size is number => Number.isInteger(size) && size > 0);
      if (sizes.length) return sizes;
    }
  } catch {
    // Fall back to international grouping for invalid language data.
  }
  return [3];
}

function groupInteger(value: string, sizes: number[], separator: string): string {
  const parts: string[] = [];
  let end = value.length;
  let sizeIndex = 0;
  while (end > 0) {
    const size = sizes[Math.min(sizeIndex, sizes.length - 1)];
    const start = Math.max(0, end - size);
    parts.unshift(value.slice(start, end));
    end = start;
    sizeIndex += 1;
  }
  return parts.join(separator);
}

export function formatNumericValue(raw: unknown): string {
  const value = Number(raw);
  if (!Number.isFinite(value)) return raw == null ? "" : String(raw);

  const [integerWithSign, fraction] = String(value).split(".");
  const negative = integerWithSign.startsWith("-");
  const integer = negative ? integerWithSign.slice(1) : integerWithSign;
  const grouped = groupInteger(integer, groupingSizes(numberFormat.grouping), numberFormat.thousandsSep);
  return `${negative ? "-" : ""}${grouped}${fraction === undefined ? "" : numberFormat.decimalPoint + fraction}`;
}