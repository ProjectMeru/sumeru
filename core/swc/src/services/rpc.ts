import { SwcError } from "../runtime/error.js";

interface RpcEnvelope<T = unknown> {
  ok?: boolean;
  result?: T;
  error?: { code?: string; message: string; details?: unknown };
}

function isRpcEnvelope(data: unknown): data is RpcEnvelope {
  return typeof data === "object" && data !== null;
}

export class RpcService {
  private readonly url: string;
  private readonly csrfToken: string;
  private readonly searchReadCache = new Map<string, Promise<Record<string, unknown>[]>>();

  constructor(url: string, csrfToken: string) {
    this.url = url;
    this.csrfToken = csrfToken;
  }

  private searchReadKey(
    model: string,
    domain: unknown[],
    fields: string[],
    limit: number,
  ): string {
    return JSON.stringify({ model, domain, fields, limit });
  }

  /** Clears cached search_read results (e.g. after writes). */
  invalidateSearchReadCache(): void {
    this.searchReadCache.clear();
  }

  private async dispatch<T>(model: string, method: string, args: unknown[] = [], kwargs: Record<string, unknown> = {}): Promise<T> {
    const body: Record<string, unknown> = { model, method, args };
    if (Object.keys(kwargs).length > 0) {
      body.kwargs = kwargs;
    }
    const res = await fetch(this.url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": this.csrfToken,
      },
      credentials: "same-origin",
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      throw new SwcError(`RPC HTTP ${res.status}`, "rpc_http");
    }
    const raw: unknown = await res.json();
    if (!isRpcEnvelope(raw)) {
      throw new SwcError("RPC response is not an object", "rpc_error");
    }
    const data = raw as RpcEnvelope<T>;
    if (data.ok === false || data.error) {
      throw new SwcError(data.error?.message ?? "RPC failed", "rpc_error", data.error);
    }
    return data.result as T;
  }

  searchRead(model: string, domain: unknown[] = [], fields: string[] = [], limit = 80): Promise<Record<string, unknown>[]> {
    const key = this.searchReadKey(model, domain, fields, limit);
    let pending = this.searchReadCache.get(key);
    if (!pending) {
      pending = this.dispatch<Record<string, unknown>[]>(model, "search_read", [domain, fields], {
        limit,
      });
      this.searchReadCache.set(key, pending);
      void pending.catch(() => {
        this.searchReadCache.delete(key);
      });
    }
    return pending;
  }

  read(model: string, ids: number[], fields: string[] = []): Promise<Record<string, unknown>[]> {
    return this.dispatch(model, "read", [ids, fields]);
  }

  write(model: string, ids: number[], values: Record<string, unknown>): Promise<boolean> {
    this.invalidateSearchReadCache();
    return this.dispatch(model, "write", [ids, values]);
  }

  create(model: string, values: Record<string, unknown>): Promise<number> {
    this.invalidateSearchReadCache();
    return this.dispatch(model, "create", [values]);
  }

  unlink(model: string, ids: number[]): Promise<boolean> {
    this.invalidateSearchReadCache();
    return this.dispatch(model, "unlink", [ids]);
  }

  callMethod(
    model: string,
    method: string,
    recordId: number,
    vals?: Record<string, string>,
  ): Promise<unknown> {
    const args: unknown[] = vals ? [recordId, method, vals] : [recordId, method];
    return this.dispatch(model, "call", args);
  }

  readGroup(
    model: string,
    domain: unknown[],
    fields: string[],
    groupBy: string[],
    limit = 80,
  ): Promise<Record<string, unknown>[]> {
    const spec = {
      domain,
      groupby: groupBy,
      fields: fields.map((name) => ({ name, field: name, measure: name === "id" ? "count" : "sum" })),
    };
    return this.dispatch(model, "read_group", [spec], { limit });
  }

  onchange(
    model: string,
    values: Record<string, unknown>,
    field: string,
  ): Promise<Record<string, unknown>> {
    return this.dispatch(model, "onchange", [values, field]);
  }
}
