package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/errcode"
	"sumeru/core/orm"
)

// SwitchCompanyPost sets the signed-in user's company_id (must exist in core.company).
func SwitchCompanyPost(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}

	form := parseCompanySwitchForm(r)
	userID := SessionUserID(r)
	if userID <= 0 {
		http.Redirect(w, r, loginRoute, http.StatusSeeOther)
		return
	}

	switchActiveCompany(r.Context(), userID, form.CompanyID)
	redirectToWebNext(w, r, form.Next)
}

type companySwitchForm struct {
	CompanyID int
	Next      string
}

func parseCompanySwitchForm(r *http.Request) companySwitchForm {
	companyID, _ := strconv.Atoi(strings.TrimSpace(r.PostFormValue(companyIDFormField)))
	return companySwitchForm{
		CompanyID: companyID,
		Next:      r.PostFormValue(nextField),
	}
}

func switchActiveCompany(ctx context.Context, userID, companyID int) {
	if companyID <= 0 || !companyRecordExists(ctx, companyID) {
		auditCompanySwitchDeny(ctx, userID, companyID, "invalid company")
		return
	}
	if !orm.UserAllowedCompany(ctx, userID, int64(companyID)) {
		auditCompanySwitchDeny(ctx, userID, companyID, "not allowed")
		return
	}
	previous := int(orm.ActiveCompanyIDForUser(ctx, userID))
	if err := updateUserActiveCompany(ctx, userID, companyID); err != nil {
		WebLogEvent(ctx, WebLogInput{
			Route:     companySwitchRoute,
			Message:   "Could not update active company",
			Code:      errcode.InternalError,
			Operation: "company_switch",
			Status:    logStatusFailure,
			Err:       err,
			ContextFields: map[string]interface{}{
				"company_id": companyID,
				"user_id":    userID,
			},
		})
		return
	}
	detail := fmt.Sprintf("user=%d company=%d", userID, companyID)
	if previous > 0 && previous != companyID {
		detail = fmt.Sprintf("user=%d %d->%d", userID, previous, companyID)
	}
	orm.AppendAudit(ctx, "company_switch", "sys.security", 0, nil, nil, detail)
	WebLogNavigation(ctx, companySwitchRoute, "company_switch", "Active company switched", map[string]interface{}{
		"company_id": companyID,
		"user_id":    userID,
	})
}

func auditCompanySwitchDeny(ctx context.Context, userID, companyID int, reason string) {
	detail := fmt.Sprintf("user=%d company=%d: %s", userID, companyID, reason)
	orm.AppendAudit(ctx, "company_switch_deny", "sys.security", 0, nil, nil, detail)
}

func companyRecordExists(ctx context.Context, companyID int) bool {
	_, err := orm.SearchOne(ctx, coreCompanyModel, map[string]interface{}{"id": companyID})
	return err == nil
}

func updateUserActiveCompany(ctx context.Context, userID, companyID int) error {
	userTable := orm.MustQuotedTableName(coreUserModel)
	_, err := orm.DB.ExecContext(ctx,
		`UPDATE `+userTable+` SET company_id = $1 WHERE id = $2`,
		companyID, userID,
	)
	return err
}

func redirectToWebNext(w http.ResponseWriter, r *http.Request, rawNext string) {
	http.Redirect(w, r, SafeWebNext(rawNext, homeRoute), http.StatusSeeOther)
}
