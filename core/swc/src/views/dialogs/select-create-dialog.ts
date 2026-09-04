import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcEnv } from "../../runtime/env.js";

export interface SelectCreateOptions {
  comodel: string;
  title?: string;
  domain?: unknown[];
  initialQuery?: string;
  onSelect: (row: Record<string, unknown>) => void;
  onCancel?: () => void;
}

/** Modal search/create picker for many2one and many2many (Enterprise SelectCreate equivalent). */
export class SelectCreateDialog extends SwcComponent<SelectCreateOptions> {
  private query = "";
  private rows: Record<string, unknown>[] = [];
  private loading = false;
  private onKeydownBound = (event: KeyboardEvent): void => {
    if (event.key === "Escape") {
      event.preventDefault();
      this.close();
    }
  };

  static open(env: SwcEnv, opts: SelectCreateOptions): void {
    const host = document.createElement("div");
    host.className = "sum-modal-host";
    document.body.appendChild(host);
    const dialog = new SelectCreateDialog(opts, env);
    dialog.query = opts.initialQuery?.trim() ?? "";
    dialog.callSetup();
    host.appendChild(dialog.render());
  }

  override onMount(): void {
    document.addEventListener("keydown", this.onKeydownBound);
    void this.search(this.query).then(() => {
      const input = this.rootElement?.querySelector<HTMLInputElement>(".sum-field-input");
      input?.focus();
    });
  }

  override onWillUnmount(): void {
    document.removeEventListener("keydown", this.onKeydownBound);
  }

  private close(): void {
    this.props.onCancel?.();
    const host = this.rootElement?.closest(".sum-modal-host");
    this.destroy();
    host?.remove();
  }

  private async search(q: string): Promise<void> {
    this.loading = true;
    this.rerender();
    const domain = [...(this.props.domain ?? [])];
    if (q.trim()) domain.push(["name", "ilike", q.trim()]);
    this.rows =
      (await this.env.services.rpc.searchRead(this.props.comodel, domain, ["id", "name"], 40)) ?? [];
    this.loading = false;
    this.rerender();
  }

  private async createNew(): Promise<void> {
    const name = this.query.trim() || "New";
    const id = await this.env.services.rpc.create(this.props.comodel, { name });
    if (typeof id === "number" && id > 0) {
      this.props.onSelect({ id, name });
      const host = this.rootElement?.closest(".sum-modal-host");
      this.destroy();
      host?.remove();
    }
  }

  private pick(row: Record<string, unknown>): void {
    this.props.onSelect(row);
    const host = this.rootElement?.closest(".sum-modal-host");
    this.destroy();
    host?.remove();
  }

  override template() {
    const title = this.props.title ?? "Select record";
    return html`<div class="sum-modal-backdrop" @click=${() => this.close()}>
      <div class="sum-modal sum-select-create" @click=${(e: Event) => e.stopPropagation()}>
        <header class="sum-modal-header">
          <h3>${title}</h3>
          <button type="button" class="sum-modal-close" @click=${() => this.close()}>×</button>
        </header>
        <div class="sum-modal-body">
          <input
            class="sum-field-input"
            placeholder="Search..."
            value=${this.query}
            @input=${(e: Event) => {
              this.query = (e.target as HTMLInputElement).value;
              void this.search(this.query);
            }}
          />
          ${this.loading
            ? html`<p class="sum-muted">Loading…</p>`
            : this.rows.length === 0
              ? html`<p class="sum-select-create-empty">No records found.</p>`
              : html`<ul class="sum-select-create-list">
                  ${this.rows.map(
                    (row) => html`<li>
                      <button type="button" class="sum-select-create-item" @click=${() => this.pick(row)}>
                        ${String(row.name ?? row.id)}
                      </button>
                    </li>`,
                  )}
                </ul>`}
        </div>
        <footer class="sum-modal-footer">
          <button type="button" class="sum-btn" @click=${() => void this.createNew()}>Create</button>
          <button type="button" class="sum-btn" @click=${() => this.close()}>Cancel</button>
        </footer>
      </div>
    </div>`;
  }
}
