import type { SumElement, SumNode } from "./parser.js";
import { parseSumXml } from "./parser.js";
import { templateMeta, type TemplateSourceMeta } from "./meta.js";

function esc(s: string): string {
  return s.replace(/\\/g, "\\\\").replace(/`/g, "\\`").replace(/\$/g, "\\$");
}

function attrToHtml(attrs: Record<string, string>): string {
  return Object.entries(attrs)
    .filter(([k]) => !k.startsWith("t-"))
    .map(([k, v]) => (v ? `${k}="${esc(v)}"` : k))
    .join(" ");
}

function codegenNode(node: SumNode, ctx: string): string {
  if (node.type === "text") return esc(node.value);
  if (node.type === "interpolation") return `\${${node.expr}}`;

  const el = node;
  const tIf = el.attrs["t-if"];
  const tElif = el.attrs["t-elif"];
  const tElse = "t-else" in el.attrs;
  const tForeach = el.attrs["t-foreach"];
  const tKey = el.attrs["t-key"];
  const tEsc = el.attrs["t-esc"];
  const tRaw = el.attrs["t-raw"];
  const tComponent = el.attrs["t-component"];
  const tModel = el.attrs["t-model"];
  const tRef = el.attrs["t-ref"];
  const tPortal = el.attrs["t-portal"];
  const tSlot = el.attrs["t-slot"];

  if (tForeach) {
    const m = tForeach.match(/^(\w+)\s+in\s+(.+)$/);
    if (!m) throw new Error(`Invalid t-foreach: ${tForeach}`);
    const [, item, collection] = m;
    const keyExpr = tKey ? tKey.replace(item, `${item}`) : `\`${item}\``;
    const inner = el.children.map((c) => codegenNode(c, ctx)).join("");
    return `\${forEach(${collection}, (${item}) => ${keyExpr}, (${item}) => html\`<${el.tag} ${attrToHtml(el.attrs)}>${inner}</${el.tag}>\`)}`;
  }

  if (tIf) {
    const inner = el.children.map((c) => codegenNode(c, ctx)).join("");
    const block = `html\`<${el.tag} ${attrToHtml(el.attrs)}>${inner}</${el.tag}>\``;
    return `\${when(${tIf}, () => ${block})}`;
  }

  if (tComponent) {
    return `\${mountComponent(${tComponent}, { ...${ctx} }, env).render()}`;
  }

  if (tEsc) return `\${${tEsc}}`;
  if (tRaw) return `\${${tRaw}}`;

  const attrs = { ...el.attrs };
  delete attrs["t-ref"];
  delete attrs["t-model"];
  delete attrs["t-portal"];
  delete attrs["t-slot"];

  let attrStr = attrToHtml(attrs);
  if (tModel) attrStr += ` @input=\${(ev) => { ${tModel} = ev.target.value; }} value=\${${tModel}}`;
  if (tRef) attrStr += ` data-ref="${esc(tRef)}"`;
  if (tPortal) attrStr += ` data-portal="${esc(tPortal)}"`;
  if (tSlot) attrStr += ` data-slot="${esc(tSlot)}"`;

  const inner = el.children.map((c) => codegenNode(c, ctx)).join("");
  const voidTag = ["img", "br", "hr", "input", "meta", "link"].includes(el.tag.toLowerCase());

  if (voidTag) return `<${el.tag} ${attrStr} />`;
  if (tElse) return inner;
  if (tElif) return inner;
  return `<${el.tag} ${attrStr}>${inner}</${el.tag}>`;
}

export interface CodegenResult {
  code: string;
  meta: TemplateSourceMeta;
}

export function codegen(template: SumElement, componentName: string, file: string): CodegenResult {
  const body = codegenNode(template, "props");
  const code = `import { html } from "../../template/html.js";
import { forEach, when } from "../../template/helpers.js";
import { mountComponent } from "../../runtime/component-host.js";
import type { SwcEnv } from "../../runtime/env.js";

export function template(props: Record<string, unknown>, env: SwcEnv) {
  return html\`${body}\`;
}
`;
  return {
    code,
    meta: templateMeta(componentName, file, { snippet: body.slice(0, 200) }),
  };
}

export function compileSumXml(source: string, componentName: string, file: string): CodegenResult {
  return codegen(parseSumXml(source), componentName, file);
}
