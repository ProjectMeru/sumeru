package orm

import (
	"context"
	"database/sql"
	"fmt"
)

func queryUserCompanyID(ctx context.Context, uid int) (sql.NullInt64, error) {
	var companyID sql.NullInt64
	err := DB.QueryRowContext(ctx,
		`SELECT company_id FROM `+MustQuotedTableName("core.user")+` WHERE id = $1`, uid,
	).Scan(&companyID)
	return companyID, err
}

// UserCompanyIDs returns allowed company ids for uid (M2M + current company_id).
func UserCompanyIDs(ctx context.Context, uid int) ([]int64, error) {
	if uid <= 0 || DB == nil {
		return nil, nil
	}
	seen := map[int64]struct{}{}
	var out []int64
	add := func(id int64) {
		if id <= 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	companyID, _ := queryUserCompanyID(ctx, uid)
	if companyID.Valid {
		add(companyID.Int64)
	}
	rel := MustQuotedTableName("core.user.company.rel")
	rows, err := DB.QueryContext(ctx, `SELECT company_id FROM `+rel+` WHERE user_id = $1`, uid)
	if err != nil {
		// Join table may not exist yet during early bootstrap.
		return out, nil
	}
	defer rows.Close()
	for rows.Next() {
		var cid int64
		if err := rows.Scan(&cid); err != nil {
			return out, err
		}
		add(cid)
	}
	return out, rows.Err()
}

// UserAllowedCompany returns true if cid is in the user's company set (or user has none configured → allow current only).
func UserAllowedCompany(ctx context.Context, uid int, cid int64) bool {
	if uid == superuserUID {
		return true
	}
	ids, err := UserCompanyIDs(ctx, uid)
	if err != nil || len(ids) == 0 {
		return true
	}
	for _, id := range ids {
		if id == cid {
			return true
		}
	}
	return false
}

// ActiveCompanyIDForUser reads core.user.company_id directly (no ORM Search/log path).
func ActiveCompanyIDForUser(ctx context.Context, uid int) int64 {
	if uid <= 0 || DB == nil {
		return 0
	}
	companyID, _ := queryUserCompanyID(ctx, uid)
	if !companyID.Valid {
		return 0
	}
	return companyID.Int64
}

// --- Company membership (M2M links) ---

// UserCompanyIDsForUser returns allowed company ids from core.user.company.rel.
func UserCompanyIDsForUser(ctx context.Context, userID int) ([]int, error) {
	if userID <= 0 || DB == nil {
		return nil, nil
	}
	table := MustQuotedTableName("core.user.company.rel")
	rows, err := DB.QueryContext(ctx, `SELECT company_id FROM `+table+` WHERE user_id = $1 ORDER BY company_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var companyID int
		if err := rows.Scan(&companyID); err != nil {
			return nil, err
		}
		if companyID > 0 {
			out = append(out, companyID)
		}
	}
	return out, rows.Err()
}

// SetUserCompanyLinks replaces allowed-company membership for a user.
func SetUserCompanyLinks(ctx context.Context, userID int, companyIDs []int) error {
	if DB == nil {
		return fmt.Errorf("no database")
	}
	if userID <= 0 {
		return fmt.Errorf("invalid user id")
	}
	table := MustQuotedTableName("core.user.company.rel")
	if _, err := DB.ExecContext(ctx, `DELETE FROM `+table+` WHERE user_id = $1`, userID); err != nil {
		return err
	}
	seen := map[int]struct{}{}
	for _, companyID := range companyIDs {
		if companyID <= 0 {
			continue
		}
		if _, ok := seen[companyID]; ok {
			continue
		}
		seen[companyID] = struct{}{}
		if _, err := DB.ExecContext(ctx,
			`INSERT INTO `+table+` (user_id, company_id) VALUES ($1, $2) ON CONFLICT (user_id, company_id) DO NOTHING`,
			userID, companyID); err != nil {
			return err
		}
	}
	return nil
}
