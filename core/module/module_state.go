package module

import (
	"context"

	"sumeru/core/orm"
)

func setModuleState(ctx context.Context, moduleName, state string, active bool) error {
	_, err := orm.DB.ExecContext(ctx,
		`UPDATE `+orm.MustQuotedTableName("sys.module")+` SET state = $1, active = $2 WHERE name = $3`,
		state, active, moduleName,
	)
	return err
}

// setModuleStateOnly updates state without changing active (CLI -u must not force active=true).
func setModuleStateOnly(ctx context.Context, moduleName, state string) error {
	_, err := orm.DB.ExecContext(ctx,
		`UPDATE `+orm.MustQuotedTableName("sys.module")+` SET state = $1 WHERE name = $2`,
		state, moduleName,
	)
	return err
}

func setModuleLastError(ctx context.Context, moduleName, msg string) error {
	_, err := orm.DB.ExecContext(ctx,
		`UPDATE `+orm.MustQuotedTableName("sys.module")+` SET last_error = $1 WHERE name = $2`,
		msg, moduleName,
	)
	return err
}

func setModuleToRemove(ctx context.Context, moduleName string) error {
	_, err := orm.DB.ExecContext(ctx,
		`UPDATE `+orm.MustQuotedTableName("sys.module")+` SET state = 'to_remove' WHERE name = $1`,
		moduleName,
	)
	return err
}

func setModuleActiveOnly(ctx context.Context, moduleName string, active bool) error {
	_, err := orm.DB.ExecContext(ctx,
		`UPDATE `+orm.MustQuotedTableName("sys.module")+` SET active = $1 WHERE name = $2`,
		active, moduleName,
	)
	return err
}
