import * as esbuild from "esbuild";
import { mkdirSync, readFileSync } from "node:fs";
import { dirname, join, basename } from "node:path";
import { fileURLToPath } from "node:url";
import { compileSumXml } from "./sum-compile.mjs";

const __dirname = dirname(fileURLToPath(import.meta.url));
const swcOutFile = join(__dirname, "../engine/assets/swc/swc.js");
const passwordToggleOutFile = join(__dirname, "../engine/assets/js/sumeru-password-toggle.js");
const passwordMatchOutFile = join(__dirname, "../engine/assets/js/sumeru-password-match.js");

mkdirSync(dirname(swcOutFile), { recursive: true });
mkdirSync(dirname(passwordMatchOutFile), { recursive: true });

const sumXmlPlugin = {
  name: "sum-xml",
  setup(build) {
    build.onResolve({ filter: /\.sum\.xml$/ }, (args) => ({
      path: join(args.resolveDir, args.path),
      namespace: "sum-xml",
    }));
    build.onLoad({ filter: /.*/, namespace: "sum-xml" }, (args) => {
      const source = readFileSync(args.path, "utf8");
      const name = basename(args.path, ".sum.xml");
      const { code } = compileSumXml(source, name, args.path);
      return { contents: code, loader: "js" };
    });
  },
};

await esbuild.build({
  entryPoints: [join(__dirname, "src/main.ts")],
  bundle: true,
  format: "iife",
  globalName: "SumeruSWC",
  outfile: swcOutFile,
  target: "es2022",
  sourcemap: true,
  minify: process.env.NODE_ENV === "production",
  define: {
    "process.env.NODE_ENV": JSON.stringify(process.env.NODE_ENV ?? "development"),
  },
  plugins: [sumXmlPlugin],
});

await esbuild.build({
  entryPoints: [join(__dirname, "src/login/password-toggle.ts")],
  bundle: true,
  format: "iife",
  outfile: passwordToggleOutFile,
  target: "es2022",
  sourcemap: true,
  minify: process.env.NODE_ENV === "production",
});

await esbuild.build({
  entryPoints: [join(__dirname, "src/login/password-match-entry.ts")],
  bundle: true,
  format: "iife",
  outfile: passwordMatchOutFile,
  target: "es2022",
  sourcemap: true,
  minify: process.env.NODE_ENV === "production",
});

console.log(`SWC bundle → ${swcOutFile}`);
console.log(`Login password toggle → ${passwordToggleOutFile}`);
console.log(`Login password match → ${passwordMatchOutFile}`);
