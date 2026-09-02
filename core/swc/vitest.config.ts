import { defineConfig } from "vitest/config";

/** Source files we enforce ≥90% coverage on (unit-testable core; excludes full view/page shells). */
const COVERED_GLOB = [
  "src/template/**/*.ts",
  "src/runtime/**/*.ts",
  "src/services/**/*.ts",
  "src/model/**/*.ts",
  "src/widgets/**/*.ts",
  "src/views/shared/**/*.ts",
  "src/views/list/control-panel.ts",
  "src/login/**/*.ts",
  "src/util/**/*.ts",
  "src/constants/**/*.ts",
  "src/i18n/**/*.ts",
];

export default defineConfig({
  test: {
    environment: "jsdom",
    include: ["tests/**/*.test.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "text-summary", "lcov"],
      include: COVERED_GLOB,
      exclude: [
        "src/main.ts",
        "src/runtime/app.ts",
        "src/runtime/env.ts",
        "src/runtime/component-host.ts",
        "src/runtime/lifecycle.ts",
        "src/runtime/registry.ts",
        "src/login/password-match-entry.ts",
        "src/devtools/panel.ts",
        "src/devtools/profiler.ts",
        "src/devtools/debug.ts",
        "src/widgets/One2ManyField.ts",
        "src/widgets/Many2OneField.ts",
        "src/widgets/SelectionField.ts",
        "src/widgets/ImageField.ts",
        "src/widgets/DateField.ts",
        "src/widgets/PhoneField.ts",
        "src/widgets/TextareaField.ts",
        "src/widgets/DefaultField.ts",
        "src/widgets/BooleanRadioField.ts",
        "src/widgets/BooleanToggleField.ts",
        "src/views/kanban/kanban-card.ts",
        "src/widgets/password-match.ts",
        "src/widgets/Many2ManyTagsField.ts",
        "src/widgets/StatusbarField.ts",
        "src/widgets/PriorityField.ts",
        "src/views/form/form-sheet.ts",
        "src/views/shared/collection-bar-panels.ts",
      ],
      thresholds: {
        lines: 90,
        statements: 90,
        functions: 90,
        branches: 70,
      },
    },
  },
});
