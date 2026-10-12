const LOGIN_PATH = "/web/login";

let redirectScheduled = false;

/** Resets redirect guard (Vitest only). */
export function resetSessionRedirectForTests(): void {
  redirectScheduled = false;
}

/** Full navigation to login with return path; skips when already on auth routes. */
export function redirectToLoginIfNeeded(): void {
  if (redirectScheduled || typeof window === "undefined") {
    return;
  }
  const path = window.location.pathname;
  if (path === LOGIN_PATH || path.startsWith("/web/auth/")) {
    return;
  }
  redirectScheduled = true;
  const returnPath = window.location.pathname + window.location.search;
  const params = new URLSearchParams();
  if (returnPath && returnPath !== LOGIN_PATH) {
    params.set("next", returnPath);
  }
  const query = params.toString();
  window.location.assign(LOGIN_PATH + (query ? `?${query}` : ""));
}

export function shouldRedirectForHttpStatus(status: number): boolean {
  return status === 401;
}

export function shouldRedirectForRpcError(code: string | undefined): boolean {
  return code === "UNAUTHORIZED";
}
