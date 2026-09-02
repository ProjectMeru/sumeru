import { describe, expect, it } from "vitest";
import {
  activeFilterTags,
  appendDomainTriple,
  coerceFilterValue,
  filterOperatorsForField,
  parseDomainJSON,
  parseGroupByCSV,
  removeDomainTriple,
  removeFilterTag,
  syncCollectionQuery,
  toggleGroupByField,
  togglePresetFilter,
} from "../../src/views/shared/collection-query.js";
import type { SwcWorkspacePayload } from "../../src/types/workspace.js";

function payload(overrides: Partial<SwcWorkspacePayload> = {}): SwcWorkspacePayload {
  return {
    actionId: 1,
    menuId: "2",
    viewType: "list",
    model: "demo.model",
    recordId: 0,
    formEdit: false,
    csrfToken: "",
    arch: {
      type: "list",
      model: "demo.model",
      fields: [],
      search: {
        filters: [
          { name: "draft", string: "Draft", domain: '[["state","=","draft"]]' },
          { name: "group_state", string: "Status", groupBy: "state" },
        ],
        filterFields: [{ name: "priority", string: "Priority", type: "integer" }],
        groupByFields: [
          { name: "state", string: "Status", type: "selection" },
          { name: "user_id", string: "User", type: "many2one" },
        ],
      },
    },
    viewTabs: [],
    breadcrumbs: [],
    listSearch: "alpha",
    listFilter: "draft",
    listDomain: '[["priority",">",1],["name","=","x"]]',
    listGroupBy: "state,user_id",
    ...overrides,
  };
}

describe("collection-query", () => {
  it("syncs query fields from payload", () => {
    const q = syncCollectionQuery(payload());
    expect(q.search).toBe("alpha");
    expect(q.presetFilters).toEqual(["draft"]);
    expect(q.customDomain).toContain("priority");
    expect(q.groupBy).toEqual(["state", "user_id"]);
  });

  it("parses group-by CSV", () => {
    expect(parseGroupByCSV("state, user_id")).toEqual(["state", "user_id"]);
    expect(parseGroupByCSV("")).toEqual([]);
  });

  it("toggles preset filters", () => {
    expect(togglePresetFilter(["draft"], "my")).toEqual(["draft", "my"]);
    expect(togglePresetFilter(["draft", "my"], "draft")).toEqual(["my"]);
  });

  it("enforces exclusive CRM type and won/lost facets", () => {
    expect(togglePresetFilter(["opportunity"], "lead")).toEqual(["lead"]);
    expect(togglePresetFilter(["lead", "my"], "opportunity")).toEqual(["my", "opportunity"]);
    expect(togglePresetFilter(["won"], "lost")).toEqual(["lost"]);
    expect(togglePresetFilter(["won", "my"], "lost")).toEqual(["my", "lost"]);
  });

  it("toggles group-by fields", () => {
    expect(toggleGroupByField(["state"], "user_id")).toEqual(["state", "user_id"]);
    expect(toggleGroupByField(["state", "user_id"], "state")).toEqual(["user_id"]);
  });

  it("builds active filter tags with per-domain and per-group entries", () => {
    const q = syncCollectionQuery(payload());
    const tags = activeFilterTags(q, payload().arch.search);
    expect(tags.some((t) => t.kind === "preset" && t.label === "Draft")).toBe(true);
    expect(tags.filter((t) => t.kind === "domain").length).toBe(2);
    expect(tags.some((t) => t.kind === "domain" && t.label === "Priority > 1")).toBe(true);
    expect(tags.filter((t) => t.kind === "group").length).toBe(2);
  });

  it("parses and appends domain triples", () => {
    expect(parseDomainJSON("")).toEqual([]);
    const merged = appendDomainTriple("", ["name", "=", "x"]);
    expect(parseDomainJSON(merged)).toEqual([["name", "=", "x"]]);
  });

  it("removes individual domain triple by index", () => {
    const raw = '[["a","=","1"],["b","=","2"]]';
    expect(removeDomainTriple(raw, 0)).toBe('[["b","=","2"]]');
    const q = syncCollectionQuery(payload());
    const tag = activeFilterTags(q, payload().arch.search).find((t) => t.key === "domain:0")!;
    const next = removeFilterTag(q, tag);
    expect(parseDomainJSON(next.customDomain).length).toBe(1);
  });

  it("removes individual group-by field", () => {
    const q = syncCollectionQuery(payload());
    const tag = activeFilterTags(q, payload().arch.search).find((t) => t.key === "group:state")!;
    const next = removeFilterTag(q, tag);
    expect(next.groupBy).toEqual(["user_id"]);
  });

  it("coerces filter values by field type", () => {
    const fields = payload().arch.search!.filterFields!;
    expect(coerceFilterValue("priority", "3", fields)).toBe(3);
    expect(coerceFilterValue("priority", "x", fields)).toBe("x");
    expect(coerceFilterValue("active", "true", [{ name: "active", type: "boolean" }])).toBe(true);
  });

  it("returns operators for filter field types", () => {
    const fields = payload().arch.search!.filterFields!;
    expect(filterOperatorsForField("priority", fields)).toContain(">");
    expect(filterOperatorsForField("missing", fields)).toContain("=");
  });
});
