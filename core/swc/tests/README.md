# SWC tests

All Vitest cases live under `core/swc/tests/` only. Do not add `*.test.ts` beside source under `src/`.

## Layout

```
tests/
  harness/          Shared helpers (DOM, fetch stubs, view fixtures)
  views/            View shell and shared view utilities
  widgets/          Field widgets and registry
  runtime/          Component runtime, hooks, scheduler
  services/         RPC, HTTP, dialog, notification
  template/         HTML template engine and sum compiler
  model/            Record store and modifiers
  ...
```

## COVERED_GLOB policy

Coverage thresholds (≥90% lines/statements/functions, ≥70% branches) apply only to files matched by `COVERED_GLOB` in `vitest.config.ts`:

- Core layers: `template`, `runtime`, `services`, `model`, `widgets`, `login`, `util`, `constants`, `i18n`
- View layer: `src/views/**/*.ts` (collection/form shells and shared view helpers)

Files listed in `coverage.exclude` are omitted from the threshold (integration-heavy shells such as `FormView`, `KanbanView`, and `form-interactions`; devtools; selected field widgets). When adding a new view or service module, either add tests under `tests/` or add a deliberate exclude with a short comment in `vitest.config.ts`.

## Running tests

```bash
cd core/swc
npm test                 # watch mode
npm run test:run         # single run
npm run test:coverage    # CI gate (90% on COVERED_GLOB)
```

## Writing new tests

1. Add `*.test.ts` under the matching `tests/` subtree — never under `src/`
2. Reuse `tests/harness/dom.ts` for DOM helpers and fetch stubs
3. Reuse `tests/harness/view.ts` for `SwcEnv` and workspace payload fixtures
4. Mock external assets (Leaflet, Chart.js) at the module boundary when testing view shells
