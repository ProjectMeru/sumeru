import { describe, expect, it } from "vitest";
import { compileSumXml } from "../../src/template/sum/codegen.js";

describe("sum-template compiler", () => {
  it("compiles t-foreach and t-if", () => {
    const xml = `<div t-foreach="item in items" t-key="item.id"><span t-if="item.active">{{ item.name }}</span></div>`;
    const { code, meta } = compileSumXml(xml, "Demo", "demo.sum.xml");
    expect(code).toContain("forEach");
    expect(code).toContain("when");
    expect(meta.file).toBe("demo.sum.xml");
  });
});
