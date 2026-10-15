package models

import (
	"sumeru/core/sdk"
)

type SysImportTemplate struct {
	sdk.Model `sumeru:"model=sys.import.template"`

	Name           sdk.String                `sumeru:"required,string=Name"`
	TargetModel    sdk.String                `sumeru:"required,index,string=Target Model"`
	ImportMode     sdk.String                `sumeru:"string=Import Mode"`
	SelectedFields sdk.Text                  `sumeru:"string=Selected Fields"`
	ColumnMapping  sdk.Text                  `sumeru:"string=Column Mapping"`
	Shared         sdk.Boolean               `sumeru:"string=Shared"`
	UserID         sdk.Many2One[CoreUser]    `sumeru:"string=User"`
	CompanyID      sdk.Many2One[CoreCompany] `sumeru:"string=Company"`
}
