package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const debugRuleDomainMax = 512

// BuildDebugAccessTrace returns ACL/rule summary for the current user (support UI only).
func BuildDebugAccessTrace(ctx context.Context, uid int, model string) (map[string]interface{}, error) {
	model = strings.TrimSpace(model)
	if uid <= 0 {
		return nil, fmt.Errorf("login required")
	}
	if UserIsPortalOnly(ctx, uid) {
		return nil, fmt.Errorf("access denied")
	}
	if model == "" {
		return nil, fmt.Errorf("model required")
	}
	if RegistryModel(model) == nil {
		return nil, fmt.Errorf("unknown model %q", model)
	}
	if err := CheckModelAccess(ctx, uid, model, "read"); err != nil {
		return nil, err
	}

	dc := domainContextForUser(ctx, uid)
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return nil, err
	}

	readDenied, err := fieldAccessDenied(ctx, uid, model, "read")
	if err != nil {
		return nil, err
	}
	writeDenied, err := fieldAccessDenied(ctx, uid, model, "write")
	if err != nil {
		return nil, err
	}

	fields := []map[string]interface{}{}
	regModel := RegistryModel(model)
	for _, f := range regModel.Fields() {
		name := strings.TrimSpace(f.Name)
		if name == "" || name == "id" {
			continue
		}
		entry := map[string]interface{}{"name": name}
		if readDenied[name] {
			entry["read_denied"] = true
		}
		if writeDenied[name] {
			entry["write_denied"] = true
		}
		fields = append(fields, entry)
	}

	parts, err := loadRuleDomainParts(ctx, uid, model, "read")
	if err != nil {
		return nil, err
	}
	ruleSummary, truncated := summarizeRuleParts(parts)

	out := map[string]interface{}{
		"uid":              uid,
		"model":            model,
		"company_id":       dc.CompanyID,
		"allowed_companies": dc.CompanyIDs,
		"group_ids":        groups,
		"model_read_ok":    true,
		"fields":           fields,
		"rule_domain":      ruleSummary,
	}
	if truncated {
		out["rule_domain_truncated"] = true
	}
	return out, nil
}

func summarizeRuleParts(parts ruleDomainParts) (string, bool) {
	payload := map[string]interface{}{
		"globals": parts.globals,
		"groups":  parts.groups,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", false
	}
	s := string(raw)
	if len(s) <= debugRuleDomainMax {
		return s, false
	}
	return s[:debugRuleDomainMax], true
}
