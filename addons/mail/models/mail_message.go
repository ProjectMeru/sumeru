package models

import (
	"sumeru/core/sdk"
)

type MailMessage struct {
	sdk.Model `sumeru:"model=mail.message"`

	ResModel   sdk.String  `sumeru:"required,index,column=model"`
	CoreID     sdk.Integer `sumeru:"required"`
	Body       sdk.Text    `sumeru:"required"`
	Subtype    sdk.String  `sumeru:"required"`
	Author     sdk.String
	CreateDate sdk.DateTime              `sumeru:"required"`
	CompanyID  sdk.Many2One[CoreCompany] `sumeru:"string=Company"`
}
