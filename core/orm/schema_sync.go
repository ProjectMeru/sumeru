package orm

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"sumeru/core/applog"
)

// Additive DDL from model FieldDefinition definitions (PostgreSQL information_schema).
//
// Compares registered models to live columns and creates missing columns/indexes.
//
// Called after SyncModels() on startup and on module install (-i) / update (-u).

// SyncRegistrySchema adds missing columns and indexes for every model in Registry.
func SyncRegistrySchema() error {
	if DB == nil {
		return nil
	}
	ctx := ContextWithBypass(context.Background(), true)
	installed, err := InstalledModuleNames(ctx)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		if len(installed) == 0 {
			if owner := DeclaringModule(name); owner != "" && !IsPlatformModule(owner) {
				continue
			}
		} else if !ShouldMaterializeModel(name, installed) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		m := Registry[name]
		if err := syncModelSchema(ctx, m); err != nil {
			return fmt.Errorf("schema sync %s: %w", name, err)
		}
	}
	return ensureExtraIndexes()
}

func syncModelSchema(ctx context.Context, model Model) error {
	modelName := model.ModelName()
	tableName, err := ModelToTableName(modelName)
	if err != nil {
		return err
	}
	quotedTable, err := QuotedTableName(modelName)
	if err != nil {
		return err
	}
	exists, err := tableExists(tableName)
	if err != nil {
		return err
	}
	if !exists {
		return createTable(model)
	}
	existing, err := loadTableColumns(tableName)
	if err != nil {
		return err
	}
	for _, field := range model.Fields() {
		if field.Name == "id" || IsVirtualField(field) {
			continue
		}
		if _, ok := existing[strings.ToLower(field.Name)]; ok {
			continue
		}
		baseType, ok := ColumnTypeSQL(field)
		if !ok {
			continue
		}
		colDef := FormatAddColumnDefinition(field, baseType)
		colQuoted, err := QuotedColumnForModel(modelName, field.Name)
		if err != nil {
			return err
		}
		q := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", quotedTable, colQuoted, colDef)
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("%s: %w", q, err)
		}
		applog.L(ctx).Info("schema_sync", "table", tableName, "field", field.Name)
	}
	if err := dropStaleColumnUniques(modelName, tableName, quotedTable, model); err != nil {
		return err
	}
	return ensureModelIndexes(modelName, tableName, quotedTable, model)
}

