package models

import (
	"sumeru/core/sdk"
)

type MailActivity struct {
	sdk.Model `sumeru:"model=mail.activity"`

	Name         sdk.String             `sumeru:"required,string=Activity"`
	ResModel     sdk.String             `sumeru:"required,index,column=model,string=Model"`
	ResID        sdk.Integer            `sumeru:"required,index,string=Record"`
	UserID       sdk.Many2One[CoreUser] `sumeru:"string=Assigned To"`
	Summary      sdk.Text               `sumeru:"string=Summary"`
	DateDeadline sdk.Date               `sumeru:"string=Due Date"`
	State        sdk.String             `sumeru:"string=State,default=planned,selection=planned:Planned,done:Done,cancelled:Cancelled"`
}
