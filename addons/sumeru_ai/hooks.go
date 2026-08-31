package sumeru_ai

import (
	"context"
	"html/template"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

func init() {
	registerHooks()
}

func registerHooks() {
	orm.RegisterSearchInterceptor(func(ctx context.Context, model string, domain [][]interface{}) ([][]interface{}, error) {
		return domain, nil
	})

	render.RegisterShellHook(func(ctx context.Context, vr *render.ViewRecordData, ro bool) template.HTML {
		return template.HTML(`
			<div id="sumeru-ai-assistant">
				<button type="button" class="sum-ai-fab" aria-label="AI assistant">
					<svg class="sum-ai-fab-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"></path></svg>
					<span class="sum-ai-fab-tooltip">AI Assistant</span>
				</button>
			</div>
		`)
	})

	render.RegisterNotebookHook("core.partner", "ai insights", func(ctx context.Context, vr *render.ViewRecordData, ro bool) template.HTML {
		return template.HTML(`
			<div class="sum-ai-panel">
				<h3 class="sum-ai-panel-title">
					<svg class="sum-ai-panel-title-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
					AI relationship insights
				</h3>
				<p class="sum-ai-panel-text">
					This partner shows high potential for churn. AI suggests reaching out with a personalized discount code.
				</p>
			</div>
		`)
	})
}
