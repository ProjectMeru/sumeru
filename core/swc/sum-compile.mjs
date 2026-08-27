/** Load the TypeScript SUM compiler (codegen.ts) for the esbuild plugin. */

import * as esbuild from "esbuild";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));

/** Bundle `src/template/sum/codegen.ts` and return `compileSumXml`. */
export async function loadCompileSumXml() {
  const dir = mkdtempSync(join(tmpdir(), "sum-compiler-"));
  const outfile = join(dir, "codegen.mjs");
  await esbuild.build({
    entryPoints: [join(__dirname, "src/template/sum/codegen.ts")],
    bundle: true,
    format: "esm",
    platform: "node",
    outfile,
  });
  const mod = await import(pathToFileURL(outfile).href);
  return mod.compileSumXml;
}
