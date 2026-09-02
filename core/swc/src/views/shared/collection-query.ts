import type { SwcEnv } from "../../runtime/env.js";
import type { SwcArchField, SwcSearchFilter, SwcSearchMeta, SwcWorkspacePayload } from "../../types/workspace.js";
import { parseFilterCSV, toggleFilterName } from "../list/control-panel.js";

export interface CollectionQuery {
  search: string;
  presetFilters: string[];
  customDomain: string;
  groupBy: string[];
}

export interface CollectionQueryPatch {
  search?: string;
  presetFilters?: string[];
  customDomain?: string;
  groupBy?: string[];
  listSort?: string;
  listOffset?: number;
}

export function parseGroupByCSV(raw?: string): string[] {
  return parseFilterCSV(raw);
}

export function syncCollectionQuery(payload: SwcWorkspacePayload): CollectionQuery {
  return {
    search: payload.listSearch ?? "",
    presetFilters: parseFilterCSV(payload.listFilter),
    customDomain: payload.listDomain ?? "",
    groupBy: parseGroupByCSV(payload.listGroupBy),
  };
}

export function navigateCollectionQuery(
  env: SwcEnv,
  payload: SwcWorkspacePayload,
  viewType: string,
  patch: CollectionQueryPatch,
): void {
  const current = syncCollectionQuery(payload);
  const next: CollectionQuery = {
    search: patch.search ?? current.search,
    presetFilters: patch.presetFilters ?? current.presetFilters,
    customDomain: patch.customDomain ?? current.customDomain,
    groupBy: patch.groupBy ?? current.groupBy,
  };
  env.services.action.navigate(
    env.services.router.workspaceUrl({
      actionId: payload.actionId,
      menuId: payload.menuId,
      viewType,
      listSearch: next.search,
      listFilter: next.presetFilters.join(","),
      listDomain: next.customDomain,
      listGroupBy: next.groupBy.join(","),
      listSort: patch.listSort ?? payload.listSort ?? "",
      listOffset: patch.listOffset ?? payload.listOffset ?? 0,
      model: payload.actionId ? "" : payload.model,
    }),
  );
}

/** Mutually exclusive CRM pipeline facets (type and won/lost). */
const PRESET_EXCLUSIVE_GROUPS: string[][] = [
  ["opportunity", "lead"],
  ["won", "lost"],
];

export function togglePresetFilter(filters: string[], name: string): string[] {
  if (filters.includes(name)) {
    return filters.filter((n) => n !== name);
  }
  const group = PRESET_EXCLUSIVE_GROUPS.find((g) => g.includes(name));
  const withoutGroup = group ? filters.filter((n) => !group.includes(n)) : filters;
  return [...withoutGroup, name];
}

export function toggleGroupByField(active: string[], field: string): string[] {
  return toggleFilterName(active, field);
}

export interface FilterTag {
  key: string;
  label: string;
  kind: "preset" | "domain" | "group";
}

function fieldLabel(name: string, search?: SwcSearchMeta): string {
  return (
    search?.filterFields?.find((f) => f.name === name)?.string
    ?? search?.groupByFields?.find((f) => f.name === name)?.string
    ?? name
  );
}

export function formatDomainTripleLabel(
  triple: DomainTriple,
  filterFields?: SwcArchField[],
): string {
  const [field, op, value] = triple;
  const label = filterFields?.find((f) => f.name === field)?.string ?? field;
  return `${label} ${op} ${String(value)}`;
}

export function activeFilterTags(query: CollectionQuery, search?: SwcSearchMeta): FilterTag[] {
  const tags: FilterTag[] = [];
  const presets = search?.filters ?? [];
  for (const name of query.presetFilters) {
    const preset = presets.find((f) => f.name === name);
    tags.push({ key: `preset:${name}`, label: preset?.string ?? name, kind: "preset" });
  }
  const triples = parseDomainJSON(query.customDomain);
  for (let i = 0; i < triples.length; i++) {
    tags.push({
      key: `domain:${i}`,
      label: formatDomainTripleLabel(triples[i], search?.filterFields),
      kind: "domain",
    });
  }
  for (const field of query.groupBy) {
    const gb =
      presets.find((f) => f.groupBy === field)?.string
      ?? fieldLabel(field, search);
    tags.push({ key: `group:${field}`, label: `Group: ${gb}`, kind: "group" });
  }
  return tags;
}

