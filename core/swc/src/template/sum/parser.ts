/** Minimal XML parser for sum-template (.sum.xml). */

export type SumNodeType = "element" | "text" | "interpolation";

export interface SumElement {
  type: "element";
  tag: string;
  attrs: Record<string, string>;
  children: SumNode[];
}

export interface SumText {
  type: "text";
  value: string;
}

export interface SumInterpolation {
  type: "interpolation";
  expr: string;
}

export type SumNode = SumElement | SumText | SumInterpolation;

export interface SumDocument {
  templates: Map<string, SumElement>;
  main: SumElement;
}

const VOID = new Set(["br", "hr", "img", "input", "meta", "link"]);

function parseSumRoot(source: string): SumElement {
  const root: SumElement = { type: "element", tag: "t", attrs: {}, children: [] };
  const stack: SumElement[] = [root];
  let i = 0;

  const skipWs = (): void => {
    while (i < source.length && /\s/.test(source[i])) i++;
  };

  const readUntil = (end: string): string => {
    let out = "";
    while (i < source.length && !source.startsWith(end, i)) {
      out += source[i++];
    }
    return out;
  };

  const parseAttrs = (raw: string): Record<string, string> => {
    const attrs: Record<string, string> = {};
    const re = /([@\w:-]+)(?:="([^"]*)")?/g;
    let m: RegExpExecArray | null;
    while ((m = re.exec(raw))) {
      attrs[m[1]] = m[2] ?? "";
    }
    return attrs;
  };

  while (i < source.length) {
    skipWs();
    if (i >= source.length) break;

    if (source.startsWith("<!--", i)) {
      i = source.indexOf("-->", i) + 3;
      continue;
    }

    if (source[i] === "<") {
      if (source.startsWith("</", i)) {
        i += 2;
        const tag = readUntil(">").trim();
        i++;
        if (stack.length > 1 && stack[stack.length - 1].tag === tag) stack.pop();
        continue;
      }

      i++;
      const selfClose = source.indexOf("/>", i);
      const close = source.indexOf(">", i);
      const end = selfClose !== -1 && selfClose < close ? selfClose : close;
      const inner = source.slice(i, end).trim();
      i = end + (source[end] === "/" ? 2 : 1);

      const tagMatch = inner.match(/^([^\s/]+)/);
      const tag = tagMatch?.[1] ?? "div";
      const attrRaw = inner.slice(tag.length);
      const el: SumElement = { type: "element", tag, attrs: parseAttrs(attrRaw), children: [] };
      stack[stack.length - 1].children.push(el);

      const isSelfClosing = selfClose !== -1 && selfClose < close;
      if (!isSelfClosing && !VOID.has(tag.toLowerCase())) {
        stack.push(el);
      }
      continue;
    }

    if (source.startsWith("{{", i)) {
      i += 2;
      const expr = readUntil("}}").trim();
      i += 2;
      stack[stack.length - 1].children.push({ type: "interpolation", expr });
      continue;
    }

    const textStart = i;
    while (i < source.length && source[i] !== "<" && !source.startsWith("{{", i)) i++;
    const value = source.slice(textStart, i).replace(/\s+/g, " ");
    if (value.trim()) {
      stack[stack.length - 1].children.push({ type: "text", value: value.trim() });
    }
  }

  return root;
}

export function parseSumDocument(source: string): SumDocument {
  const root = parseSumRoot(source);
  const templates = new Map<string, SumElement>();
  let main: SumElement | undefined;

  for (const child of root.children) {
    if (child.type !== "element") continue;
    const tName = child.attrs["t-name"]?.trim();
    if (tName) {
      if (templates.has(tName)) {
        throw new Error(`sum-template: duplicate t-name ${tName}`);
      }
      templates.set(tName, child);
      continue;
    }
    if (main) {
      throw new Error("sum-template: multiple root templates (use t-name for extras)");
    }
    main = child;
  }

  if (!main) {
    throw new Error("sum-template: missing root element");
  }
  return { templates, main };
}
