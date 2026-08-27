import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { compileSumXml } from "../../src/template/sum/codegen.js";

const fixtureDir = dirname(fileURLToPath(import.meta.url));

describe("sum-template compiler", () => {
  it("compiles t-foreach and t-if", () => {
    const xml = `<div t-foreach="item in items" t-key="item.id"><span t-if="item.active">{{ item.name }}</span></div>`;
    const { code, meta } = compileSumXml(xml, "Demo", "demo.sum.xml");
    expect(code).toContain("forEach");
    expect(code).toContain("when");
    expect(meta.file).toBe("demo.sum.xml");
  });

  it("compiles t-elif, t-else, t-out, and t-model from a fixture", () => {
    const source = readFileSync(join(fixtureDir, "fixtures/branch-out.sum.xml"), "utf8");
    const { code } = compileSumXml(source, "BranchOut", "branch-out.sum.xml");
    expect(code).toContain("when(n === 1");
    expect(code).toContain("[n === 2,");
    expect(code).toContain("[true,");
    expect(code).toContain("${markup}");
    expect(code).toContain("inputValueFromEvent");
  });
});
