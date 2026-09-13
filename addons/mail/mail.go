package mail

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"sumeru/core/applog"
	"sumeru/core/orm"
)

// Subtype values for mail.message rows.
const (
	SubtypeComment      = "comment"
	SubtypeNotification = "notification"
	SubtypeModule       = "module"
)

type Row struct {
	Body       string
	Subtype    string
	Author     string
	CreateDate time.Time
	Model      string
	CoreID     int64
}

type companyMailSettings struct {
	id                   int64
	chatterEnabled       bool
	activityPanelEnabled bool
}

func firstCompanyMailSettings(ctx context.Context) (companyMailSettings, bool) {
	if orm.DB == nil {
		return companyMailSettings{chatterEnabled: true, activityPanelEnabled: true}, false
	}
	tn := orm.MustQuotedTableName("core.company")
	var id sql.NullInt64
	var chatter, activity sql.NullBool
	err := orm.DB.QueryRowContext(ctx,
		`SELECT id, mail_chatter_enabled, mail_activity_panel_enabled FROM `+tn+` ORDER BY id ASC LIMIT 1`,
	).Scan(&id, &chatter, &activity)
	if err != nil {
		return companyMailSettings{chatterEnabled: true, activityPanelEnabled: true}, false
	}
	out := companyMailSettings{chatterEnabled: true, activityPanelEnabled: true}
	if id.Valid {
		out.id = id.Int64
	}
	if chatter.Valid {
		out.chatterEnabled = chatter.Bool
	}
	if activity.Valid {
		out.activityPanelEnabled = activity.Bool
	}
	return out, id.Valid
}

// CompanyChatterEnabled reads mail_chatter_enabled from the first core.company row (default true).
func CompanyChatterEnabled(ctx context.Context) bool {
	settings, _ := firstCompanyMailSettings(ctx)
	return settings.chatterEnabled
}

// CompanyActivityPanelEnabled reads mail_activity_panel_enabled from the first core.company row (default true).
func CompanyActivityPanelEnabled(ctx context.Context) bool {
	settings, _ := firstCompanyMailSettings(ctx)
	return settings.activityPanelEnabled
}

// PostMessage inserts a mail.message row. author may be empty (stored as "System").
func PostMessage(ctx context.Context, model string, coreID int64, body, subtype, author string) error {
	if orm.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	model = strings.TrimSpace(model)
	body = strings.TrimSpace(body)
	subtype = strings.TrimSpace(subtype)
	if model == "" || body == "" || subtype == "" {
		return fmt.Errorf("model, body, and subtype are required")
	}
	if _, ok := orm.Registry[model]; !ok {
		return fmt.Errorf("unknown model %q", model)
	}
	uid := orm.SecurityUID(ctx)
	if err := orm.CheckModelAccess(ctx, uid, model, "write"); err != nil {
		return err
	}
	if _, err := orm.SearchOne(ctx, model, map[string]interface{}{"id": int(coreID)}); err != nil {
		return fmt.Errorf("record not found or access denied")
	}
	inst, ok := orm.Registry["mail.message"]
	if !ok {
		return fmt.Errorf("unknown model %q", "mail.message")
	}
	author = strings.TrimSpace(author)
	if author == "" {
		author = "System"
	}
	vals := map[string]interface{}{
		"model":       model,
		"core_id":     int(coreID),
		"body":        body,
		"subtype":     subtype,
		"author":      author,
		"create_date": time.Now().UTC(),
	}
	if settings, ok := firstCompanyMailSettings(ctx); ok && settings.id > 0 {
		vals["company_id"] = int(settings.id)
	}
	_, err := orm.Create(ctx, inst, vals)
	return err
}

// ListCommentsForRecord returns user chatter lines (subtype comment) for a record, oldest first.
func ListCommentsForRecord(ctx context.Context, model string, coreID int64, limit int) ([]Row, error) {
	if orm.DB == nil {
		return nil, nil
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if _, ok := orm.Registry[model]; !ok {
		return nil, fmt.Errorf("unknown model %q", model)
	}
	if err := orm.CheckModelAccess(ctx, orm.SecurityUID(ctx), model, "read"); err != nil {
		return nil, err
	}
	if _, err := orm.SearchOne(ctx, model, map[string]interface{}{"id": int(coreID)}); err != nil {
		return nil, fmt.Errorf("record not found or access denied")
	}
	if limit <= 0 || limit > 500 {
		limit = 120
	}
	domain := [][]interface{}{
		{"model", "=", model},
		{"core_id", "=", int(coreID)},
		{"subtype", "=", SubtypeComment},
	}
	records, err := orm.SearchPage(ctx, "mail.message", domain, limit, 0, "create_date ASC")
	if err != nil {
		return nil, err
	}
	return rowsFromSearchResults(records), nil
}

// QueryActivityLog returns audit-oriented lines (module events, notifications, record saves).
// User chatter comments are excluded. Optional ctxModel/ctxID adds notifications on that record only.
func QueryActivityLog(ctx context.Context, limit int, ctxModel string, ctxID int64) ([]Row, error) {
	if orm.DB == nil {
		return nil, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 40
	}
	uid := orm.SecurityUID(ctx)
	if err := orm.CheckModelAccess(ctx, uid, "mail.message", "read"); err != nil {
		return nil, err
	}
	ctxModel = strings.TrimSpace(ctxModel)
	if ctxModel != "" && ctxID > 0 {
		if _, ok := orm.Registry[ctxModel]; !ok {
			ctxModel, ctxID = "", 0
		}
	}
	domain := [][]interface{}{{"subtype", "=", SubtypeNotification}}
	if ctxModel != "" && ctxID > 0 {
		if err := orm.CheckModelAccess(ctx, uid, ctxModel, "read"); err != nil {
			return nil, err
		}
		if _, err := orm.SearchOne(ctx, ctxModel, map[string]interface{}{"id": int(ctxID)}); err != nil {
			return nil, fmt.Errorf("record not found or access denied")
		}
		domain = append(domain,
			[]interface{}{"model", "=", ctxModel},
			[]interface{}{"core_id", "=", int(ctxID)},
		)
	}
	records, err := orm.SearchPage(ctx, "mail.message", domain, limit, 0, "create_date DESC")
	if err != nil {
		return nil, err
	}
	return rowsFromSearchResults(records), nil
}

func rowsFromSearchResults(records []map[string]interface{}) []Row {
	out := make([]Row, 0, len(records))
	for _, rec := range records {
		r := Row{
			Body:    rowString(rec, "body"),
			Subtype: rowString(rec, "subtype"),
			Author:  rowString(rec, "author"),
			Model:   rowString(rec, "model"),
			CoreID:  rowInt64(rec, "core_id"),
		}
		switch v := rec["create_date"].(type) {
		case time.Time:
			r.CreateDate = v.UTC()
		case *time.Time:
			if v != nil {
				r.CreateDate = v.UTC()
			}
		}
		out = append(out, r)
	}
	return out
}

func rowString(rec map[string]interface{}, key string) string {
	if rec == nil {
		return ""
	}
	switch v := rec[key].(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}

func rowInt64(rec map[string]interface{}, key string) int64 {
	if rec == nil {
		return 0
	}
	switch v := rec[key].(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		return 0
	}
}

// LogModuleEvent records a module lifecycle line in app.log (not mail.message).
func LogModuleEvent(ctx context.Context, moduleName, verb, detail string) {
	if err := orm.AppendAppLog(ctx, moduleName, verb, detail); err != nil {
		applog.L(ctx).Warn("applog.log_failed", "err", err)
	}
}
