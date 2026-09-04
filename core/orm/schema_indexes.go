package orm

import (
	"context"
	"fmt"
)

// ensureExtraIndexes creates composite indexes not expressible via single-field Index flags.
func ensureExtraIndexes(ctx context.Context) error {
	if DB == nil {
		return nil
	}
	if err := ensureMailMessageListIndex(ctx); err != nil {
		return err
	}
	return ensureSysTranslationUniqueIndex(ctx)
}

func ensureMailMessageListIndex(ctx context.Context) error {
	tablePhysical := MustModelToTableName("mail.message")
	if tablePhysical == "" {
		return nil
	}
	ok, err := tableExists(ctx, tablePhysical)
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
	_, err = DB.ExecContext(ctx, q)
	return err
}

func ensureSysTranslationUniqueIndex(ctx context.Context) error {
	tablePhysical := MustModelToTableName("sys.translation")
	if tablePhysical == "" {
		return nil
	}
	ok, err := tableExists(ctx, tablePhysical)
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
	_, err = DB.ExecContext(ctx, q)
	return err
}
