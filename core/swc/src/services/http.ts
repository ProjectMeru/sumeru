import { SwcError } from "../runtime/error.js";

export class HttpService {
  private readonly csrfToken: string;

  constructor(csrfToken: string) {
    this.csrfToken = csrfToken;
  }

  get csrf(): string {
    return this.csrfToken;
  }

  async getJSON<T>(url: string): Promise<T> {
    const res = await fetch(url, {
      credentials: "same-origin",
      headers: { Accept: "application/json" },
    });
    if (!res.ok) {
      throw new SwcError(`GET ${url} failed: ${res.status}`, "http_get");
    }
    return (await res.json()) as T;
  }

  async postForm(url: string, data: Record<string, string>): Promise<Response> {
    const body = new URLSearchParams({ ...data, csrf_token: this.csrfToken });
    return fetch(url, {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body,
    });
  }

  async postJSON<T>(url: string, body: Record<string, unknown>): Promise<T> {
    const payload = { ...body, csrf_token: this.csrfToken };
    const res = await fetch(url, {
      method: "POST",
      credentials: "same-origin",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
        "X-CSRF-Token": this.csrfToken,
      },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      throw new SwcError(`POST ${url} failed: ${res.status}`, "http_post");
    }
    return (await res.json()) as T;
  }

  async delete(url: string): Promise<void> {
    const res = await fetch(url, {
      method: "DELETE",
      credentials: "same-origin",
      headers: {
        Accept: "application/json",
        "X-CSRF-Token": this.csrfToken,
      },
    });
    if (!res.ok) {
      throw new SwcError(`DELETE ${url} failed: ${res.status}`, "http_delete");
    }
  }
}
