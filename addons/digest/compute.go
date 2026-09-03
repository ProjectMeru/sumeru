package digest

import (
	"context"
	"fmt"

	"sumeru/core/orm"
)

// ComputeKPI evaluates a digest KPI compute_code and returns a numeric value.
// ponytail: stub registry only; add handlers as modules register KPIs.
func ComputeKPI(ctx context.Context, code string) (float64, error) {
	switch code {
	case "crm.lead_count":
		return countModel(ctx, "crm.lead", nil)
	case "crm.lead_won_count":
		return countModel(ctx, "crm.lead", [][]interface{}{{"won_status", "=", "won"}})
	case "crm.opportunity_count":
		return countModel(ctx, "crm.lead", [][]interface{}{{"type", "=", "opportunity"}})
	default:
		return 0, fmt.Errorf("digest: unknown compute code %q", code)
	}
}

func countModel(ctx context.Context, model string, domain [][]interface{}) (float64, error) {
	if _, ok := orm.Registry[model]; !ok {
		return 0, nil
	}
	rows, err := orm.Search(ctx, model, domain)
	if err != nil {
		return 0, err
	}
	return float64(len(rows)), nil
}
