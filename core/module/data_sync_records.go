package module

import (
	"context"
	"strconv"
	"strings"

	"sumeru/core/engine/eval"
	"sumeru/core/engine/parser"
	"sumeru/core/orm"
	"sumeru/core/sdk/platformmsg"
)

type naturalKeyFunc func(fieldValues map[string]interface{}) map[string]interface{}

type recordSyncSpec struct {
	conflictKey string
	naturalKeys naturalKeyFunc
}

var recordSyncSpecs = map[string]recordSyncSpec{
	"core.user":            {conflictKey: "login"},
	"core.country":         {conflictKey: "code"},
	"core.lang":            {conflictKey: "code"},
	"account.account":      {conflictKey: "code"},
	"core.country.state":   {naturalKeys: naturalKeyNameCountry},
	"core.city":            {naturalKeys: naturalKeyNameCountryState},
	"account.tax":          {naturalKeys: naturalKeyName},
	"account.payment.term": {naturalKeys: naturalKeyName},
	"account.move":         {naturalKeys: naturalKeyName},
	"core.partner":         {naturalKeys: naturalKeyName},
	"account.journal":      {naturalKeys: naturalKeyCode},
	"account.move.line":    {naturalKeys: naturalKeyMoveLine},
}

func recordConflictColumn(model string) string {
	spec, ok := recordSyncSpecs[model]
	if !ok {
		return "name"
	}
	if spec.naturalKeys != nil {
		return ""
	}
	if spec.conflictKey != "" {
		return spec.conflictKey
	}
	return "name"
}

func naturalKeyName(fieldValues map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"name": fieldValues["name"]}
}

func naturalKeyNameCountry(fieldValues map[string]interface{}) map[string]interface{} {
	criteria := map[string]interface{}{"name": fieldValues["name"]}
	if cid, ok := fieldValues["country_id"]; ok && cid != nil {
		criteria["country_id"] = cid
	}
	return criteria
}

func naturalKeyNameCountryState(fieldValues map[string]interface{}) map[string]interface{} {
	criteria := naturalKeyNameCountry(fieldValues)
	if sid, ok := fieldValues["state_id"]; ok && sid != nil {
		criteria["state_id"] = sid
	}
	return criteria
}

func naturalKeyCode(fieldValues map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"code": fieldValues["code"]}
}

func naturalKeyMoveLine(fieldValues map[string]interface{}) map[string]interface{} {
	criteria := map[string]interface{}{"name": fieldValues["name"]}
	if mid, ok := fieldValues["move_id"]; ok && mid != nil {
		criteria["move_id"] = mid
	}
	return criteria
}

func syncGenericRegistryRecord(ctx context.Context, moduleName string, xmlRecord parser.Record) {
	if xmlRecord.Model == "sys.action.window" || xmlRecord.Model == "sys.action.url" || xmlRecord.Model == "sys.view" {
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

	if spec, ok := recordSyncSpecs[xmlRecord.Model]; ok && spec.naturalKeys != nil {
		if existing, err := orm.SearchOne(ctx, xmlRecord.Model, spec.naturalKeys(fieldValues)); err == nil {
			if eid, ok := orm.CoerceInt64(existing["id"]); ok && eid > 0 {
				if err := orm.UpdateRecordByID(ctx, xmlRecord.Model, int(eid), fieldValues); err != nil {
					syncWarn(ctx, platformmsg.FmtGenericUpsertWarn, xmlRecord.Model, xmlRecord.ID, err)
					return
				}
				if err := linkXMLRecord(ctx, moduleName, xmlRecord.ID, xmlRecord.Model, int(eid)); err != nil {
					return
				}
				return
			}
		}
	}

	conflictColumn := recordConflictColumn(xmlRecord.Model)
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
	if err := linkXMLRecord(ctx, moduleName, xmlRecord.ID, xmlRecord.Model, id); err != nil {
		return
	}
}

// ConvertRecordScalar coerces XML/form string values into types used for registry upserts.
func ConvertRecordScalar(ctx context.Context, moduleName, model, column, rawValue string) interface{} {
	trimmedValue := strings.TrimSpace(rawValue)
	if strings.HasPrefix(column, "perm_") {
		return parseBoolish(trimmedValue)
	}
	if column == "group_id" || column == "user_id" || column == "rule_id" || column == "implied_group_id" || column == "parent_id" || column == "category_id" || column == "country_id" || column == "state_id" || column == "city_id" || strings.HasSuffix(column, "_id") {
		if isEmptyRef(trimmedValue) {
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
		return parseBoolish(trimmedValue)
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
	return trimmedValue
}

func parseBoolish(value string) bool {
	if boolValue, err := strconv.ParseBool(value); err == nil {
		return boolValue
	}
	return strings.EqualFold(value, "true") || value == "1"
}

func isEmptyRef(value string) bool {
	return value == "" || strings.EqualFold(value, "false") || value == "0"
}

func resolveEvalTuple(ctx context.Context, moduleName string, parts []interface{}) interface{} {
	if len(parts) == 0 {
		return parts
	}
	if cmd, ok := parts[0].(int64); ok && cmd == 4 && len(parts) >= 2 {
		if ref, ok := parts[1].(string); ok {
			if id, err := resolveXMLIDInModule(ctx, moduleName, strings.Trim(ref, `"'`)); err == nil && id > 0 {
				return id
			}
		}
	}
	return parts
}
