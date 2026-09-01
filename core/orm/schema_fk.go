package orm

import (
	"context"
	"fmt"
	"strings"

	applog "sumeru/core/applog"
)

// ensureForeignKeys adds missing FK constraints for many2one columns (best-effort on existing DBs).
func ensureForeignKeys(ctx context.Context, tbl schemaTable) error {
	for _, field := range tbl.Model.Fields() {
		if field.Type != Many2One || field.Relation == "" || field.Name == "id" {
			continue
		}
		targetTable, err := ModelToTableName(field.Relation)
		if err != nil {
			continue
		}
		colQuoted, err := QuotedColumnForModel(tbl.ModelName, field.Name)
		if err != nil {
			return err
		}
		constraintName := fkConstraintName(tbl.TableName, field.Name)
		if !pgIdentOK(constraintName) {
			continue
		}
		onDelete := normalizeOnDelete(field.OnDelete)
		targetQuoted, err := QuotedTableName(field.Relation)
		if err != nil {
			continue
		}
		q := fmt.Sprintf(
			`ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (id) ON DELETE %s NOT VALID`,
			tbl.QuotedTable, quoteIdent(constraintName), colQuoted, targetQuoted, onDelete,
		)
		if _, err := DB.Exec(q); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "already exists") {
				continue
			}
			applog.L(ctx).Warn("schema_sync_fk_skip", "table", tbl.TableName, "field", field.Name, "error", err.Error())
			continue
		}
		applog.L(ctx).Info("schema_sync_fk", "table", tbl.TableName, "field", field.Name, "target", targetTable)
	}
	return nil
}

func fkConstraintName(table, field string) string {
	name := table + "_" + field + "_fkey"
	if len(name) > 63 {
		name = name[:63]
	}
	return name
}

func normalizeOnDelete(policy string) string {
	switch strings.ToLower(strings.TrimSpace(policy)) {
	case "cascade":
		return "CASCADE"
	case "set null", "setnull", "set_null":
		return "SET NULL"
	case "restrict":
		return "RESTRICT"
	default:
		return "SET NULL"
	}
}
