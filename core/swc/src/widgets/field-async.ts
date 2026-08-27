/** Local re-render control for field widgets — never triggers root app schedulePatch. */
export class AsyncFieldController {
  private generation = 0;

  constructor(private readonly comp: { el: HTMLElement | null; patch(): void }) {}

  begin(): number {
    this.generation += 1;
    return this.generation;
  }

  cancel(): void {
    this.generation += 1;
  }

  refresh(): void {
    if (this.comp.el?.parentElement) {
      this.comp.patch();
    }
  }

  finish(gen: number): void {
    if (gen !== this.generation) return;
    this.refresh();
  }
}

export function recordDisplayName(record: { get: (f: string) => unknown }, fieldName: string): string {
  const named = record.get(`${fieldName}_name`);
  if (named != null && named !== "") return String(named);
  const raw = record.get(fieldName);
  if (raw == null || raw === "") return "";
  return `#${raw}`;
}
