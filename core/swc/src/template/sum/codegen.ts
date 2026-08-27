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

function elementBlock(el: SumElement, ctx: string): string {
  const attrs = { ...el.attrs };
  delete attrs["t-if"];
  delete attrs["t-elif"];
  delete attrs["t-else"];
  delete attrs["t-foreach"];
  delete attrs["t-key"];
  return `html\`<${el.tag} ${attrToHtml(attrs)}>${codegenChildren(el.children, ctx)}</${el.tag}>\``;
}

function codegenChildren(nodes: SumNode[], ctx: string): string {
  let out = "";
  let i = 0;
  while (i < nodes.length) {
    const node = nodes[i];
    if (node.type === "element" && node.attrs["t-if"]) {
      const branches: SumElement[] = [node];
      i += 1;
      while (i < nodes.length) {
        const next = nodes[i];
        if (next.type !== "element") break;
        if ("t-elif" in next.attrs) {
          branches.push(next);
          i += 1;
          continue;
        }
        if ("t-else" in next.attrs) {
          branches.push(next);
          i += 1;
        }
        break;
      }
      const first = branches[0];
      const elifs = branches.slice(1).map((branch) => {
        const cond = "t-else" in branch.attrs ? "true" : branch.attrs["t-elif"];
        return `[${cond}, () => ${elementBlock(branch, ctx)}]`;
      });
      out += `\${when(${first.attrs["t-if"]}, () => ${elementBlock(first, ctx)}${
        elifs.length ? `, ${elifs.join(", ")}` : ""
      })}`;
      continue;
    }
    out += codegenNode(node, ctx);
    i += 1;
  }
  return out;
}

function codegenNode(node: SumNode, ctx: string): string {
  if (node.type === "text") return esc(node.value);
  if (node.type === "interpolation") return `\${${node.expr}}`;

  const el = node;
  const tForeach = el.attrs["t-foreach"];
  const tKey = el.attrs["t-key"];
  const tEsc = el.attrs["t-esc"];
  const tRaw = el.attrs["t-raw"] ?? el.attrs["t-out"];
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
    const inner = codegenChildren(el.children, ctx);
    return `\${forEach(${collection}, (${item}) => ${keyExpr}, (${item}) => html\`<${el.tag} ${attrToHtml(el.attrs)}>${inner}</${el.tag}>\`)}`;
  }

  if (el.attrs["t-if"]) {
    return codegenChildren([el], ctx);
  }

  if (tComponent) {
    return `\${mountComponent(${tComponent}, { ...${ctx} }, env)}`;
  }

  if (tEsc) return `\${${tEsc}}`;
  if (tRaw) return `\${${tRaw}}`;

  const attrs = { ...el.attrs };
  delete attrs["t-ref"];
  delete attrs["t-model"];
  delete attrs["t-portal"];
  delete attrs["t-slot"];
  delete attrs["t-out"];
  delete attrs["t-raw"];

  let attrStr = attrToHtml(attrs);
  if (tModel) {
    attrStr += ` @input=\${(event) => { ${tModel} = inputValueFromEvent(event); }} value=\${${tModel}}`;
  }
  if (tRef) attrStr += ` data-ref="${esc(tRef)}"`;
  if (tPortal) attrStr += ` data-portal="${esc(tPortal)}"`;
  if (tSlot) attrStr += ` data-slot="${esc(tSlot)}"`;

  const inner = codegenChildren(el.children, ctx);
  const voidTag = ["img", "br", "hr", "input", "meta", "link"].includes(el.tag.toLowerCase());

  if (voidTag) return `<${el.tag} ${attrStr} />`;
  return `<${el.tag} ${attrStr}>${inner}</${el.tag}>`;
}

export interface CodegenResult {
  code: string;
  meta: TemplateSourceMeta;
}

export function codegen(template: SumElement, componentName: string, file: string): CodegenResult {
  const body = codegenChildren(template.children.length ? template.children : [template], "props");
  const wrapped = template.tag === "t" ? body : codegenNode(template, "props");
  const code = `import { html } from "../../template/html.js";
import { forEach, when } from "../../template/helpers.js";
import { mountComponent } from "../../runtime/component-host.js";
import { inputValueFromEvent } from "../../widgets/field-events.js";
import type { SwcEnv } from "../../runtime/env.js";

export function template(props: Record<string, unknown>, env: SwcEnv) {
  return html\`${wrapped}\`;
}
`;
  return {
    code,
    meta: templateMeta(componentName, file, { snippet: wrapped.slice(0, 200) }),
  };
}

export function compileSumXml(source: string, componentName: string, file: string): CodegenResult {
  return codegen(parseSumXml(source), componentName, file);
}
