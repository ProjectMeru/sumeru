package models

import (
	"sumeru/core/sdk"
)

type SysAuditCleanup struct {
	sdk.Model `sumeru:"model=sys.audit.cleanup"`

	RetentionPreset sdk.String  `sumeru:"required,string=Retention,default=30_days,selection=7_days:Older than 7 days,30_days:Older than 30 days,custom:Custom days"`
	RetentionDays   sdk.Integer `sumeru:"string=Custom days"`
	FilterModel     sdk.String  `sumeru:"string=Limit to model"`
	DryRun          sdk.Boolean `sumeru:"string=Dry run only"`
	ResultSummary   sdk.Text    `sumeru:"string=Result"`
}
