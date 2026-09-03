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

// firstCompanyID returns the primary company row id, or 0 if none.
func firstCompanyID(ctx context.Context) int64 {
	settings, ok := firstCompanyMailSettings(ctx)
	if !ok {
		return 0
	}
	return settings.id
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
	tn := orm.MustQuotedTableName("mail.message")
	q := `SELECT body, subtype, author, create_date, model, core_id FROM ` + tn +
		` WHERE model = $1 AND core_id = $2 AND subtype = $3 ORDER BY create_date ASC, id ASC LIMIT $4`
	rows, err := orm.DB.QueryContext(ctx, q, model, coreID, SubtypeComment, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessageRows(rows)
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
	tn := orm.MustQuotedTableName("mail.message")
	ctxModel = strings.TrimSpace(ctxModel)
	var rows *sql.Rows
	var err error
	if ctxModel != "" && ctxID > 0 {
		if _, ok := orm.Registry[ctxModel]; !ok {
			ctxModel, ctxID = "", 0
		}
	}
	if ctxModel != "" && ctxID > 0 {
		q := `SELECT body, subtype, author, create_date, model, core_id FROM ` + tn +
			` WHERE subtype = 'notification' AND model = $1 AND core_id = $2` +
			` ORDER BY create_date DESC, id DESC LIMIT $3`
		rows, err = orm.DB.QueryContext(ctx, q, ctxModel, ctxID, limit)
	} else {
		q := `SELECT body, subtype, author, create_date, model, core_id FROM ` + tn +
			` WHERE subtype = 'notification'` +
			` ORDER BY create_date DESC, id DESC LIMIT $1`
		rows, err = orm.DB.QueryContext(ctx, q, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessageRows(rows)
}

func scanMessageRows(rows *sql.Rows) ([]Row, error) {
	var out []Row
	for rows.Next() {
		var r Row
		var ts time.Time
		if err := rows.Scan(&r.Body, &r.Subtype, &r.Author, &ts, &r.Model, &r.CoreID); err != nil {
			return out, err
		}
		r.CreateDate = ts.UTC()
		out = append(out, r)
	}
	return out, rows.Err()
}

// LogModuleEvent records a module lifecycle line in app.log (not mail.message).
func LogModuleEvent(ctx context.Context, moduleName, verb, detail string) {
	if err := orm.AppendAppLog(ctx, moduleName, verb, detail); err != nil {
		applog.L(ctx).Warn("applog.log_failed", "err", err)
	}
}
