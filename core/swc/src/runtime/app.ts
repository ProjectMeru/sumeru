import { SwcComponent } from "./component.js";
import { SwcEnv } from "./env.js";
import { setActiveComponent, runMountCallbacks, runUnmountCallbacks } from "./hooks.js";
import { html } from "../template/html.js";
import { SwcError } from "./error.js";

class ErrorBoundary extends SwcComponent<{ error: SwcError; retry: () => void }> {
  template() {
    const { error, retry } = this.props;
    return html`
      <div class="sum-flash sum-flash--error">
        <strong>SWC error</strong>
        <p>${error.message}</p>
        <button type="button" class="sum-btn sum-btn--secondary" @click=${() => retry()}>Retry</button>
      </div>
    `;
  }
}

function renderErrorFallback(error: SwcError, retry: () => void): HTMLElement {
  const wrap = document.createElement("div");
  wrap.className = "sum-flash sum-flash--error";

  const title = document.createElement("strong");
  title.textContent = "SWC error";
  wrap.appendChild(title);

  const message = document.createElement("p");
  message.textContent = error.message;
  wrap.appendChild(message);

  const button = document.createElement("button");
  button.type = "button";
  button.className = "sum-btn sum-btn--secondary";
  button.textContent = "Retry";
  button.addEventListener("click", retry);
  wrap.appendChild(button);

  return wrap;
}

function showError(rootEl: HTMLElement, env: SwcEnv, error: SwcError, retry: () => void): void {
  try {
    const boundary = new ErrorBoundary({ error, retry }, env);
    rootEl.replaceChildren(boundary.render());
  } catch {
    rootEl.replaceChildren(renderErrorFallback(error, retry));
  }
}

export class SwcApp {
  private readonly env: SwcEnv;
  private readonly Root: new (props: Record<string, unknown>, env: SwcEnv) => SwcComponent;
  private rootEl: HTMLElement | null = null;
  private component: SwcComponent | null = null;
  private scheduled = false;

  constructor(env: SwcEnv, Root: new (props: Record<string, unknown>, env: SwcEnv) => SwcComponent) {
    this.env = env;
    this.Root = Root;
  }

  static start(
    mountEl: HTMLElement,
    env: SwcEnv,
    Root: new (props: Record<string, unknown>, env: SwcEnv) => SwcComponent,
  ): SwcApp {
    const app = new SwcApp(env, Root);
    app.mount(mountEl);
    return app;
  }

  mount(el: HTMLElement): void {
    this.rootEl = el;
    this.renderRoot();
  }

  private schedulePatch(): void {
    if (this.scheduled) return;
    this.scheduled = true;
    requestAnimationFrame(() => {
      this.scheduled = false;
      this.renderRoot();
    });
  }

  private renderRoot(): void {
    if (!this.rootEl) return;
    try {
      if (!this.component) {
        this.component = new this.Root({}, this.env);
        this.component.setup?.();
        setActiveComponent({ schedulePatch: () => this.schedulePatch() });
        runMountCallbacks();
        this.rootEl.replaceChildren(this.component.render());
      } else {
        setActiveComponent({ schedulePatch: () => this.schedulePatch() });
        this.component.patch();
      }
    } catch (err) {
      const swcErr = err instanceof SwcError ? err : new SwcError(String(err));
      showError(this.rootEl, this.env, swcErr, () => this.retry());
    }
  }

  private retry(): void {
    runUnmountCallbacks();
    this.component?.destroy();
    this.component = null;
    this.renderRoot();
  }

  destroy(): void {
    runUnmountCallbacks();
    this.component?.destroy();
    this.component = null;
    this.rootEl = null;
  }
}
