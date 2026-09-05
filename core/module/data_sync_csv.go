package module

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"

	"sumeru/core/orm"
	"sumeru/core/sdk/platformmsg"
)

func (addon *Addon) syncCSVModelAccess(ctx context.Context) error {
	csvPath := filepath.Join(addon.Path, "sys.access.csv")
	if _, err := os.Stat(csvPath); err != nil {
		csvPath = filepath.Join(addon.Path, "security", "sys.access.csv")
		if _, err := os.Stat(csvPath); err != nil {
			return nil // No CSV ACL file found
		}
	}

	csvFile, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	// Skip header: id,name,model_id:id,group_id:id,perm_read,perm_write,perm_create,perm_unlink
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
		if len(csvRecord) < 8 {
			continue
		}

		recordXmlId := csvRecord[0]
		accessName := csvRecord[1]
		modelName := csvRecord[2]
		groupXmlId := csvRecord[3]
		permRead := csvRecord[4] == "1"
		permWrite := csvRecord[5] == "1"
		permCreate := csvRecord[6] == "1"
		permUnlink := csvRecord[7] == "1"

		var groupId int
		if groupXmlId != "" {
			gid, resolveErr := resolveXMLIDInModule(ctx, moduleName, groupXmlId)
			if resolveErr != nil {
				syncWarn(ctx, "Warning: sys.access %s group %q unresolved: %v", recordXmlId, groupXmlId, resolveErr)
			}
			groupId = gid
		}

		accessValues := map[string]interface{}{
			"name":        accessName,
			"model":       modelName,
			"perm_read":   permRead,
			"perm_write":  permWrite,
			"perm_create": permCreate,
			"perm_unlink": permUnlink,
		}
		if groupId > 0 {
			accessValues["group_id"] = groupId
		}

		id, err := orm.Upsert(ctx, orm.RegistryModel("sys.access"), accessValues, "name")
		if err != nil {
			syncWarn(ctx, platformmsg.FmtGenericUpsertWarn, "sys.access", recordXmlId, err)
			continue
		}
		if err := linkXMLRecord(ctx, moduleName, recordXmlId, "sys.access", id); err != nil {
			continue
		}
	}
	return nil
}
