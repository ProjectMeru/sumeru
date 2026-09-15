package models

import (
	"sumeru/core/sdk"
)

type SysReportAction struct {
	sdk.Model `sumeru:"model=sys.report.action"`

	Name         sdk.String  `sumeru:"required,string=Name"`
	ResModel     sdk.String  `sumeru:"required,index,column=model,string=Model"`
	ReportType   sdk.String  `sumeru:"string=Report Type,default=pdf"`
	Arch         sdk.Text    `sumeru:"string=Document Arch"`
	TemplatePath sdk.String  `sumeru:"string=Template Path"`
	Paperformat  sdk.String  `sumeru:"string=Paper Format,default=a4"`
	Active       sdk.Boolean `sumeru:"string=Active,default=true"`
}
