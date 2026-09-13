package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// NextSequence returns the next formatted number for code and advances the counter.
func NextSequence(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" || DB == nil {
		return "", fmt.Errorf("sequence code required")
	}
	bypass := AuditedBypass(ctx, "sequence.next")
	tbl := MustQuotedTableName("sys.sequence")
	tx, err := DB.BeginTx(bypass, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	var id, padding, numberNext int
	var prefix, suffix sql.NullString
	var active bool
	err = tx.QueryRowContext(bypass,
		`SELECT id, COALESCE(prefix,''), COALESCE(suffix,''), COALESCE(padding,5), number_next, active
		 FROM `+tbl+` WHERE code = $1 FOR UPDATE`, code,
	).Scan(&id, &prefix, &suffix, &padding, &numberNext, &active)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("unknown sequence %q", code)
	}
	if err != nil {
		return "", err
	}
	if !active {
		return "", fmt.Errorf("sequence %q is inactive", code)
	}
	if padding < 0 {
		padding = 0
	}
	if _, err := tx.ExecContext(bypass, `UPDATE `+tbl+` SET number_next = $1 WHERE id = $2`, numberNext+1, id); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	num := fmt.Sprintf("%0*d", padding, numberNext)
	return prefix.String + num + suffix.String, nil
}

// GetConfig returns sys.config.parameter value for key, or def if missing.
func GetConfig(ctx context.Context, key, def string) string {
	key = strings.TrimSpace(key)
	if key == "" || DB == nil {
		return def
	}
	var val sql.NullString
	err := DB.QueryRowContext(AuditedBypass(ctx, "config.read"),
		`SELECT value FROM `+MustQuotedTableName("sys.config.parameter")+` WHERE key = $1`, key,
	).Scan(&val)
	if err != nil || !val.Valid {
		return def
	}
	return val.String
}

// SetConfig upserts a config parameter.
func SetConfig(ctx context.Context, key, value string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("config key required")
	}
	bypass := AuditedBypass(ctx, "config.write")
	tbl := MustQuotedTableName("sys.config.parameter")
	res, err := DB.ExecContext(bypass, `UPDATE `+tbl+` SET value = $1 WHERE key = $2`, value, key)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return nil
	}
	inst, ok := Registry["sys.config.parameter"]
	if !ok {
		return fmt.Errorf("unknown model %q", "sys.config.parameter")
	}
	_, err = Create(bypass, inst, map[string]interface{}{"key": key, "value": value})
	return err
}
