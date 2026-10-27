package module

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/errcode"
	"sumeru/core/orm"
)

// resolveXMLIDInModule resolves module.external_id: uses full ref if it contains a dot,
// else current module's xml id, then a global name-only lookup.
func resolveXMLIDInModule(ctx context.Context, moduleName, ref string) (int, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return 0, nil
	}
	if strings.Contains(ref, ".") {
		id, _, err := orm.ResolveXmlId(ctx, ref)
		return id, err
	}
	if moduleName != "" {
		if id, _, err := orm.ResolveXmlId(ctx, moduleName+"."+ref); id != 0 {
			return id, nil
		} else if err != nil {
			return 0, err
		}
	}
	id, _, err := orm.ResolveXmlId(ctx, ref)
	return id, err
}

// linkedCoreID returns the row id stored for an XML id in this module.
func linkedCoreID(ctx context.Context, moduleName, model, xmlID string) (int, error) {
	if strings.TrimSpace(xmlID) == "" {
		return 0, nil
	}
	md, err := orm.SearchOne(ctx, "sys.model.data", map[string]interface{}{
		"module": moduleName,
		"model":  model,
		"name":   xmlID,
	})
	if err != nil {
		return 0, nil
	}
	cid, ok := orm.CoerceInt64(md["core_id"])
	if !ok || 0 >= cid {
		return 0, nil
	}
	return int(cid), nil
}

// upsertLinkedRecord updates the row already linked to xmlID, or creates one and links it.
func upsertLinkedRecord(ctx context.Context, moduleName, model, xmlID string, values map[string]interface{}) (int, error) {
	id, err := linkedCoreID(ctx, moduleName, model, xmlID)
	if err != nil {
		return 0, err
	}
	if id > 0 {
		if err := orm.UpdateRecordByID(ctx, model, id, values); err != nil {
			return 0, err
		}
		if err := linkXMLRecord(ctx, moduleName, xmlID, model, id); err != nil {
			return 0, err
		}
		return id, nil
	}
	newID, err := orm.Create(ctx, orm.RegistryModel(model), values)
	if err != nil {
		return 0, err
	}
	if err := linkXMLRecord(ctx, moduleName, xmlID, model, newID); err != nil {
		return 0, err
	}
	return newID, nil
}

// linkXMLRecord upserts sys.model.data for a synced XML record and logs failures.
func linkXMLRecord(ctx context.Context, moduleName, xmlID, model string, coreID int) error {
	if 0 >= coreID || xmlID == "" {
		return nil
	}
	_, err := orm.Upsert(ctx, orm.RegistryModel("sys.model.data"), map[string]interface{}{
		"module":  moduleName,
		"name":    xmlID,
		"model":   model,
		"core_id": coreID,
	}, "name")
	if err != nil {
		syncWarnCode(ctx, errcode.XMLLinkFailed, "Failed to link XML record %s.%s (%s id=%d): %v", moduleName, xmlID, model, coreID, err)
	}
	return err
}

func syncWarn(ctx context.Context, format string, args ...interface{}) {
	syncWarnCode(ctx, errcode.SyncPartial, format, args...)
}

func syncWarnCode(ctx context.Context, code, format string, args ...interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	applog.WarnCode(ctx, code, msg, applog.Event{
		Component: "module",
		Operation: "sync",
		Status:    "partial",
	})
}
