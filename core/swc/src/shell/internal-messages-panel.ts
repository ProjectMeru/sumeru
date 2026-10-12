import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { SWC_API_BASE, VIEW_FORM } from "../constants/routes.js";
import { SwcError } from "../runtime/error.js";
import { inputValueFromEvent } from "../widgets/field-events.js";
import { ACTIVITY_CONTEXT, activityBoardHref, type ActivityContextPayload } from "./activity-context.js";
import { setMessagesUnreadBadge } from "./activity-panel.js";
import { rankByFuzzy } from "../util/fuzzy.js";
import { formatRelativeTime } from "../util/format-relative-time.js";
import { clearImDraft, readImDraft, writeImDraft } from "../util/im-draft.js";
import { userInitials } from "../util/user-initials.js";
import {
  attachmentPreviewModal,
  type AttachmentPreviewTarget,
} from "../views/shared/AttachmentPreview.js";

interface DirectUser {
  id: number;
  name: string;
  login: string;
}

interface DirectConversation {
  userId: number;
  name: string;
  preview?: string;
  lastMessage?: string;
  unreadCount?: number;
}

interface DirectAttachment {
  id: number;
  name: string;
  url: string;
  mimetype?: string;
}

interface DirectMessage {
  id: number;
  body: string;
  createDate: string;
  outgoing: boolean;
  pending?: boolean;
  resModel?: string;
  resId?: number;
  attachments?: DirectAttachment[];
}

interface DirectStatus {
  enabled?: boolean;
  unread?: number;
  companyFormHref?: string;
}

const PICKER_LIMIT = 15;
const DRAFT_DEBOUNCE_MS = 400;

export class InternalMessagesPanel extends SwcComponent {
  private search = "";
  private serverSearchResults: DirectUser[] = [];
  private userDirectory: DirectUser[] = [];
  private conversations: DirectConversation[] = [];
  private peerId = 0;
  private peerName = "";
  private thread: DirectMessage[] = [];
  private draft = "";
  private loading = false;
  private bootstrapLoading = true;
  private posting = false;
  private imDisabled = false;
  private disabledReason = "";
  private companyFormHref = "";
  private banner = "";
  private linkModel = "";
  private linkRecordId = 0;
  private activityCtx: ActivityContextPayload = { model: "", recordId: 0, recordMessagesEligible: false, recordLogEligible: false };
  private pickerIndex = 0;
  private pickerOpen = false;
  private scrollThreadEnd = false;
  private focusComposerNext = false;
  private uploadPct = 0;
  private uploading = false;
  private attachmentPreview: AttachmentPreviewTarget | null = null;
  private draftTimer: ReturnType<typeof setTimeout> | null = null;
  private searchTimer: ReturnType<typeof setTimeout> | null = null;
  private unsubActivity: (() => void) | null = null;

  override onMount(): void {
    this.companyFormHref = this.env.bootstrap.companyFormHref ?? "";
    if (this.env.bootstrap.internalMessagesEnabled === false) {
      this.imDisabled = true;
      this.disabledReason = "Internal messages are disabled for this company.";
    }
    this.unsubActivity = this.env.services.bus.subscribe(ACTIVITY_CONTEXT, (payload) => {
      const ctx = payload as ActivityContextPayload;
      this.activityCtx = ctx;
      if (ctx.recordId > 0 && ctx.model) {
        this.linkModel = ctx.model;
        this.linkRecordId = ctx.recordId;
      } else {
        this.linkModel = "";
        this.linkRecordId = 0;
      }
      this.rerender();
    });
    void this.bootstrapPanel();
  }

  override onWillUnmount(): void {
    this.unsubActivity?.();
    this.unsubActivity = null;
  }

  override afterPatch(): void {
    if (this.scrollThreadEnd) {
      this.scrollThreadEnd = false;
      const el = this.rootElement?.querySelector(".sum-msg-thread");
      if (el) el.scrollTop = el.scrollHeight;
    }
    if (this.focusComposerNext) {
      this.focusComposerNext = false;
      const ta = this.rootElement?.querySelector<HTMLTextAreaElement>(".sum-msg-input:not([disabled])");
      ta?.focus();
    }
  }

  private apiBase(): string {
    return this.env.bootstrap.swcApiBase || SWC_API_BASE;
  }

  private async bootstrapPanel(): Promise<void> {
    this.bootstrapLoading = true;
    this.rerender();
    await this.loadStatus();
    if (!this.imDisabled) {
      await Promise.all([this.refreshConversations(), this.loadDirectoryUsers()]);
    }
    this.bootstrapLoading = false;
    this.rerender();
  }

