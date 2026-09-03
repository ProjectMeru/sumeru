package models

import (
	"sumeru/core/sdk"
)

type MailActivityPlan struct {
	sdk.Model `sumeru:"model=mail.activity.plan"`

	Name        sdk.String                             `sumeru:"required,string=Name"`
	ResModel    sdk.String                             `sumeru:"required,index,column=model,string=Model"`
	Active      sdk.Boolean                            `sumeru:"string=Active,default=true"`
	TemplateIDs sdk.One2Many[MailActivityPlanTemplate] `sumeru:"string=Activity Templates"`
}
