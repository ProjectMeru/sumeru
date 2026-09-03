/** Remove ?msg= from the Apps page URL after flash/toast bootstrap (avoids repeat on refresh). */
export function stripAppsFlashQueryParam(locationLike: Pick<Location, "href"> = window.location): void {
  if (!window.history?.replaceState) {
    return;
  }
  const url = new URL(locationLike.href);
  if (!url.searchParams.has("msg")) {
    return;
  }
  url.searchParams.delete("msg");
  const query = url.searchParams.toString();
  const next = url.pathname + (query ? `?${query}` : "") + url.hash;
  window.history.replaceState({}, "", next);
}

function shouldIgnoreListRowClick(target: EventTarget | null): boolean {
  if (!(target instanceof Element)) {
    return false;
  }
  return !!target.closest("button, a, form, input, select, textarea, .sum-apps-list-actions");
}

/** Navigate to module detail when a list row is clicked (except action controls). */
export function handleAppsListRowClick(event: MouseEvent): void {
  if (shouldIgnoreListRowClick(event.target)) {
    return;
  }
  const row = event.currentTarget;
  if (!(row instanceof HTMLTableRowElement)) {
    return;
  }
  const detailURL = row.dataset.detailUrl;
  if (detailURL) {
    window.location.href = detailURL;
  }
}

function initAppsListRowClick(root: Document | ParentNode): void {
  root.querySelectorAll<HTMLTableRowElement>("tr.sum-apps-list-row[data-detail-url]").forEach((row) => {
    row.addEventListener("click", handleAppsListRowClick);
  });
}

function initAppsControlBarAutoSubmit(root: Document | ParentNode): void {
  const form = root.querySelector<HTMLFormElement>("form.sum-apps-control-bar");
  if (!form) {
    return;
  }
  form.querySelectorAll<HTMLSelectElement>("select[data-auto-submit]").forEach((select) => {
    select.addEventListener("change", () => {
      form.requestSubmit();
    });
  });
}

/** Wire Apps page client behaviors (flash strip, list row navigation, control bar auto-submit). */
export function initAppsPage(root: Document | ParentNode = document): void {
  stripAppsFlashQueryParam();
  initAppsListRowClick(root);
  initAppsControlBarAutoSubmit(root);
}
