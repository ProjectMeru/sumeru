package orm

import (
	"context"
	"fmt"
)

// UserSecurityGroupRow is one assignable group for the user form security section.
type UserSecurityGroupRow struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Selected     bool   `json:"selected"`
	IsUserType   bool   `json:"isUserType"`
	CategoryID   int    `json:"categoryId,omitempty"`
	CategoryName string `json:"categoryName,omitempty"`
	Exclusive    bool   `json:"exclusive,omitempty"`
}

// UserSecurityCompanyRow is one company checkbox for allowed companies.
type UserSecurityCompanyRow struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Selected bool   `json:"selected"`
}

// UserSecurityMeta is serialized on workspace payload for core.user forms.
type UserSecurityMeta struct {
	CanEdit       bool                     `json:"canEdit"`
	IsNew         bool                     `json:"isNew"`
	Groups        []UserSecurityGroupRow   `json:"groups"`
	Companies     []UserSecurityCompanyRow `json:"companies"`
	PasswordLabel string                   `json:"passwordLabel"`
}

// BuildUserSecurityMeta loads group/company pickers for the user security UI.
func BuildUserSecurityMeta(ctx context.Context, actor, targetUserID int) (UserSecurityMeta, error) {
	out := UserSecurityMeta{
		PasswordLabel: "Change password",
	}
	if 0 >= targetUserID {
		out.IsNew = true
		out.PasswordLabel = "Set initial password"
	}
	if !UserHasGroupXML(ctx, actor, "base.group_system") {
		return out, nil
	}
	out.CanEdit = true

	typeXML := []string{"base.group_user", "base.group_portal", "base.group_public"}
	typeIDSet := map[int]struct{}{}
	for _, x := range typeXML {
		gid, _, err := ResolveXmlId(ctx, x)
		if err == nil && gid > 0 {
			typeIDSet[gid] = struct{}{}
		}
	}

	selectedGroups := map[int]struct{}{}
	if targetUserID > 0 {
		rows, err := DB.QueryContext(ctx,
			`SELECT group_id FROM `+MustQuotedTableName(tableGroupUserRel)+` WHERE user_id = $1`, targetUserID)
		if err != nil {
			return out, fmt.Errorf("load user groups: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var gid int
			if err := rows.Scan(&gid); err != nil {
				return out, err
			}
			selectedGroups[gid] = struct{}{}
		}
		if err := rows.Err(); err != nil {
			return out, err
		}
	}

	groupRows, err := ListAllGroupRows(ctx)
	if err != nil {
		return out, err
	}
	categories := loadModuleCategories(ctx)
	for _, row := range groupRows {
		id := intFromAny(row["id"])
		if 0 >= id {
			continue
		}
		_, isType := typeIDSet[id]
		_, sel := selectedGroups[id]
		categoryID := intFromAny(row["category_id"])
		cat := categories[categoryID]
		out.Groups = append(out.Groups, UserSecurityGroupRow{
			ID:           id,
			Name:         AsString(row["name"]),
			Selected:     sel,
			IsUserType:   isType,
			CategoryID:   categoryID,
			CategoryName: cat.name,
			Exclusive:    cat.exclusive,
		})
	}

	selectedCompanies := map[int]struct{}{}
	if targetUserID > 0 {
		cids, err := UserCompanyIDsForUser(ctx, targetUserID)
		if err != nil {
			return out, err
		}
		for _, cid := range cids {
			selectedCompanies[cid] = struct{}{}
		}
	}

	companyRows, err := Search(ctx, "core.company", nil)
	if err != nil {
		return out, err
	}
	for _, row := range companyRows {
		id := intFromAny(row["id"])
		if 0 >= id {
			continue
		}
		_, sel := selectedCompanies[id]
		out.Companies = append(out.Companies, UserSecurityCompanyRow{
			ID:       id,
			Name:     AsString(row["name"]),
			Selected: sel,
		})
	}
	return out, nil
}

type moduleCategory struct {
	name      string
	exclusive bool
}

func loadModuleCategories(ctx context.Context) map[int]moduleCategory {
	out := map[int]moduleCategory{}
	if DB == nil {
		return out
	}
	rows, err := DB.QueryContext(ctx, `SELECT id, name, COALESCE(exclusive, false) FROM `+MustQuotedTableName("sys.module.category"))
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		var exclusive bool
		if err := rows.Scan(&id, &name, &exclusive); err != nil {
			continue
		}
		out[id] = moduleCategory{name: name, exclusive: exclusive}
	}
	return out
}

func intFromAny(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}