  private async loadStatus(): Promise<void> {
    try {
      const data = await this.env.services.http.getJSON<DirectStatus>(`${this.apiBase()}/direct/status`);
      if (data.enabled === false) {
        this.imDisabled = true;
        this.disabledReason = "Internal messages are disabled for this company.";
      } else {
        this.imDisabled = false;
      }
      if (data.companyFormHref) this.companyFormHref = data.companyFormHref;
      setMessagesUnreadBadge(data.unread ?? 0);
    } catch {
      /* keep bootstrap defaults */
    }
  }

  private rankedPickerUsers(): DirectUser[] {
    const q = this.search.trim();
    const ranked = rankByFuzzy(q, this.userDirectory, (u) => [u.name, u.login]);
    let list = ranked.slice(0, PICKER_LIMIT).map((r) => r.item);
    if (q.length >= 2 && list.length < 3) {
      const seen = new Set(list.map((u) => u.id));
      for (const u of this.serverSearchResults) {
        if (seen.has(u.id)) continue;
        list.push(u);
        seen.add(u.id);
        if (list.length >= PICKER_LIMIT) break;
      }
    }
    return list;
  }

  private async refreshConversations(): Promise<void> {
    try {
      const data = await this.env.services.http.getJSON<{ conversations: DirectConversation[] }>(
        `${this.apiBase()}/direct/conversations`,
      );
      this.conversations = data.conversations ?? [];
      this.banner = "";
    } catch (err) {
      this.conversations = [];
      if (err instanceof SwcError && err.message.includes("403")) {
        this.banner =
          "Recent chats are unavailable (permissions or IM not fully set up). You can still search colleagues below.";
      } else {
        this.banner = "Recent chats could not be loaded. Search for a colleague to start messaging.";
      }
    }
  }

  private async loadDirectoryUsers(): Promise<void> {
    try {
      const data = await this.env.services.http.getJSON<{ users: DirectUser[] }>(
        `${this.apiBase()}/direct/users`,
      );
      this.userDirectory = data.users ?? [];
    } catch {
      this.userDirectory = [];
    }
  }

  private async fetchServerSearch(q: string): Promise<void> {
    try {
      const data = await this.env.services.http.getJSON<{ users: DirectUser[] }>(
        `${this.apiBase()}/direct/users?q=${encodeURIComponent(q)}`,
      );
      this.serverSearchResults = data.users ?? [];
    } catch {
      this.serverSearchResults = [];
    }
    this.rerender();
  }

  private scheduleSearch(): void {
    if (this.searchTimer) clearTimeout(this.searchTimer);
    this.searchTimer = setTimeout(() => {
      const q = this.search.trim();
      this.pickerIndex = 0;
      this.pickerOpen = q.length > 0;
      if (q.length >= 2 && this.rankedPickerUsers().length < 3) {
        void this.fetchServerSearch(q);
      } else {
        this.serverSearchResults = [];
        this.rerender();
      }
    }, 150);
  }

  private scheduleDraftSave(): void {
    if (this.draftTimer) clearTimeout(this.draftTimer);
    this.draftTimer = setTimeout(() => writeImDraft(this.peerId, this.draft), DRAFT_DEBOUNCE_MS);
  }

  private onSearchKeydown(event: Event): void {
    const ev = event as KeyboardEvent;
    const q = this.search.trim();
    if (!q) return;
    const picks = this.rankedPickerUsers();
    if (picks.length === 0) return;

    if (ev.key === "ArrowDown") {
      ev.preventDefault();
      this.pickerIndex = Math.min(this.pickerIndex + 1, picks.length - 1);
      this.rerender();
    } else if (ev.key === "ArrowUp") {
      ev.preventDefault();
      this.pickerIndex = Math.max(this.pickerIndex - 1, 0);
      this.rerender();
    } else if (ev.key === "Enter") {
      ev.preventDefault();
      const u = picks[this.pickerIndex] ?? picks[0];
      if (u) void this.openPeer(u.id, u.name || u.login);
    } else if (ev.key === "Escape") {
      this.search = "";
      this.pickerOpen = false;
      this.serverSearchResults = [];
      this.rerender();
    }
  }

  private onComposerKeydown(event: Event): void {
    const ev = event as KeyboardEvent;
    if (ev.key !== "Enter") return;
    if (ev.shiftKey) return;
    ev.preventDefault();
    void this.postMessage();
  }

  private async openPeer(userId: number, name: string): Promise<void> {
    if (this.peerId > 0) writeImDraft(this.peerId, this.draft);
    this.peerId = userId;
    this.peerName = name;
    this.search = "";
    this.serverSearchResults = [];
    this.pickerOpen = false;
    this.draft = readImDraft(userId);
    this.focusComposerNext = true;
    await this.loadThread();
  }

