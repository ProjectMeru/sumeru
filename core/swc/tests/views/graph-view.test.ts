import { beforeEach, describe, expect, it, vi } from "vitest";
import type { SwcEnv } from "../../src/runtime/env.js";
import type { SwcViewArch, SwcWorkspacePayload } from "../../src/types/workspace.js";
import { flushScheduledRenders } from "../../src/runtime/scheduler.js";
import { runWillStart } from "../../src/runtime/lifecycle.js";

const chartCtor = vi.fn();
const chartDestroy = vi.fn();
const chartUpdate = vi.fn();

vi.mock("chart.js/auto", () => ({
  default: class MockChart {
    data = { labels: [] as string[], datasets: [{ data: [] as number[], label: "" }] };
    constructor(_canvas: HTMLCanvasElement, config: { type: string }) {
      chartCtor(config.type);
    }
    destroy = chartDestroy;
    update = chartUpdate;
  },
}));

import { GraphView } from "../../src/views/graph/GraphView.js";

function graphPayload(chart = "bar"): SwcWorkspacePayload {
  return {
    actionId: 1,
    menuId: "2",
    viewType: "graph",
    model: "demo.model",
    recordId: 0,
    formEdit: false,
    csrfToken: "",
    arch: {
      type: "graph",
      model: "demo.model",
      fields: [
        { name: "create_date", pivotType: "row" },
        { name: "amount", pivotType: "measure" },
      ],
      graph: { chart },
    } as SwcViewArch,
    records: [],
    viewTabs: [],
    breadcrumbs: [],
  };
}

function graphEnv(readGroup: SwcEnv["services"]["rpc"]["readGroup"]): SwcEnv {
  return {
    bootstrap: {} as never,
    services: {
      rpc: { readGroup },
      action: { openRecord: vi.fn() },
    },
  } as unknown as SwcEnv;
}

function groupsOf(view: GraphView): Record<string, unknown>[] {
  return (view as unknown as { groups: Record<string, unknown>[] }).groups;
}

describe("GraphView", () => {
  beforeEach(() => {
    chartCtor.mockClear();
    chartDestroy.mockClear();
    chartUpdate.mockClear();
  });

  it("creates chart on first render after readGroup resolves", async () => {
    const readGroup = vi.fn(async () => [{ create_date: "2026-01", amount: 5 }]);
    const view = new GraphView({ payload: graphPayload() }, graphEnv(readGroup));
    view.callSetup();
    view.render();
    await runWillStart(view);
    await flushScheduledRenders();
    expect(readGroup).toHaveBeenCalled();
    expect(chartCtor).toHaveBeenCalledWith("bar");
    view.destroy();
  });

  it("ignores stale readGroup responses", async () => {
    let resolveFirst!: (value: Record<string, unknown>[]) => void;
    const first = new Promise<Record<string, unknown>[]>((resolve) => {
      resolveFirst = resolve;
    });
    const readGroup = vi
      .fn()
      .mockImplementationOnce(() => first)
      .mockResolvedValueOnce([{ create_date: "fresh", amount: 9 }]);
    const view = new GraphView({ payload: graphPayload() }, graphEnv(readGroup));
    view.callSetup();
    view.render();
    void runWillStart(view);
    view.updateProps({ payload: graphPayload("line") });
    await flushScheduledRenders();
    expect(groupsOf(view)[0]?.create_date).toBe("fresh");
    resolveFirst([{ create_date: "stale", amount: 1 }]);
    await flushScheduledRenders();
    expect(groupsOf(view)[0]?.create_date).toBe("fresh");
    view.destroy();
  });

  it("recreates chart when chart type changes", async () => {
    const readGroup = vi.fn(async () => [{ create_date: "2026-01", amount: 3 }]);
    const view = new GraphView({ payload: graphPayload("bar") }, graphEnv(readGroup));
    view.callSetup();
    view.render();
    await runWillStart(view);
    await flushScheduledRenders();
    chartCtor.mockClear();
    chartDestroy.mockClear();
    view.updateProps({ payload: graphPayload("pie") });
    await vi.waitFor(async () => {
      await flushScheduledRenders();
      expect(chartCtor).toHaveBeenCalledWith("pie");
    });
    expect(chartDestroy).toHaveBeenCalled();
    view.destroy();
  });
});
