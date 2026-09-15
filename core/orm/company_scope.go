package orm

import (
	"context"
	"strings"
)

// CompanyIsolationDomain is the default multi-company OR-domain (shared rows OR allowed companies).
func CompanyIsolationDomain() [][]interface{} {
	return [][]interface{}{
		{"|"},
		{"company_id", "=", false},
		{"company_id", "in", "$company_ids"},
	}
}

var (
	companySharedModels = map[string]bool{}
	companyGlobalModels = map[string]bool{
		"core.company": true,
		"core.user":    true,
	}
)

// SetModelCompanyShared marks a model as globally shared (no auto company isolation).
func SetModelCompanyShared(modelName string, shared bool) {
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return
	}
	if shared {
		companySharedModels[modelName] = true
		return
	}
	delete(companySharedModels, modelName)
}

// ModelHasCompanyID reports whether model has a stored company_id column.
func ModelHasCompanyID(modelName string) bool {
	fd := FieldDef(modelName, "company_id")
	if fd == nil || IsVirtualField(*fd) {
		return false
	}
	return true
}

// ModelRequiresCompanyIsolation reports whether automatic company record rules apply.
func ModelRequiresCompanyIsolation(modelName string) bool {
	if companySharedModels[modelName] || companyGlobalModels[modelName] {
		return false
	}
	if strings.HasPrefix(modelName, "sys.") {
		return false
	}
	return ModelHasCompanyID(modelName)
}

func appendCompanyIsolationRule(parts ruleDomainParts, model string, dc DomainContext) ruleDomainParts {
	if !ModelRequiresCompanyIsolation(model) {
		return parts
	}
	dom := SubstituteDomainContext(CompanyIsolationDomain(), dc)
	if len(dom) > 0 {
		parts.globals = append(parts.globals, dom)
	}
	return parts
}

func domainContextForUser(ctx context.Context, uid int) DomainContext {
	dc := DomainContext{UID: uid}
	if cids, err := UserCompanyIDs(ctx, uid); err == nil {
		dc.CompanyIDs = cids
	}
	active := CompanyIDFromContext(ctx)
	if active > 0 && int64InSlice(active, dc.CompanyIDs) {
		dc.CompanyID = active
	} else if len(dc.CompanyIDs) > 0 {
		dc.CompanyID = dc.CompanyIDs[0]
	}
	return dc
}

func int64InSlice(v int64, list []int64) bool {
	for _, id := range list {
		if id == v {
			return true
		}
	}
	return false
}