  private postFields(body: string): Record<string, string> {
    const fields: Record<string, string> = {
      recipient_id: String(this.peerId),
      body,
    };
    if (this.linkModel && this.linkRecordId > 0) {
      fields.res_model = this.linkModel;
      fields.res_id = String(this.linkRecordId);
    }
    return fields;
  }

  private recordLinkHref(model: string, id: number): string {
    return `/web?model=${encodeURIComponent(model)}&view_type=${VIEW_FORM}&id=${id}`;
  }

  private async loadThread(): Promise<void> {
    if (this.peerId <= 0) return;
    this.loading = true;
    this.rerender();
    try {
      const data = await this.env.services.http.getJSON<{ messages: DirectMessage[] }>(
        `${this.apiBase()}/direct/thread?userId=${this.peerId}`,
      );
      this.thread = data.messages ?? [];
      this.scrollThreadEnd = true;
      await this.loadStatus();
      await this.refreshConversations();
    } finally {
      this.loading = false;
      this.rerender();
    }
  }

  private async postMessage(): Promise<void> {
    const body = this.draft.trim();
    if (!body || this.peerId <= 0 || this.posting) return;
    const tempId = -Date.now();
    const optimistic: DirectMessage = {
      id: tempId,
      body,
      createDate: new Date().toISOString(),
      outgoing: true,
      pending: true,
    };
    this.thread = [...this.thread, optimistic];
    this.draft = "";
    clearImDraft(this.peerId);
    this.posting = true;
    this.scrollThreadEnd = true;
    this.rerender();
    try {
      await this.env.services.http.postForm(`${this.apiBase()}/direct/post`, this.postFields(body));
      await this.loadThread();
      await this.refreshConversations();
    } catch {
      this.thread = this.thread.filter((m) => m.id !== tempId);
      this.draft = body;
      this.rerender();
    } finally {
      this.posting = false;
      this.rerender();
    }
  }

  private async uploadFile(file: File): Promise<void> {
    if (this.peerId <= 0) return;
    let messageId = 0;
    if (this.draft.trim()) {
      const res = await this.env.services.http.postForm(
        `${this.apiBase()}/direct/post`,
        this.postFields(this.draft.trim()),
      );
      const parsed = (await res.json()) as { id?: number };
      messageId = parsed.id ?? 0;
      this.draft = "";
      clearImDraft(this.peerId);
    } else if (this.thread.length > 0) {
      messageId = this.thread[this.thread.length - 1].id;
    } else {
      const res = await this.env.services.http.postForm(`${this.apiBase()}/direct/post`, this.postFields(" "));
      const parsed = (await res.json()) as { id?: number };
      messageId = parsed.id ?? 0;
    }
    if (messageId <= 0) return;
    const form = new FormData();
    form.set("message_id", String(messageId));
    form.set("file", file);
    this.uploading = true;
    this.uploadPct = 0;
    this.rerender();
    try {
      await this.env.services.http.postMultipart(`${this.apiBase()}/direct/upload`, form, (pct) => {
        this.uploadPct = pct;
        this.rerender();
      });
    } finally {
      this.uploading = false;
      this.uploadPct = 0;
    }
    await this.loadThread();
    await this.refreshConversations();
    this.scrollThreadEnd = true;
    this.rerender();
  }

  private colleagueDirectory(): DirectUser[] {
    const convIds = new Set(this.conversations.map((c) => c.userId));
    return this.userDirectory.filter((u) => !convIds.has(u.id));
  }

  private userRow(
    u: DirectUser,
    active: boolean,
    sub: string,
    unread = 0,
    pickerIdx?: number,
  ): ReturnType<typeof html> {
    const label = u.name || u.login;
    const initials = userInitials(u.name, u.login);
    return html`<li>
      <button
        type="button"
        class="sum-msg-peer-btn${active ? " is-active" : ""}${pickerIdx === this.pickerIndex ? " is-highlight" : ""}"
        id=${pickerIdx !== undefined ? `sum-msg-pick-${pickerIdx}` : undefined}
        @click=${() => void this.openPeer(u.id, label)}
      >
        <span class="sum-msg-peer-row">
          <span class="sum-msg-avatar" aria-hidden="true">${initials}</span>
          <span class="sum-msg-peer-text">
            <span class="sum-msg-author">${label}</span>
            <span class="sum-msg-time">${sub}</span>
          </span>
          ${unread > 0 ? html`<span class="sum-msg-unread-dot" aria-label="${unread} unread">${unread > 9 ? "9+" : unread}</span>` : ""}
        </span>
      </button>
    </li>`;
  }

