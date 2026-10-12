package models

import "sumeru/core/sdk"

type MailActivity struct {
	sdk.Model `sumeru:"model=mail.activity"`

	ResModel       sdk.String                     `sumeru:"required,index,column=model,string=Document Model"`
	ResID          sdk.Integer                    `sumeru:"required,index,string=Document"`
	ActivityTypeID sdk.Many2One[MailActivityType] `sumeru:"required,index,string=Type"`
	UserID         sdk.Many2One[CoreUser]         `sumeru:"index,string=Assigned To"`
	DateDeadline   sdk.Date                       `sumeru:"index,string=Due Date"`
	Summary        sdk.String                     `sumeru:"string=Summary"`
	Note           sdk.Text                       `sumeru:"string=Note"`
	State          sdk.String                     `sumeru:"index,string=State,default=planned,selection=planned:Planned,done:Done"`
	DateDone       sdk.Date                       `sumeru:"string=Done On"`
}
