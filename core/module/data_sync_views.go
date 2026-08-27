package module

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/viewinherit"
	"sumeru/core/orm"
	"sumeru/core/sdk/platformmsg"
)

// viewArchXML persists the full parsed view (header, sheet, notebook, etc.) for sys.view.arch.
func viewArchXML(viewDef *parser.View) string {
	if viewDef == nil {
		return "<view/>"
	}
	marshaledXml, err := xml.Marshal(viewDef)
	if err != nil {
		return fmt.Sprintf("<view model=\"%s\" type=\"%s\"></view>", viewDef.Model, viewDef.Type)
	}
	return string(marshaledXml)
}

// upsertSysViewFromRecord persists <record model="sys.view">…</record> data (non-inherit rows).
// Inline <view> elements use upsertInlineViewDef instead; inherit-only records are handled elsewhere.
func upsertSysViewFromRecord(ctx context.Context, moduleName string, xmlRecord parser.Record) {
	recordValues := recordValuesFromXML(xmlRecord, "inherit_id")
	if _, ok := recordValues["name"]; !ok || strings.TrimSpace(orm.AsString(recordValues["name"])) == "" {
		recordValues["name"] = xmlRecord.ID
	}
	modelName := strings.TrimSpace(orm.AsString(recordValues["model"]))
	if modelName == "" {
		syncWarn(ctx, "Warning: sys.view record %s (module %s): model is required", xmlRecord.ID, moduleName)
		return
	}
	arch := strings.TrimSpace(orm.AsString(recordValues["arch"]))
	if arch == "" {
		syncWarn(ctx, "Warning: sys.view record %s (module %s): arch is required", xmlRecord.ID, moduleName)
		return
	}
	recordValues["arch"] = arch

	vt := strings.TrimSpace(strings.ToLower(orm.AsString(recordValues["type"])))
	if vt == "" {
		vt = InferSysViewTypeFromArch(arch)
	}
	if vt == "" {
		syncWarn(ctx, "Warning: sys.view record %s (module %s): could not infer type from arch", xmlRecord.ID, moduleName)
		return
	}
	recordValues["type"] = vt

	id, err := orm.Upsert(ctx, orm.RegistryModel("sys.view"), recordValues, "name")
	if err != nil {
		syncWarn(ctx, platformmsg.FmtGenericUpsertWarn, "sys.view", xmlRecord.ID, err)
		return
	}
	_ = linkXMLRecord(ctx, moduleName, xmlRecord.ID, "sys.view", id)
}

func InferSysViewTypeFromArch(arch string) string {
	a := strings.TrimSpace(arch)
	if a == "" {
		return ""
	}
	la := strings.ToLower(a)
	switch {
	case strings.HasPrefix(la, "<list"):
		return "list"
	case strings.HasPrefix(la, "<form"):
		return "form"
	case strings.HasPrefix(la, "<kanban"):
		return "kanban"
	case strings.HasPrefix(la, "<search"):
		return "search"
	case strings.HasPrefix(la, "<graph"):
		return "graph"
	case strings.HasPrefix(la, "<calendar"):
		return "calendar"
	case strings.HasPrefix(la, "<pivot"):
		return "pivot"
	case strings.HasPrefix(la, "<view"):
		if v, err := parser.ParseViewFromArch(a); err == nil {
			return strings.ToLower(strings.TrimSpace(v.Type))
		}
	}
	return ""
}

func upsertInlineViewDef(ctx context.Context, moduleName string, viewDef *parser.View) {
	viewArchitecture := viewArchXML(viewDef)
	viewName := strings.TrimSpace(viewDef.ID)
	if viewName == "" {
		viewName = viewDef.Model + "." + viewDef.Type
	}
	viewType := strings.TrimSpace(strings.ToLower(viewDef.Type))
	id, err := orm.Upsert(ctx, orm.RegistryModel("sys.view"), map[string]interface{}{
		"name":     viewName,
		"model":    viewDef.Model,
		"type":     viewType,
		"arch":     viewArchitecture,
		"priority": viewDef.Priority,
	}, "name")
	if err == nil {
		_ = linkXMLRecord(ctx, moduleName, viewDef.ID, "sys.view", id)
	}
}

// applySysUIViewInherit merges an sys.view inherit <record> into the parent view row (same DB id).
func applySysUIViewInherit(ctx context.Context, moduleName string, xmlRecord parser.Record) error {
	fieldMap := parser.RecordFieldMap(xmlRecord)
	inheritReference := strings.TrimSpace(fieldMap["inherit_id"])
	architectureFragment := fieldMap["arch"]
	if inheritReference == "" {
		return fmt.Errorf("inherit_id missing on record %q", xmlRecord.ID)
	}
	if strings.TrimSpace(architectureFragment) == "" {
		return fmt.Errorf("arch missing on inherit record %q", xmlRecord.ID)
	}
	parentID, err := resolveXMLIDInModule(ctx, moduleName, inheritReference)
	if err != nil || parentID == 0 {
		return fmt.Errorf("resolve inherit_id %q: %w", inheritReference, err)
	}
	parentView, err := orm.SearchOne(ctx, "sys.view", map[string]interface{}{"id": parentID})
	if err != nil {
		return fmt.Errorf("load parent view id %d: %w", parentID, err)
	}
	parentArchitecture := orm.AsString(parentView["arch"])
	if strings.TrimSpace(parentArchitecture) == "" {
		return fmt.Errorf("parent view %d has empty arch", parentID)
	}
	mergedArchitecture, err := viewinherit.ApplyInheritArch(parentArchitecture, architectureFragment)
	if err != nil {
		return fmt.Errorf("merge inherit %q: %w", xmlRecord.ID, err)
	}
	viewTableName := orm.MustQuotedTableName("sys.view")
	if _, err := orm.DB.ExecContext(ctx, `UPDATE `+viewTableName+` SET arch = $1 WHERE id = $2`, mergedArchitecture, parentID); err != nil {
		return err
	}
	if xmlRecord.ID != "" {
		if _, err := orm.Upsert(ctx, orm.RegistryModel("sys.model.data"), map[string]interface{}{
			"module":  moduleName,
			"name":    xmlRecord.ID,
			"model":   "sys.view",
			"core_id": parentID,
		}, "name"); err != nil {
			return err
		}
	}
	return nil
}
