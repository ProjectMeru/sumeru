package module

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/sdk/platformmsg"
)

func (addon *Addon) syncCSVFieldAccess(ctx context.Context) error {
	csvPath := filepath.Join(addon.Path, "sys.field.access.csv")
	if _, err := os.Stat(csvPath); err != nil {
		csvPath = filepath.Join(addon.Path, "security", "sys.field.access.csv")
		if _, err := os.Stat(csvPath); err != nil {
			return nil
		}
	}

	csvFile, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	if _, err := csvReader.Read(); err != nil {
		return err
	}

	moduleName := addon.Manifest.Name
	for {
		csvRecord, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(csvRecord) < 7 {
			continue
		}

		recordXmlId := csvRecord[0]
		accessName := strings.TrimSpace(csvRecord[1])
		modelName := strings.TrimSpace(csvRecord[2])
		fieldName := strings.TrimSpace(csvRecord[3])
		groupXmlId := strings.TrimSpace(csvRecord[4])
		permRead := csvRecord[5] == "1"
		permWrite := csvRecord[6] == "1"

		if modelName == "" || fieldName == "" {
			syncWarn(ctx, "Warning: sys.field.access %s missing model or field_name", recordXmlId)
			continue
		}
		if !registryHasField(modelName, fieldName) {
			syncWarn(ctx, "Warning: sys.field.access %s unknown field %s.%s", recordXmlId, modelName, fieldName)
			continue
		}
		if accessName == "" {
			accessName = fieldAccessDefaultName(modelName, fieldName, groupXmlId)
		}

		var groupId int
		if groupXmlId != "" {
			gid, resolveErr := resolveXMLIDInModule(ctx, moduleName, groupXmlId)
			if resolveErr != nil {
				syncWarn(ctx, "Warning: sys.field.access %s group %q unresolved: %v", recordXmlId, groupXmlId, resolveErr)
			}
			groupId = gid
		}

		accessValues := map[string]interface{}{
			"name":       accessName,
			"model":      modelName,
			"field_name": fieldName,
			"perm_read":  permRead,
			"perm_write": permWrite,
		}
		if groupId > 0 {
			accessValues["group_id"] = groupId
		}

		id, err := orm.Upsert(ctx, orm.RegistryModel("sys.field.access"), accessValues, "name")
		if err != nil {
			syncWarn(ctx, platformmsg.FmtGenericUpsertWarn, "sys.field.access", recordXmlId, err)
			continue
		}
		if err := linkXMLRecord(ctx, moduleName, recordXmlId, "sys.field.access", id); err != nil {
			continue
		}
	}
	return nil
}

func fieldAccessDefaultName(model, field, groupXML string) string {
	g := strings.TrimSpace(groupXML)
	if g == "" {
		g = "global"
	}
	return "field_access." + model + "." + field + "." + strings.ReplaceAll(g, ".", "_")
}

func registryHasField(modelName, fieldName string) bool {
	inst := orm.RegistryModel(modelName)
	if inst == nil {
		return false
	}
	for _, f := range inst.Fields() {
		if strings.TrimSpace(f.Name) == fieldName {
			return true
		}
	}
	return false
}

// ValidateFieldAccessOrphans warns about sys.field.access rows pointing at unknown models or fields.
func ValidateFieldAccessOrphans(ctx context.Context) {
	if orm.DB == nil {
		return
	}
	if _, ok := orm.Registry["sys.field.access"]; !ok {
		return
	}
	tbl := orm.MustQuotedTableName("sys.field.access")
	rows, err := orm.DB.QueryContext(ctx, `SELECT name, model, field_name FROM `+tbl)
	if err != nil {
		syncWarn(ctx, "Warning: field access orphan check failed: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var name, model, field string
		if err := rows.Scan(&name, &model, &field); err != nil {
			syncWarn(ctx, "Warning: field access orphan scan: %v", err)
			return
		}
		model = strings.TrimSpace(model)
		field = strings.TrimSpace(field)
		if orm.RegistryModel(model) == nil {
			syncWarn(ctx, "Warning: orphan sys.field.access %q model %q not registered", name, model)
			continue
		}
		if !registryHasField(model, field) {
			syncWarn(ctx, "Warning: orphan sys.field.access %q field %s.%s not on model", name, model, field)
		}
	}
	_ = rows.Err()
}
