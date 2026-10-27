import { SwcError } from "../runtime/error.js";
import {
  redirectToLoginIfNeeded,
  shouldRedirectForHttpStatus,
  shouldRedirectForRpcError,
} from "./session-redirect.js";

interface RpcEnvelope<T = unknown> {
  success?: boolean;
  data?: T;
  error?: { code?: string; message?: string; context?: { code?: string } };
}

function isRpcEnvelope(data: unknown): data is RpcEnvelope {
  return typeof data === "object" && data !== null;
}

function rpcErrorCode(error: RpcEnvelope["error"]): string | undefined {
  if (!error) return undefined;
  if (error.code) return error.code;
  const ctx = error.context;
  if (ctx && typeof ctx.code === "string") return ctx.code;
  return undefined;
}

export class RpcService {
  private readonly rpcBase: string;
  private readonly csrfToken: string;
  private readonly apiDatabase: string;
  private readonly apiTenant: string;
  private readonly searchReadCache = new Map<string, Promise<Record<string, unknown>[]>>();

  constructor(rpcBase: string, csrfToken: string, apiDatabase: string, apiTenant: string) {
    this.rpcBase = rpcBase.replace(/\/$/, "");
    this.csrfToken = csrfToken;
    this.apiDatabase = apiDatabase;
    this.apiTenant = apiTenant;
  }

  private rpcURL(model: string, method: string): string {
    return `${this.rpcBase}/${model}/${method}`;
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

  private async dispatch<T>(model: string, method: string, body: Record<string, unknown>): Promise<T> {
    const res = await fetch(this.rpcURL(model, method), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": this.csrfToken,
        "X-Sumeru-Database": this.apiDatabase,
        "X-Sumeru-Tenant": this.apiTenant,
      },
      credentials: "same-origin",
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      if (shouldRedirectForHttpStatus(res.status)) {
        redirectToLoginIfNeeded();
      }
      throw new SwcError(`RPC HTTP ${res.status}`, "rpc_http");
    }
    const raw: unknown = await res.json();
    if (!isRpcEnvelope(raw)) {
      throw new SwcError("RPC response is not an object", "rpc_error");
    }
    const data = raw as RpcEnvelope<T>;
    if (data.success === false || data.error) {
      const code = rpcErrorCode(data.error);
      if (shouldRedirectForRpcError(code)) {
        redirectToLoginIfNeeded();
      }
      throw new SwcError(data.error?.message ?? "RPC failed", "rpc_error", data.error);
    }
    return data.data as T;
  }

  searchRead(model: string, domain: unknown[] = [], fields: string[] = [], limit = 80): Promise<Record<string, unknown>[]> {
    const key = this.searchReadKey(model, domain, fields, limit);
    let pending = this.searchReadCache.get(key);
    if (!pending) {
      pending = this.dispatch<Record<string, unknown>[]>(model, "search_read", {
        domain,
        fields,
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
    const body: Record<string, unknown> = { ids };
    if (fields.length > 0) {
      body.fields = fields;
    }
    return this.dispatch(model, "read", body);
  }

  write(model: string, ids: number[], values: Record<string, unknown>): Promise<boolean> {
    this.invalidateSearchReadCache();
    return this.dispatch(model, "write", { ids, values });
  }

  create(model: string, values: Record<string, unknown>): Promise<number> {
    this.invalidateSearchReadCache();
    return this.dispatch(model, "create", { values });
  }

  unlink(model: string, ids: number[]): Promise<boolean> {
    this.invalidateSearchReadCache();
    return this.dispatch(model, "unlink", { ids });
  }

  callMethod(
    model: string,
    method: string,
    recordId: number,
    vals?: Record<string, string>,
  ): Promise<unknown> {
    const body: Record<string, unknown> = { id: recordId, method };
    if (vals) {
      body.vals = vals;
    }
    return this.dispatch(model, "call", body);
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
    return this.dispatch(model, "read_group", { spec, limit });
  }

  onchange(
    model: string,
    values: Record<string, unknown>,
    field: string,
  ): Promise<Record<string, unknown>> {
    return this.dispatch(model, "onchange", { values, field });
  }
}
