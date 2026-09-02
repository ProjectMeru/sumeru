package models

import (
	"sumeru/core/sdk"
)

type MailTemplate struct {
	sdk.Model `sumeru:"model=mail.template"`

	Name     sdk.String  `sumeru:"required,string=Name"`
	ResModel sdk.String  `sumeru:"required,index,column=model,string=Applies to"`
	Subject  sdk.String  `sumeru:"string=Subject"`
	BodyHTML sdk.Text    `sumeru:"string=Body"`
	Active   sdk.Boolean `sumeru:"string=Active,default=true"`
}