  private composerHint(): ReturnType<typeof html> {
    if (this.activityCtx.recordId > 0 && this.activityCtx.model) {
      return html`<p class="sum-msg-form-hint">
        Linked to open record · Enter to send ·
        <a class="sum-chatter-attachment-link" href=${activityBoardHref(this.activityCtx)}>Activities for this record</a>
      </p>`;
    }
    return html`<p class="sum-msg-form-hint">Enter to send · Shift+Enter for new line</p>`;
  }

  private composerStub(): ReturnType<typeof html> {
    return html`<div class="sum-msg-composer sum-msg-composer--stub">
      <div class="sum-msg-composer-bar">
        <span class="sum-msg-composer-attach sum-btn sum-btn--secondary" aria-disabled="true">Attach</span>
        <textarea
          class="sum-msg-input"
          rows="1"
          placeholder="Pick a colleague to chat…"
          disabled="disabled"
        ></textarea>
        <button type="button" class="sum-msg-send" disabled="disabled">Send</button>
      </div>
      ${this.composerHint()}
    </div>`;
  }

  private activeComposer(): ReturnType<typeof html> {
    return html`<div class="sum-msg-composer">
      <div class="sum-msg-composer-bar">
        <label class="sum-msg-composer-attach sum-chatter-upload">
          <input
            type="file"
            hidden
            @change=${(event: Event) => {
              const input = event.target as HTMLInputElement;
              const file = input.files?.[0];
              input.value = "";
              if (file) void this.uploadFile(file);
            }}
          />
          <span class="sum-btn sum-btn--secondary">${this.uploading ? `Uploading ${this.uploadPct}%` : "Attach"}</span>
        </label>
        <textarea
          class="sum-msg-input"
          rows="1"
          placeholder="Write a message…"
          aria-label=${`Message ${this.peerName}`}
          value=${this.draft}
          @keydown=${(event: Event) => this.onComposerKeydown(event)}
          @input=${(event: Event) => {
            this.draft = inputValueFromEvent(event);
            this.scheduleDraftSave();
            this.rerender();
          }}
        ></textarea>
        <button
          type="button"
          class="sum-msg-send"
          disabled=${this.posting ? "disabled" : undefined}
          @click=${() => void this.postMessage()}
        >
          ${this.posting ? "…" : "Send"}
        </button>
      </div>
      ${this.composerHint()}
    </div>`;
  }

  private messageBubble(m: DirectMessage): ReturnType<typeof html> {
    const when = formatRelativeTime(m.createDate) || m.createDate;
    return html`<li
      class="sum-msg-bubble${m.outgoing ? " sum-msg-bubble--out" : ""}${m.pending ? " sum-msg-bubble--pending" : ""}"
    >
      <p class="sum-msg-bubble-meta">${when}${m.pending ? " · Sending" : ""}</p>
      <div class="sum-msg-bubble-inner">
        <div class="sum-msg-body">${m.body}</div>
      </div>
      ${m.resModel && (m.resId ?? 0) > 0
        ? html`<p class="sum-msg-bubble-extra">
            <a class="sum-chatter-attachment-link" href=${this.recordLinkHref(m.resModel, m.resId ?? 0)}>Related record</a>
          </p>`
        : ""}
      ${(m.attachments ?? []).length > 0
        ? html`<ul class="sum-chatter-attachments--inline sum-msg-bubble-extra">
            ${(m.attachments ?? []).map(
              (a) => html`<li>
                <button
                  type="button"
                  class="sum-chatter-attachment-link"
                  @click=${() => {
                    this.attachmentPreview = {
                      name: a.name,
                      url: a.url,
                      mimetype: a.mimetype,
                    };
                    this.rerender();
                  }}
                >
                  ${a.name}
                </button>
              </li>`,
            )}
          </ul>`
        : ""}
    </li>`;
  }

  private activeChat(): ReturnType<typeof html> {
    const initials = userInitials(this.peerName, "");
    return html`<div class="sum-msg-chat">
      <header class="sum-msg-chat-head">
        <span class="sum-msg-avatar" aria-hidden="true">${initials}</span>
        <h3 class="sum-msg-chat-title">${this.peerName}</h3>
      </header>
      <div class="sum-msg-thread" role="log" aria-live="polite" aria-relevant="additions">
        ${this.loading
          ? html`<div class="sum-msg-skeleton"><div class="sum-msg-skeleton-line"></div></div>`
          : this.thread.length === 0
            ? html`<div class="sum-msg-thread-center">
                <p class="sum-msg-thread-empty-hint">No messages yet. Say hello.</p>
              </div>`
            : html`<ul class="sum-msg-bubbles">${this.thread.map((m) => this.messageBubble(m))}</ul>`}
      </div>
      ${this.activeComposer()}
      ${attachmentPreviewModal(this.attachmentPreview, () => {
        this.attachmentPreview = null;
        this.rerender();
      })}
    </div>`;
  }

