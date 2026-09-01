package models

import (
	"sumeru/core/sdk"
)

type SysActivity struct {
	sdk.Model `sumeru:"model=sys.activity"`

	Name       sdk.String            `sumeru:"required,string=Summary"`
	ResModel   sdk.String            `sumeru:"index,column=model,string=Model"`
	ResID      sdk.Integer           `sumeru:"index,string=Record"`
	ActivityType sdk.String          `sumeru:"string=Type,default=todo"`
	State      sdk.String            `sumeru:"string=State,default=open"`
	DueDate    sdk.Date              `sumeru:"string=Due Date"`
	UserID     sdk.Many2One[CoreUser] `sumeru:"string=Assigned To"`
	Note       sdk.Text              `sumeru:"string=Note"`
}
