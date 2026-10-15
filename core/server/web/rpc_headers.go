package web

import (
	"net/http"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/server/api"
	"sumeru/core/server/config"
)

const (
	sumeruDatabaseHeader = "X-Sumeru-Database"
	sumeruTenantHeader   = "X-Sumeru-Tenant"
	rpcAuthLoginSegment  = "auth.login"
)

func validateRPCDatabaseHeader(r *http.Request) (ok bool, resp api.APIResponse, status int) {
	got := strings.TrimSpace(r.Header.Get(sumeruDatabaseHeader))
	want := strings.TrimSpace(config.AppConfig.DbName)
	if got == "" {
		return false, api.Fail(api.CodeValidationError, "X-Sumeru-Database is required", map[string]interface{}{"header": sumeruDatabaseHeader}), http.StatusBadRequest
	}
	if want == "" || got != want {
		return false, api.Fail(api.CodeValidationError, "Invalid database", map[string]interface{}{"header": sumeruDatabaseHeader}), http.StatusBadRequest
	}
	return true, api.APIResponse{}, 0
}

func resolveRPCTenantCompanyID(r *http.Request) (companyID int64, ok bool, resp api.APIResponse, status int) {
	tenant := strings.TrimSpace(r.Header.Get(sumeruTenantHeader))
	if tenant == "" {
		return 0, false, api.Fail(api.CodeValidationError, "X-Sumeru-Tenant is required", map[string]interface{}{"header": sumeruTenantHeader}), http.StatusBadRequest
	}
	cid, err := orm.ResolveCompanyTenant(r.Context(), tenant)
	if err != nil {
		return 0, false, api.Fail(api.CodeValidationError, err.Error(), map[string]interface{}{"header": sumeruTenantHeader}), http.StatusBadRequest
	}
	return cid, true, api.APIResponse{}, 0
}

func ensureUserAllowedTenant(r *http.Request, uid int, companyID int64) (ok bool, resp api.APIResponse, status int) {
	if uid <= 0 {
		return false, api.Fail(api.CodeUnauthorized, "Unauthorized", nil), http.StatusUnauthorized
	}
	if !orm.UserAllowedCompany(r.Context(), uid, companyID) {
		return false, api.Fail(api.CodeAccessDenied, "Tenant not allowed for user", map[string]interface{}{"company_id": companyID}), http.StatusForbidden
	}
	return true, api.APIResponse{}, 0
}

func rpcUnauthorizedMessage(r *http.Request) (message string, invalidKey bool) {
	if apiKeyFromRequest(r) != "" {
		return "Invalid apikey", true
	}
	return "Unauthorized", false
}
