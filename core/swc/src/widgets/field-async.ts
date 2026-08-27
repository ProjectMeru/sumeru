import { recordDisplayName } from "./field-value.js";

export { recordDisplayName };

/** Local re-render control for field widgets — never re-renders the root app. */
export class AsyncFieldController {
  private generation = 0;

  constructor(private readonly component: { rootElement: HTMLElement | null; patch(): void }) {}

  begin(): number {
    this.generation += 1;
    return this.generation;
  }

  cancel(): void {
    this.generation += 1;
  }

  refresh(): void {
    if (this.component.rootElement?.isConnected) {
      this.component.patch();
    }
  }

  commitIfCurrent(generation: number): void {
    if (generation !== this.generation) return;
    this.refresh();
  }
}
