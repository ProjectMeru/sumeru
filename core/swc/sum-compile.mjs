/** Plain JS sum-template compiler for esbuild plugin (no TS import). */

export function compileSumXml(source, componentName, file) {
  const body = source
    .replace(/<\?xml[^?]*\?>/g, "")
    .replace(/t-esc="([^"]+)"/g, "${$1}")
    .replace(/t-if="([^"]+)"/g, "")
    .trim();
  const code = `import { html } from "../../template/html.js";
import { forEach, when } from "../../template/helpers.js";
import { mountComponent } from "../../runtime/component-host.js";
import type { SwcEnv } from "../../runtime/env.js";

export function template(props, env) {
  return html\`${body.replace(/`/g, "\\`").replace(/\$/g, "\\$")}\`;
}
`;
  return { code, meta: { component: componentName, file, snippet: body.slice(0, 200) } };
}
