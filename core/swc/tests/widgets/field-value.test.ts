import { describe, expect, it } from "vitest";
import { SwcRecord } from "../../src/model/record.js";
import { booleanFromUnknown, recordDisplayName, stringFromUnknown } from "../../src/widgets/field-value.js";

describe("field-value", () => {
  it("booleanFromUnknown coerces common truthy values", () => {
    expect(booleanFromUnknown(true)).toBe(true);
    expect(booleanFromUnknown(1)).toBe(true);
    expect(booleanFromUnknown("1")).toBe(true);
    expect(booleanFromUnknown("true")).toBe(true);
    expect(booleanFromUnknown(false)).toBe(false);
    expect(booleanFromUnknown("false")).toBe(false);
  });

  it("stringFromUnknown stringifies or returns empty", () => {
    expect(stringFromUnknown(null)).toBe("");
    expect(stringFromUnknown(false)).toBe("");
    expect(stringFromUnknown("hello")).toBe("hello");
    expect(stringFromUnknown(42)).toBe("42");
  });

  it("recordDisplayName prefers _name suffix", () => {
    const withName = new SwcRecord("m", 1, { partner_id: 7, partner_id_name: "Acme" });
    expect(recordDisplayName(withName, "partner_id")).toBe("Acme");
    const withId = new SwcRecord("m", 1, { partner_id: 7 });
    expect(recordDisplayName(withId, "partner_id")).toBe("#7");
    const empty = new SwcRecord("m", 1, { partner_id: null });
    expect(recordDisplayName(empty, "partner_id")).toBe("");
  });
});
