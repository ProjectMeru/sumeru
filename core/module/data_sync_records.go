package module

import (
	"context"
	"strconv"
	"strings"

	"sumeru/core/engine/eval"
	"sumeru/core/sdk/platformmsg"
	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func syncGenericRegistryRecord(ctx context.Context, moduleName string, xmlRecord parser.Record) {
	if xmlRecord.Model == "sys.action.window" || xmlRecord.Model == "sys.view" {
		return
	}
	if strings.HasPrefix(xmlRecord.Model, "sys.") {
		modelInstance, ok := orm.Registry[xmlRecord.Model]
		if !ok || modelInstance == nil {
			return
		}
		syncRegistryRecordByModel(ctx, moduleName, xmlRecord, modelInstance)
		return
	}
	modelInstance, ok := orm.Registry[xmlRecord.Model]
	if !ok || modelInstance == nil {
		return
	}
	syncRegistryRecordByModel(ctx, moduleName, xmlRecord, modelInstance)
}

func syncRegistryRecordByModel(ctx context.Context, moduleName string, xmlRecord parser.Record, modelInstance orm.Model) {
	fieldMapStrings := parser.RecordFieldMap(xmlRecord)
	if len(fieldMapStrings) == 0 {
		return
	}
	impliedEval := strings.TrimSpace(fieldMapStrings["implied_ids"])
	fieldValues := map[string]interface{}{}
	for key, val := range fieldMapStrings {
		if key == "implied_ids" || key == "groups" {
			continue
		}
		fieldValues[key] = ConvertRecordScalar(ctx, moduleName, xmlRecord.Model, key, val)
	}

	// Prefer external id when already synced (states/cities have no natural unique key).
	if existingID, _, err := orm.ResolveXmlId(ctx, moduleName+"."+xmlRecord.ID); err == nil && existingID > 0 {
		if err := orm.UpdateRecordByID(ctx, xmlRecord.Model, existingID, fieldValues); err == nil {
			return
		}
		// Stale xml id — fall through to create and refresh model_data.
	}

	conflictColumn := "name"
	switch xmlRecord.Model {
	case "core.user":
		conflictColumn = "login"
	case "core.country", "core.lang", "account.account":
		conflictColumn = "code"
	case "core.country.state", "core.city", "account.tax", "account.payment.term",
		"account.move", "account.move.line", "core.partner", "account.journal":
		conflictColumn = ""
	}

	// Match existing by natural keys when Upsert cannot use a UNIQUE column.
	if xmlRecord.Model == "core.country.state" || xmlRecord.Model == "core.city" ||
		xmlRecord.Model == "account.tax" || xmlRecord.Model == "account.payment.term" ||
		xmlRecord.Model == "account.move" || xmlRecord.Model == "core.partner" ||
		xmlRecord.Model == "account.journal" || xmlRecord.Model == "account.move.line" {
		criteria := map[string]interface{}{}
		switch xmlRecord.Model {
		case "core.country.state", "core.city":
			criteria["name"] = fieldValues["name"]
			if cid, ok := fieldValues["country_id"]; ok && cid != nil {
				criteria["country_id"] = cid
			}
			if xmlRecord.Model == "core.city" {
				if sid, ok := fieldValues["state_id"]; ok && sid != nil {
					criteria["state_id"] = sid
				}
			}
		case "account.tax", "account.payment.term", "core.partner", "account.move":
			criteria["name"] = fieldValues["name"]
		case "account.journal":
			criteria["code"] = fieldValues["code"]
		case "account.move.line":
			criteria["name"] = fieldValues["name"]
			if mid, ok := fieldValues["move_id"]; ok && mid != nil {
				criteria["move_id"] = mid
			}
		}
		if existing, err := orm.SearchOne(ctx, xmlRecord.Model, criteria); err == nil {
			if eid, ok := orm.CoerceInt64(existing["id"]); ok && eid > 0 {
				if err := orm.UpdateRecordByID(ctx, xmlRecord.Model, int(eid), fieldValues); err != nil {
					syncWarn(ctx, platformmsg.FmtGenericUpsertWarn, xmlRecord.Model, xmlRecord.ID, err)
					return
				}
				_ = linkXMLRecord(ctx, moduleName, xmlRecord.ID, xmlRecord.Model, int(eid))
				return
			}
		}
	}

	var id int
	var err error
	if conflictColumn == "" {
		id, err = orm.Create(ctx, modelInstance, fieldValues)
	} else {
		if _, ok := fieldValues[conflictColumn]; !ok {
			return
		}
		id, err = orm.Upsert(ctx, modelInstance, fieldValues, conflictColumn)
	}
	if err != nil {
		syncWarn(ctx, platformmsg.FmtGenericUpsertWarn, xmlRecord.Model, xmlRecord.ID, err)
		return
	}
	if xmlRecord.Model == "core.group" {
		if impliedEval != "" {
			if err := syncCoreGroupImpliedFromEval(ctx, moduleName, id, impliedEval); err != nil {
				syncWarn(ctx, "Warning: core.group implied_ids %s (%s): %v", xmlRecord.ID, moduleName, err)
			}
		}
		if err := EnsureSystemImpliesManagerGroup(ctx, moduleName, xmlRecord.ID, id); err != nil {
			syncWarn(ctx, "Warning: system→manager imply %s (%s): %v", xmlRecord.ID, moduleName, err)
		}
	}
	if xmlRecord.Model == "sys.rule" {
		if groupsEval := strings.TrimSpace(fieldMapStrings["groups"]); groupsEval != "" {
			if err := syncSysRuleGroupsFromEval(ctx, moduleName, id, groupsEval); err != nil {
				syncWarn(ctx, "Warning: sys.rule groups %s (%s): %v", xmlRecord.ID, moduleName, err)
			}
		}
	}
	_ = linkXMLRecord(ctx, moduleName, xmlRecord.ID, xmlRecord.Model, id)
}

// ConvertRecordScalar coerces XML/form string values into types used for registry upserts.
func ConvertRecordScalar(ctx context.Context, moduleName, model, column, rawValue string) interface{} {
	trimmedValue := strings.TrimSpace(rawValue)
	if strings.HasPrefix(column, "perm_") {
		if boolValue, err := strconv.ParseBool(trimmedValue); err == nil {
			return boolValue
		}
		return strings.EqualFold(trimmedValue, "true") || trimmedValue == "1"
	}
	if column == "group_id" || column == "user_id" || column == "rule_id" || column == "implied_group_id" || column == "parent_id" || column == "category_id" || column == "country_id" || column == "state_id" || column == "city_id" || strings.HasSuffix(column, "_id") {
		if trimmedValue == "" || strings.EqualFold(trimmedValue, "false") || trimmedValue == "0" {
			return nil
		}
		if id, err := resolveXMLIDInModule(ctx, moduleName, trimmedValue); err == nil && id > 0 {
			return id
		}
		if n, err := strconv.ParseInt(trimmedValue, 10, 64); err == nil {
			return n
		}
		return nil
	}
	if column == "active" || strings.HasSuffix(column, "_active") {
		if boolValue, err := strconv.ParseBool(trimmedValue); err == nil {
			return boolValue
		}
		return strings.EqualFold(trimmedValue, "true") || trimmedValue == "1"
	}
	if v, err := eval.SafeEval(trimmedValue); err == nil {
		switch typed := v.(type) {
		case bool, int64, float64, string:
			return typed
		case []interface{}:
			return resolveEvalTuple(ctx, moduleName, typed)
		case nil:
			return nil
		}
	}
	// Numeric / integer literals from eval="…"
	if n, err := strconv.ParseInt(trimmedValue, 10, 64); err == nil && !strings.Contains(trimmedValue, ".") {
		return n
	}
	if f, err := strconv.ParseFloat(trimmedValue, 64); err == nil {
		return f
	}
	return trimmedValue
}

func resolveEvalTuple(ctx context.Context, moduleName string, parts []interface{}) interface{} {
	if len(parts) == 0 {
		return parts
	}
	out := make([]interface{}, len(parts))
	for i, p := range parts {
		out[i] = p
	}
	if cmd, ok := out[0].(int64); ok && cmd == 4 && len(out) >= 2 {
		if ref, ok := out[1].(string); ok {
			if id, err := resolveXMLIDInModule(ctx, moduleName, strings.Trim(ref, `"'`)); err == nil && id > 0 {
				return id
			}
		}
	}
	return out
}
