package mail

import (
	"context"
	"strings"

	"sumeru/core/sdk"
)

func filterParentReadable(ctx context.Context, model string, rows []map[string]interface{}) []map[string]interface{} {
	if model != "mail.message" && model != "mail.activity" {
		return rows
	}
	if sdk.SecurityBypass(ctx) {
		return rows
	}
	manager := sdk.UserHasGroupXML(ctx, sdk.SecurityUID(ctx), "mail.group_mail_manager")
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		parent := strings.TrimSpace(sdk.AsString(row["model"]))
		parentID := recordID(row["res_id"])
		if parent == "" || 0 >= parentID || parent == "mail.message" || parent == "mail.activity" {
			if model == "mail.message" && strings.TrimSpace(sdk.AsString(row["state"])) == "quarantine" && manager {
				out = append(out, row)
			}
			continue
		}
		if _, err := sdk.SearchOne(ctx, parent, map[string]interface{}{"id": int(parentID)}); err != nil {
			continue
		}
		out = append(out, row)
	}
	return out
}
