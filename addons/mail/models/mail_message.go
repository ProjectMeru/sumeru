package models

import "sumeru/core/sdk"

type MailMessage struct {
	sdk.Model `sumeru:"model=mail.message"`

	ResModel    sdk.String                `sumeru:"index,column=model,string=Document Model"`
	ResID       sdk.Integer               `sumeru:"index,string=Document"`
	MessageType sdk.String                `sumeru:"string=Type,default=email,selection=email:Email,comment:Comment"`
	Subject     sdk.String                `sumeru:"string=Subject"`
	Body        sdk.Text                  `sumeru:"string=Body"`
	EmailFrom   sdk.String                `sumeru:"string=From"`
	MessageID   sdk.String                `sumeru:"index,unique,column=message_id,string=Message-ID"`
	InReplyTo   sdk.String                `sumeru:"index,string=In-Reply-To"`
	AuthorID    sdk.Many2One[CorePartner] `sumeru:"string=Author"`
	State       sdk.String                `sumeru:"index,string=State,default=received,selection=received:Received,quarantine:Quarantine,rejected:Rejected"`
	CompanyID   sdk.Many2One[CoreCompany] `sumeru:"string=Company"`
}
