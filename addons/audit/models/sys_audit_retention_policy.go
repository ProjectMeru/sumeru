package models

import (
	"sumeru/core/sdk"
)

type SysAuditRetentionPolicy struct {
	sdk.Model `sumeru:"model=sys.audit.retention.policy"`

	Name              sdk.String   `sumeru:"required,string=Name,default=Audit retention"`
	HotRetentionDays  sdk.Integer  `sumeru:"required,string=Hot retention days,default=90"`
	FilterModel       sdk.String   `sumeru:"string=Limit to model"`
	AutoEnabled       sdk.Boolean  `sumeru:"string=Automatic retention,default=false"`
	BatchSize         sdk.Integer  `sumeru:"string=Batch size,default=500"`
	MaxRowsPerRun     sdk.Integer  `sumeru:"string=Max rows per run,default=5000"`
	Compress          sdk.Boolean  `sumeru:"string=Gzip exports,default=true"`
	DeleteAfterExport sdk.Boolean  `sumeru:"string=Delete from database after export,default=false"`
	LegalHold         sdk.Boolean  `sumeru:"string=Legal hold no deletes,default=false"`
	DryRunNext        sdk.Boolean  `sumeru:"string=Dry run on next manual run,default=false"`
	LastRunAt         sdk.DateTime `sumeru:"string=Last run"`
	LastRunSummary    sdk.Text     `sumeru:"string=Last run summary"`
}
