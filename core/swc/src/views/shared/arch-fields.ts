/**
 * Shared rules for reading view XML arch serialized as SwcViewArch JSON (see swcmeta/arch.go).
 *
 * - Collection views: `<field name="..." widget="..." invisible="1"/>` → arch.fields[]
 * - Form views: `<sheet>`, `<group>`, `<notebook>` → arch.sheet (not covered here)
 * - Kanban extras: columns_per_row, default_group_by → arch.kanban
 * - Search facets: separate search view XML → arch.search
 */
import type { SwcArchField, SwcViewArch } from "../../types/workspace.js";

export interface KanbanCardFieldOptions {
  /** When set, omit the group-by field from card body (still loaded if invisible in XML). */
  groupField?: string;
  /** Include invisible arch fields (e.g. color for picker logic). */
  includeInvisible?: boolean;
}

export interface GraphAxes {
  groupField: string;
  measureField: string;
  chart: string;
}

export interface PivotFieldGroups {
  rowFields: string[];
  colFields: string[];
  measureFields: string[];
}

export function archFields(arch: Pick<SwcViewArch, "fields">): SwcArchField[] {
  return arch.fields ?? [];
}

/** Filter a field array by arch invisible flag (form sheet groups, export lists). */
export function visibleArchFields(fields: SwcArchField[]): SwcArchField[] {
  return fields.filter((f) => !f.invisible);
}

export function visibleFields(arch: Pick<SwcViewArch, "fields">): SwcArchField[] {
  return visibleArchFields(archFields(arch));
}

export function fieldByName(arch: Pick<SwcViewArch, "fields">, name: string): SwcArchField | undefined {
  return archFields(arch).find((f) => f.name === name);
}

/** List, hierarchy, and activity table columns. */
export function listColumns(arch: Pick<SwcViewArch, "fields">): SwcArchField[] {
  return visibleFields(arch);
}

/** Kanban card arch fields; hides group field and data-only helpers from card body by default. */
export function kanbanCardFields(
  arch: Pick<SwcViewArch, "fields" | "kanban">,
  options: KanbanCardFieldOptions = {},
): SwcArchField[] {
  const groupField = options.groupField ?? arch.kanban?.groupField ?? "";
  const source = options.includeInvisible ? archFields(arch) : visibleFields(arch);
  return source.filter((f) => {
    if (groupField && f.name === groupField) return false;
    if (f.name === "gender") return false;
    return true;
  });
}

export function graphAxes(arch: Pick<SwcViewArch, "fields" | "graph">): GraphAxes {
  const fields = archFields(arch);
  return {
    chart: (arch.graph?.chart || "bar").toLowerCase(),
    groupField: fields.find((f) => f.pivotType === "row")?.name ?? "create_date",
    measureField: fields.find((f) => f.pivotType === "measure")?.name ?? "id",
  };
}

export function pivotFields(arch: Pick<SwcViewArch, "fields">): PivotFieldGroups {
  const rowFields: string[] = [];
  const colFields: string[] = [];
  const measureFields: string[] = [];
  for (const f of archFields(arch)) {
    const kind = (f.pivotType ?? "").toLowerCase();
    if (kind === "row") rowFields.push(f.name);
    else if (kind === "col" || kind === "column") colFields.push(f.name);
    else if (kind === "measure") measureFields.push(f.name);
  }
  return { rowFields, colFields, measureFields };
}