export function filterCount(query: CollectionQuery): number {
  return query.presetFilters.length + parseDomainJSON(query.customDomain).length;
}

export function groupByCount(query: CollectionQuery): number {
  return query.groupBy.length;
}

export function presetDomainFilters(filters: SwcSearchFilter[] = []): SwcSearchFilter[] {
  return filters.filter((f) => f.domain);
}

export function presetGroupByFilters(filters: SwcSearchFilter[] = []): SwcSearchFilter[] {
  return filters.filter((f) => f.groupBy);
}

export type DomainTriple = [string, string, unknown];

export function parseDomainJSON(raw: string): DomainTriple[] {
  const text = raw.trim();
  if (!text || text === "[]") return [];
  try {
    const parsed = JSON.parse(text) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((row): row is DomainTriple =>
      Array.isArray(row) && row.length === 3 && typeof row[0] === "string" && typeof row[1] === "string",
    );
  } catch {
    return [];
  }
}

export function stringifyDomain(triples: DomainTriple[]): string {
  if (triples.length === 0) return "";
  return JSON.stringify(triples);
}

export function appendDomainTriple(raw: string, triple: DomainTriple): string {
  const next = [...parseDomainJSON(raw), triple];
  return stringifyDomain(next);
}

export function removeDomainTriple(raw: string, index: number): string {
  const triples = parseDomainJSON(raw);
  if (index < 0 || index >= triples.length) return raw;
  triples.splice(index, 1);
  return stringifyDomain(triples);
}

export function removeFilterTag(query: CollectionQuery, tag: FilterTag): CollectionQuery {
  if (tag.kind === "preset") {
    const name = tag.key.replace(/^preset:/, "");
    return { ...query, presetFilters: query.presetFilters.filter((n) => n !== name) };
  }
  if (tag.kind === "domain") {
    const index = Number(tag.key.replace(/^domain:/, ""));
    return { ...query, customDomain: removeDomainTriple(query.customDomain, index) };
  }
  const field = tag.key.replace(/^group:/, "");
  return { ...query, groupBy: query.groupBy.filter((f) => f !== field) };
}

const FILTER_OPERATORS_BY_TYPE: Record<string, string[]> = {
  char: ["=", "!=", "ilike"],
  text: ["=", "!=", "ilike"],
  integer: ["=", "!=", ">", "<", ">=", "<="],
  float: ["=", "!=", ">", "<", ">=", "<="],
  boolean: ["=", "!="],
  date: ["=", "!=", ">", "<", ">=", "<="],
  datetime: ["=", "!=", ">", "<", ">=", "<="],
  selection: ["=", "!="],
  many2one: ["=", "!="],
};

export function filterOperatorsForField(fieldName: string, filterFields: SwcArchField[]): string[] {
  const field = filterFields.find((f) => f.name === fieldName);
  return FILTER_OPERATORS_BY_TYPE[field?.type ?? "char"] ?? ["=", "!="];
}

/** Coerce a custom filter input string to the appropriate type for the field. */
export function coerceFilterValue(fieldName: string, raw: string, filterFields: SwcArchField[]): unknown {
  const field = filterFields.find((f) => f.name === fieldName);
  const trimmed = raw.trim();
  if (field?.type === "boolean") return trimmed === "true" || trimmed === "1";
  if (field?.type === "integer" || field?.type === "float") {
    const n = Number(trimmed);
    return Number.isNaN(n) ? trimmed : n;
  }
  return trimmed;
}
