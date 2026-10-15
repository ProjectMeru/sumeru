import { SWC_API_BASE } from "../constants/routes.js";

interface ImportBootstrap {
  batchId: number;
  targetModel: string;
  importMode: string;
  headers: string[];
  mapping: Record<string, string>;
  fields: string[];
  fieldLabels: Record<string, string>;
  nextUrl?: string;
}

interface ImportMessage {
  type?: string;
  message?: string;
  row?: number;
}

interface DryRunResponse {
  preview?: unknown;
  messages?: ImportMessage[];
}

export function mountImportWizard(host: HTMLElement | null, batchId: string, csrf: string): void {
  if (!host || !batchId) return;
  void load(host, batchId, csrf);
}

async function load(host: HTMLElement, batchId: string, csrf: string): Promise<void> {
  const res = await fetch(`${SWC_API_BASE}/import?batch=${encodeURIComponent(batchId)}`, { credentials: "same-origin" });
  if (!res.ok) {
    host.textContent = "Could not load import batch.";
    return;
  }
  const data = (await res.json()) as ImportBootstrap;
  render(host, data, csrf);
}

function render(host: HTMLElement, data: ImportBootstrap, csrf: string): void {
  host.innerHTML = "";
  const mapping = { ...data.mapping };
  let dryRunPassed = false;

  const form = document.createElement("div");
  form.className = "sum-import-wizard";

  const table = document.createElement("table");
  table.className = "sum-import-map-table";
  const thead = document.createElement("thead");
  thead.innerHTML = "<tr><th>File column</th><th>Field</th></tr>";
  table.appendChild(thead);
  const tbody = document.createElement("tbody");
  for (const col of data.headers) {
    const tr = document.createElement("tr");
    const tdCol = document.createElement("td");
    tdCol.textContent = col;
    const tdField = document.createElement("td");
    const select = document.createElement("select");
    select.dataset.column = col;
    appendOption(select, "-", "Ignore");
    for (const f of data.fields) {
      appendOption(select, f, data.fieldLabels[f] || f);
    }
    select.value = mapping[col] ?? "-";
    select.addEventListener("change", () => {
      mapping[col] = select.value;
      dryRunPassed = false;
      updateImportButton();
    });
    tdField.appendChild(select);
    tr.append(tdCol, tdField);
    tbody.appendChild(tr);
  }
  table.appendChild(tbody);
  form.appendChild(table);

  const messageBox = document.createElement("div");
  messageBox.className = "sum-import-preview";
  form.appendChild(messageBox);

  const details = document.createElement("pre");
  details.className = "sum-import-preview-details";
  details.hidden = true;
  form.appendChild(details);

  const actions = document.createElement("div");
  actions.className = "sum-import-actions";

  const importBtn = button("Import", () => {
    if (!dryRunPassed) {
      window.alert("Run a dry run with no errors before importing.");
      return;
    }
    void runImport(data.batchId, mapping, csrf, data.nextUrl ?? "/web/home");
  });
  importBtn.className = "sum-btn sum-btn--primary";

  const updateImportButton = (): void => {
    importBtn.disabled = !dryRunPassed;
    importBtn.title = dryRunPassed ? "" : "Run dry run with zero errors first";
  };
  updateImportButton();

  actions.append(
    button("Preview", () => void runPreview(messageBox, details, data.batchId, mapping, csrf)),
    button("Dry run", () =>
      void runDryRun(messageBox, details, data.batchId, mapping, csrf, (ok) => {
        dryRunPassed = ok;
        updateImportButton();
      }),
    ),
    button("Save template", () => void runSaveTemplate(messageBox, data.batchId, mapping, csrf)),
    importBtn,
  );
  form.appendChild(actions);

  host.appendChild(form);
}

function appendOption(select: HTMLSelectElement, value: string, label: string): void {
  const opt = document.createElement("option");
  opt.value = value;
  opt.textContent = label;
  select.appendChild(opt);
}

function button(label: string, onClick: () => void): HTMLButtonElement {
  const b = document.createElement("button");
  b.type = "button";
  b.className = "sum-btn sum-btn--secondary";
  b.textContent = label;
  b.addEventListener("click", onClick);
  return b;
}

