package orm

import "context"

// UserIsPortalOnly reports portal group without internal user group.
func UserIsPortalOnly(ctx context.Context, uid int) bool {
	if uid <= 0 || uid == superuserUID {
		return false
	}
	if !UserHasGroupXML(ctx, uid, "base.group_portal") {
		return false
	}
	return !UserHasGroupXML(ctx, uid, "base.group_user")
}

var portalRPCAllowlist = map[string]bool{
	"core.partner": true,
}

// EnforcePortalRPCModel restricts portal-only users to allowlisted RPC models.
func EnforcePortalRPCModel(ctx context.Context, model string) error {
	uid := UIDFromContext(ctx)
	if uid <= 0 || !UserIsPortalOnly(ctx, uid) {
		return nil
	}
	if portalRPCAllowlist[model] {
		return nil
	}
	return &AccessDeniedError{Model: model, Operation: "rpc"}
}

// PortalAllowsModel reports whether portal UI/RPC may access model for op (read|write).
func PortalAllowsModel(model, op string) bool {
	if !portalRPCAllowlist[model] {
		return false
	}
	switch op {
	case "read", "write", "create":
		return true
	default:
		return false
	}
}
