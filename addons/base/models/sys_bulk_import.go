package models

import (
	"sumeru/core/sdk"
)

type SysBulkImport struct {
	sdk.Model `sumeru:"model=sys.bulk.import"`

	Name           sdk.String                  `sumeru:"required,string=Name"`
	TargetModel    sdk.String                  `sumeru:"required,string=Target Model"`
	ImportMode     sdk.String                  `sumeru:"string=Import Mode"`
	AttachmentID   sdk.Many2One[SysAttachment] `sumeru:"string=Staged File"`
	SelectedFields sdk.Text                    `sumeru:"string=Selected Fields"`
	CsvHeaders     sdk.Text                    `sumeru:"string=CSV Headers"`
	ColumnMapping  sdk.Text                    `sumeru:"string=Column Mapping"`
	PreviewJson    sdk.Text                    `sumeru:"string=Preview"`
	NextUrl        sdk.String                  `sumeru:"string=Return URL"`
	UserID         sdk.Many2One[CoreUser]      `sumeru:"string=User"`
	ActionID       sdk.Integer                 `sumeru:"string=Source Action"`
	State          sdk.String                  `sumeru:"string=State"`
}
