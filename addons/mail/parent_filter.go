package mail

import (
	"context"
	"strings"

	"sumeru/core/orm"
)

func filterParentReadable(ctx context.Context, model string, rows []map[string]interface{}) []map[string]interface{} {
	if model != "mail.message" && model != "mail.activity" {
		return rows
	}
	if orm.SecurityBypass(ctx) {
		return rows
	}
	manager := orm.UserHasGroupXML(ctx, orm.SecurityUID(ctx), "mail.group_mail_manager")
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		parent := strings.TrimSpace(orm.AsString(row["model"]))
		parentID := recordID(row["res_id"])
		if parent == "" || parentID <= 0 || parent == "mail.message" || parent == "mail.activity" {
			if model == "mail.message" && strings.TrimSpace(orm.AsString(row["state"])) == "quarantine" && manager {
				out = append(out, row)
			}
			continue
		}
		if _, err := orm.SearchOne(ctx, parent, map[string]interface{}{"id": int(parentID)}); err != nil {
			continue
		}
		out = append(out, row)
	}
	return out
}