  override template() {
    if (this.imDisabled) {
      return html`<div class="sum-msg-empty" role="status">
        <p class="sum-msg-empty-title">Messages</p>
        <p class="sum-msg-empty-hint">${this.disabledReason}</p>
        ${this.companyFormHref
          ? html`<p class="sum-msg-empty-hint">
              <a class="sum-chatter-attachment-link" href=${this.companyFormHref}>Open company settings</a>
              and enable <strong>Internal chat</strong>.
            </p>`
          : ""}
      </div>`;
    }

    if (this.bootstrapLoading) {
      return html`<div class="sum-msg-shell sum-msg-shell--dock" aria-busy="true">
        <div class="sum-msg-skeleton">
          <div class="sum-msg-skeleton-line"></div>
          <div class="sum-msg-skeleton-line"></div>
          <div class="sum-msg-skeleton-line sum-msg-skeleton-line--short"></div>
        </div>
      </div>`;
    }

    const q = this.search.trim();
    const pickerUsers = q ? this.rankedPickerUsers() : [];
    const showPicker = this.pickerOpen && q.length > 0;

    const sidebarBlocks: ReturnType<typeof html>[] = [];

    if (showPicker) {
      sidebarBlocks.push(
        html`<ul class="sum-msg-peer-list sum-msg-picker" role="listbox" aria-label="Search results">
          ${pickerUsers.length === 0
            ? html`<li class="sum-msg-picker-empty">No matches for “${q}”</li>`
            : pickerUsers.map((u, i) =>
                this.userRow(u, this.peerId === u.id, u.login, 0, i),
              )}
        </ul>`,
      );
    } else {
      if (this.conversations.length > 0) {
        sidebarBlocks.push(html`<p class="sum-msg-section-label">Recent</p>`);
        sidebarBlocks.push(
          html`<ul class="sum-msg-peer-list">
            ${this.conversations.map((c) => {
              const when = formatRelativeTime(c.lastMessage ?? "");
              const sub = [c.preview, when].filter(Boolean).join(" · ");
              return this.userRow(
                { id: c.userId, name: c.name, login: "" },
                this.peerId === c.userId,
                sub,
                c.unreadCount ?? 0,
              );
            })}
          </ul>`,
        );
      }
      const colleagues = this.colleagueDirectory();
      if (colleagues.length > 0) {
        sidebarBlocks.push(html`<p class="sum-msg-section-label">All colleagues</p>`);
        sidebarBlocks.push(
          html`<ul class="sum-msg-peer-list">
            ${colleagues.map((u) => this.userRow(u, this.peerId === u.id, u.login))}
          </ul>`,
        );
      }
      if (this.conversations.length === 0 && colleagues.length === 0) {
        sidebarBlocks.push(html`<p class="sum-msg-picker-empty">No internal users found.</p>`);
      }
    }

    return html`<div class="sum-msg-shell sum-msg-shell--dock">
      ${this.banner ? html`<p class="sum-msg-banner" role="status">${this.banner}</p>` : ""}
      <div class="sum-msg-layout">
        <div class="sum-msg-sidebar">
          <input
            type="search"
            class="sum-msg-search"
            role="combobox"
            aria-expanded=${showPicker ? "true" : "false"}
            aria-controls="sum-msg-picker-list"
            placeholder="Search colleagues…"
            autocomplete="off"
            value=${this.search}
            @input=${(event: Event) => {
              this.search = inputValueFromEvent(event);
              this.pickerOpen = this.search.trim().length > 0;
              this.scheduleSearch();
              this.rerender();
            }}
            @keydown=${(event: Event) => this.onSearchKeydown(event)}
          />
          <div class="sum-msg-sidebar-scroll" id="sum-msg-picker-list">${sidebarBlocks}</div>
        </div>
        <div class="sum-msg-main">
          ${this.peerId <= 0
            ? html`<div class="sum-msg-welcome">
                <p class="sum-msg-thread-empty-title">Your messages</p>
                <p class="sum-msg-thread-empty-hint">Choose someone from the list or search to start a conversation.</p>
              </div>
              ${this.composerStub()}`
            : this.activeChat()}
        </div>
      </div>
    </div>`;
  }
}
