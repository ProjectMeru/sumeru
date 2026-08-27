import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import { useState } from "../../runtime/hooks.js";
import { RECORD_UPDATED, SWC_API_BASE } from "../../constants/routes.js";

interface ChatterMessage {
  body: string;
  author: string;
  createDate: string;
}

interface ChatterAttachment {
  id: number;
  name: string;
  url: string;
  mimetype?: string;
}

interface ChatterPayload {
  model: string;
  recordId: number;
  messages: ChatterMessage[];
  attachments?: ChatterAttachment[];
  enabled: boolean;
}

interface ChatterPanelProps {
  model: string;
  recordId: number;
  csrfToken: string;
}

export class ChatterPanel extends SwcComponent<ChatterPanelProps> {
  private messages: ChatterMessage[] = [];
  private attachments: ChatterAttachment[] = [];
  private draft = "";
  private loading = true;
  private posting = false;
  private enabled = true;
  private tab: "messages" | "attachments" = "messages";

  setup(): void {
    const [, bump] = useState(0);
    this.bump = () => bump((n) => n + 1);
    void this.load();
  }

  private bump: (() => void) | null = null;

  private async load(): Promise<void> {
    const { model, recordId } = this.props;
    if (recordId <= 0) {
      this.loading = false;
      this.bump?.();
      return;
    }
    this.loading = true;
    this.bump?.();
    try {
      const base = this.env.bootstrap.swcApiBase || SWC_API_BASE;
      const data = await this.env.services.http.getJSON<ChatterPayload>(
        `${base}/chatter?model=${encodeURIComponent(model)}&id=${recordId}`,
      );
      this.messages = data.messages ?? [];
      this.attachments = data.attachments ?? [];
      this.enabled = data.enabled !== false;
    } finally {
      this.loading = false;
      this.bump?.();
    }
  }

  private async post(): Promise<void> {
    const body = this.draft.trim();
    if (!body || this.props.recordId <= 0) return;
    this.posting = true;
    this.bump?.();
    try {
      await this.env.services.http.postForm("/web/chatter/post", {
        model: this.props.model,
        res_id: String(this.props.recordId),
        body,
        next: window.location.pathname + window.location.search,
      });
      this.draft = "";
      await this.load();
      this.env.services.bus.emit(RECORD_UPDATED, {
        model: this.props.model,
        id: this.props.recordId,
      });
    } finally {
      this.posting = false;
      this.bump?.();
    }
  }

  template() {
    if (this.props.recordId <= 0) {
      return html`<aside class="sum-chatter sum-chatter--empty">Save the record to post messages.</aside>`;
    }
    if (!this.enabled) return html``;
    if (this.loading) {
      return html`<aside class="sum-chatter sum-chatter--loading">Loading messages…</aside>`;
    }
    return html`
      <aside class="sum-chatter">
        <div class="sum-chatter-tabs">
          <button type="button" class="sum-chatter-tab${this.tab === "messages" ? " sum-chatter-tab--active" : ""}" @click=${() => { this.tab = "messages"; this.bump?.(); }}>Messages</button>
          <button type="button" class="sum-chatter-tab${this.tab === "attachments" ? " sum-chatter-tab--active" : ""}" @click=${() => { this.tab = "attachments"; this.bump?.(); }}>Attachments (${this.attachments.length})</button>
        </div>
        ${this.tab === "attachments"
          ? html`<ul class="sum-chatter-attachments">
              ${this.attachments.length === 0
                ? html`<li class="sum-chatter-empty">No attachments.</li>`
                : this.attachments.map(
                    (a) => html`<li><a href=${a.url} target="_blank" rel="noopener">${a.name}</a></li>`,
                  )}
            </ul>`
          : html`
        <div class="sum-chatter-composer">
          <textarea
            class="sum-chatter-input"
            placeholder="Write a message…"
            rows="3"
            value=${this.draft}
            @input=${(ev: Event) => {
              this.draft = (ev.target as HTMLTextAreaElement).value;
              this.bump?.();
            }}
          ></textarea>
          <button
            type="button"
            class="sum-btn sum-btn--primary sum-chatter-send"
            disabled=${this.posting ? "disabled" : undefined}
            @click=${() => void this.post()}
          >
            Post
          </button>
        </div>
        <ul class="sum-chatter-messages">
          ${this.messages.length === 0
            ? html`<li class="sum-chatter-empty">No messages yet.</li>`
            : this.messages.map(
                (m) => html`<li class="sum-chatter-message">
                  <div class="sum-chatter-meta">${m.author} · ${m.createDate}</div>
                  <div class="sum-chatter-body">${m.body}</div>
                </li>`,
              )}
        </ul>`}
      </aside>
    `;
  }
}