// dropStaleColumnUniques removes single-column UNIQUE constraints when the field
// definition no longer sets Unique (e.g. sys.menu.name after menu label collisions).
func dropStaleColumnUniques(modelName, tableName, quotedTable string, model Model) error {
	for _, field := range model.Fields() {
		if field.Unique || field.Name == "id" {
			continue
		}
		baseType, ok := ColumnTypeSQL(field)
		if !ok || baseType == "" {
			continue
		}
		rows, err := DB.Query(`
			SELECT c.conname
			FROM pg_constraint c
			JOIN pg_class t ON c.conrelid = t.oid
			JOIN pg_namespace n ON n.oid = t.relnamespace
			WHERE n.nspname = 'public'
			  AND t.relname = $1
			  AND c.contype = 'u'
			  AND array_length(c.conkey, 1) = 1
			  AND EXISTS (
			    SELECT 1 FROM pg_attribute a
			    WHERE a.attrelid = t.oid
			      AND a.attnum = c.conkey[1]
			      AND NOT a.attisdropped
			      AND a.attname = $2
			  )
		`, tableName, field.Name)
		if err != nil {
			return err
		}
		var names []string
		for rows.Next() {
			var con string
			if err := rows.Scan(&con); err != nil {
				rows.Close()
				return err
			}
			names = append(names, con)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		for _, con := range names {
			if !pgIdentOK(con) {
				return fmt.Errorf("unsafe constraint name %q on %s", con, tableName)
			}
			q := fmt.Sprintf(`ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s`, quotedTable, quoteIdent(con))
			if _, err := DB.Exec(q); err != nil {
				return fmt.Errorf("drop unique %s.%s: %w", tableName, con, err)
			}
			applog.L(context.Background()).Info("schema_sync_drop_unique", "table", tableName, "constraint", con)
		}
	}
	return nil
}

func ensureModelIndexes(modelName, tableName, quotedTable string, model Model) error {
	for _, field := range model.Fields() {
		if IsVirtualField(field) {
			continue
		}
		if !(field.Index || field.Type == Many2One) {
			continue
		}
		colQuoted, err := QuotedColumnForModel(modelName, field.Name)
		if err != nil {
			return err
		}
		idxName := fmt.Sprintf("idx_%s_%s", tableName, field.Name)
		idxQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", quoteIdent(idxName), quotedTable, colQuoted)
		if _, err := DB.Exec(idxQuery + ";"); err != nil {
			return fmt.Errorf("index %s: %w", idxName, err)
		}
	}
	return nil
}

// ensureExtraIndexes creates composite indexes not expressible via single-field Index flags.
func ensureExtraIndexes() error {
	if DB == nil {
		return nil
	}
	if err := ensureMailMessageListIndex(); err != nil {
		return err
	}
	return ensureSysTranslationUniqueIndex()
}

func ensureMailMessageListIndex() error {
	tablePhysical := MustModelToTableName("mail.message")
	if tablePhysical == "" {
		return nil
	}
	ok, err := tableExists(tablePhysical)
	if err != nil || !ok {
		return err
	}
	tableQuoted := MustQuotedTableName("mail.message")
	modelCol, err := QuotedColumnForModel("mail.message", "model")
	if err != nil {
		return err
	}
	coreCol, err := QuotedColumnForModel("mail.message", "core_id")
	if err != nil {
		return err
	}
	dateCol, err := QuotedColumnForModel("mail.message", "create_date")
	if err != nil {
		return err
	}
	idxName := "idx_" + tablePhysical + "_model_core_created"
	q := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s, %s, %s DESC)",
		quoteIdent(idxName), tableQuoted, modelCol, coreCol, dateCol)
	_, err = DB.Exec(q)
	return err
}

func ensureSysTranslationUniqueIndex() error {
	tablePhysical := MustModelToTableName("sys.translation")
	if tablePhysical == "" {
		return nil
	}
	ok, err := tableExists(tablePhysical)
	if err != nil || !ok {
		return err
	}
	tableQuoted := MustQuotedTableName("sys.translation")
	langCol, err := QuotedColumnForModel("sys.translation", "lang")
	if err != nil {
		return err
	}
	srcCol, err := QuotedColumnForModel("sys.translation", "src")
	if err != nil {
		return err
	}
	moduleCol, err := QuotedColumnForModel("sys.translation", "module")
	if err != nil {
		return err
	}
	idxName := "sys_translation_lang_src_module_uidx"
	q := fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s, %s, %s)",
		quoteIdent(idxName), tableQuoted, langCol, srcCol, moduleCol)
	_, err = DB.Exec(q)
	return err
}

func pgIdentOK(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r == '_' {
			continue
		}
		if i > 0 && r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}

func loadTableColumns(tableName string) (map[string]struct{}, error) {
	rows, err := DB.Query(`
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
	`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]struct{})
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out[strings.ToLower(c)] = struct{}{}
	}
	return out, rows.Err()
}

// FormatAddColumnDefinition builds the SQL column definition fragment for ALTER TABLE ... ADD COLUMN.
func FormatAddColumnDefinition(f FieldDefinition, baseType string) string {
	if f.Type == Boolean {
		if f.DefaultVal == true {
			return baseType + " NOT NULL DEFAULT TRUE"
		}
		if f.DefaultVal == false {
			return baseType + " NOT NULL DEFAULT FALSE"
		}
		if lit, ok := sqlDefaultLiteral(f.DefaultVal); ok {
			return baseType + " DEFAULT " + lit
		}
		return baseType
	}
	if lit, ok := sqlDefaultLiteral(f.DefaultVal); ok {
		if f.Required {
			return baseType + " NOT NULL DEFAULT " + lit
		}
		return baseType + " DEFAULT " + lit
	}
	return baseType
}