function renderMessages(box: HTMLElement, messages: ImportMessage[]): void {
  box.innerHTML = "";
  const list = document.createElement("ul");
  list.className = "sum-import-message-list";
  for (const m of messages) {
    const li = document.createElement("li");
    li.className = `sum-import-message sum-import-message--${m.type ?? "info"}`;
    const row = m.row && m.row > 0 ? `Row ${m.row}: ` : "";
    li.textContent = `${row}${m.message ?? ""}`;
    list.appendChild(li);
  }
  box.appendChild(list);
}

function countErrors(messages: ImportMessage[]): number {
  return messages.filter((m) => (m.type ?? "").toLowerCase() === "error").length;
}

async function runPreview(
  box: HTMLElement,
  details: HTMLElement,
  batchId: number,
  mapping: Record<string, string>,
  csrf: string,
): Promise<void> {
  box.textContent = "Previewing…";
  details.hidden = true;
  try {
    const res = await postJSON("/web/bulk/preview", { batch_id: batchId, column_mapping: mapping }, csrf);
    details.textContent = JSON.stringify(res, null, 2);
    details.hidden = false;
    box.textContent = "Preview loaded (see details below).";
  } catch (err) {
    box.textContent = String(err);
  }
}

async function runSaveTemplate(
  box: HTMLElement,
  batchId: number,
  mapping: Record<string, string>,
  csrf: string,
): Promise<void> {
  const name = window.prompt("Template name");
  if (!name?.trim()) {
    return;
  }
  box.textContent = "Saving template…";
  const res = await postJSON("/web/bulk/save-template", {
    name: name.trim(),
    batch_id: batchId,
    column_mapping: mapping,
    shared: false,
  }, csrf);
  box.textContent = `Saved template #${String((res as { template_id?: number }).template_id ?? "")}`;
}

async function runDryRun(
  box: HTMLElement,
  details: HTMLElement,
  batchId: number,
  mapping: Record<string, string>,
  csrf: string,
  onResult: (passed: boolean) => void,
): Promise<void> {
  box.textContent = "Dry run…";
  details.hidden = true;
  try {
    const res = (await postJSON("/web/bulk/dry-run", { batch_id: batchId, column_mapping: mapping }, csrf)) as DryRunResponse;
    const messages = res.messages ?? [];
    renderMessages(box, messages);
    details.textContent = JSON.stringify(res.preview ?? {}, null, 2);
    details.hidden = false;
    const errors = countErrors(messages);
    onResult(errors === 0);
    if (errors > 0) {
      const note = document.createElement("p");
      note.className = "sum-import-dry-run-fail";
      note.textContent = `${errors} error(s) — fix mapping or data before importing.`;
      box.appendChild(note);
    }
  } catch (err) {
    box.textContent = String(err);
    onResult(false);
  }
}

async function runImport(batchId: number, mapping: Record<string, string>, csrf: string, nextUrl: string): Promise<void> {
  const body = new URLSearchParams({
    id: String(batchId),
    column_mapping: JSON.stringify(mapping),
    csrf_token: csrf,
  });
  const res = await fetch("/web/bulk/confirm", { method: "POST", credentials: "same-origin", body });
  if (res.redirected) {
    window.location.href = res.url;
    return;
  }
  if (res.ok) {
    window.location.href = nextUrl || "/web/home";
  }
}

async function postJSON(url: string, payload: Record<string, unknown>, csrf: string): Promise<unknown> {
  const res = await fetch(url, {
    method: "POST",
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      "X-CSRF-Token": csrf,
    },
    body: JSON.stringify({ ...payload, csrf_token: csrf }),
  });
  const text = await res.text();
  if (!res.ok) {
    let message = text;
    try {
      const errBody = JSON.parse(text) as { error?: { message?: string } };
      if (errBody.error?.message) {
        message = errBody.error.message;
      }
    } catch {
      // plain-text error body
    }
    throw new Error(message || `HTTP ${res.status}`);
  }
  return JSON.parse(text) as unknown;
}
