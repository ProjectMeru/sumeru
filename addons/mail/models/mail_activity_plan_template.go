package models

import (
	"sumeru/core/sdk"
)

type MailActivityPlanTemplate struct {
	sdk.Model `sumeru:"model=mail.activity.plan.template"`

	PlanID         sdk.Many2One[MailActivityPlan] `sumeru:"required,index,string=Plan"`
	ActivityTypeID sdk.Many2One[MailActivityType] `sumeru:"required,string=Activity Type"`
	Sequence       sdk.Integer                    `sumeru:"string=Sequence,default=10"`
	Summary        sdk.String                     `sumeru:"string=Summary"`
}
