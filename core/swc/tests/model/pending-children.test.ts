import { describe, expect, it } from "vitest";
import { SwcRecord } from "../../src/model/record.js";
import {
  getPendingChildren,
  setPendingChildren,
  takePendingChildren,
  type PendingChildRecord,
} from "../../src/model/pending-children.js";

describe("pending-children", () => {
  it("stores and retrieves staged o2m lines", () => {
    const parent = new SwcRecord("parent.model", 0, {});
    const child: PendingChildRecord = {
      fieldName: "line_ids",
      comodel: "child.model",
      inverse: "parent_id",
      values: { name: "Line" },
    };
    setPendingChildren(parent, "line_ids", [child]);
    expect(getPendingChildren(parent, "line_ids")).toEqual([child]);
  });

  it("clears field when children array empty", () => {
    const parent = new SwcRecord("parent.model", 0, {});
    setPendingChildren(parent, "line_ids", []);
    expect(getPendingChildren(parent, "line_ids")).toBeUndefined();
  });

  it("takePendingChildren drains all fields", () => {
    const parent = new SwcRecord("parent.model", 0, {});
    setPendingChildren(parent, "a_ids", [
      { fieldName: "a_ids", comodel: "a", inverse: "p", values: {} },
    ]);
    setPendingChildren(parent, "b_ids", [
      { fieldName: "b_ids", comodel: "b", inverse: "p", values: {} },
    ]);
    const all = takePendingChildren(parent);
    expect(all).toHaveLength(2);
    expect(takePendingChildren(parent)).toEqual([]);
  });
});
