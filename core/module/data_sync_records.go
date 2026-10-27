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
	"sys.sequence":         {conflictKey: "code"},
	"account.account":      {conflictKey: "code"},
	"core.country.state":   {naturalKeys: naturalKeyNameCountry},
	"core.city":            {naturalKeys: naturalKeyNameCountryState},
	"account.tax":          {naturalKeys: naturalKeyName},
	"account.payment.term": {naturalKeys: naturalKeyName},
	"account.move":         {naturalKeys: naturalKeyName},
	"core.partner":         {naturalKeys: naturalKeyName},
	"core.currency.rate":   {naturalKeys: naturalKeyCurrencyRate},
	"uom.uom":              {naturalKeys: naturalKeyUom},
	"account.journal":      {naturalKeys: naturalKeyCode},
	"account.move.line":    {naturalKeys: naturalKeyMoveLine},
	"mail.activity.type":   {naturalKeys: naturalKeyName},
	"sys.report.action":    {naturalKeys: naturalKeyReportAction},
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

func naturalKeyReportAction(fieldValues map[string]interface{}) map[string]interface{} {
	if tpl, ok := fieldValues["template_path"]; ok && tpl != nil {
		return map[string]interface{}{"template_path": tpl}
	}
	return map[string]interface{}{"name": fieldValues["name"]}
}

// naturalKeyCurrencyRate keys core.currency.rate rows by (currency, date_from):
// one effective rate row per currency and start date.
func naturalKeyCurrencyRate(fieldValues map[string]interface{}) map[string]interface{} {
	criteria := map[string]interface{}{"currency_id": fieldValues["currency_id"]}
	if df, ok := fieldValues["date_from"]; ok && df != nil {
		criteria["date_from"] = df
	}
	return criteria
}

// naturalKeyUom keys uom.uom rows by (name, category): unit names are only
// unique inside their category (e.g. "oz" weight vs fluid ounce).
func naturalKeyUom(fieldValues map[string]interface{}) map[string]interface{} {
	criteria := map[string]interface{}{"name": fieldValues["name"]}
	if cid, ok := fieldValues["category_id"]; ok && cid != nil {
		criteria["category_id"] = cid
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
	if 0 == len(fieldMapStrings) {
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
			finishRegistryRecordSync(ctx, moduleName, xmlRecord, existingID, fieldMapStrings, impliedEval)
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
				finishRegistryRecordSync(ctx, moduleName, xmlRecord, int(eid), fieldMapStrings, impliedEval)
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
	finishRegistryRecordSync(ctx, moduleName, xmlRecord, id, fieldMapStrings, impliedEval)
	if err := linkXMLRecord(ctx, moduleName, xmlRecord.ID, xmlRecord.Model, id); err != nil {
		return
	}
}

func finishRegistryRecordSync(ctx context.Context, moduleName string, xmlRecord parser.Record, id int, fieldMapStrings map[string]string, impliedEval string) {
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
}

// ConvertRecordScalar coerces XML/form string values into types used for registry upserts.
func ConvertRecordScalar(ctx context.Context, moduleName, model, column, rawValue string) interface{} {
	value := strings.TrimSpace(rawValue)

	if strings.HasPrefix(column, "perm_") {
		return parseBoolish(value)
	}
	if isSyncForeignKeyColumn(column) {
		return coerceSyncForeignKey(ctx, moduleName, value)
	}
	if "active" == column || strings.HasSuffix(column, "_active") {
		return parseBoolish(value)
	}
	if evaluated, err := eval.SafeEval(value); err == nil {
		if out, ok := coerceSyncEvalResult(ctx, moduleName, evaluated); ok {
			return out
		}
	}
	return value
}

func isSyncForeignKeyColumn(column string) bool {
	switch column {
	case "group_id", "user_id", "rule_id", "implied_group_id", "parent_id",
		"category_id", "country_id", "state_id", "city_id":
		return true
	default:
		return strings.HasSuffix(column, "_id")
	}
}

func coerceSyncForeignKey(ctx context.Context, moduleName, value string) interface{} {
	if isEmptyRef(value) {
		return nil
	}
	if id, err := resolveXMLIDInModule(ctx, moduleName, value); err == nil && 0 < id {
		return id
	}
	if n, err := strconv.ParseInt(value, 10, 64); err == nil {
		return n
	}
	return nil
}

func coerceSyncEvalResult(ctx context.Context, moduleName string, evaluated interface{}) (out interface{}, ok bool) {
	switch typed := evaluated.(type) {
	case bool, int64, float64, string:
		return typed, true
	case []interface{}:
		return resolveEvalTuple(ctx, moduleName, typed), true
	case nil:
		return nil, true
	default:
		return nil, false
	}
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
	if 0 == len(parts) {
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
