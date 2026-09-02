import type { SwcArchField } from "../types/workspace.js";

type SearchReadFn = (
  model: string,
  domain: unknown[][],
  fields: string[],
  limit: number,
) => Promise<Record<string, unknown>[] | null | undefined>;

/** resolveM2oDisplayNames fills `${field}_name` on rows for many2one columns. */
export async function resolveM2oDisplayNames(
  cols: SwcArchField[],
  rows: Record<string, unknown>[],
  comodelForField: (col: SwcArchField) => string,
  searchRead: SearchReadFn,
): Promise<void> {
  for (const col of cols) {
    if (col.type !== "many2one") continue;
    const comodel = comodelForField(col);
    if (!comodel) continue;
    const ids = rows
      .map((r) => Number(r[col.name]))
      .filter((id) => Number.isFinite(id) && id > 0);
    if (ids.length === 0) continue;
    const uniq = Array.from(new Set(ids));
    let refs: Record<string, unknown>[] = [];
    try {
      refs = (await searchRead(comodel, [["id", "in", uniq]], ["id", "name"], uniq.length + 1)) ?? [];
    } catch {
      continue;
    }
    const nameById = new Map<number, string>();
    for (const ref of refs) {
      nameById.set(Number(ref.id), String(ref.name ?? ""));
    }
    for (const r of rows) {
      const id = Number(r[col.name]);
      const name = nameById.get(id);
      if (name !== undefined) r[`${col.name}_name`] = name;
    }
  }
}
