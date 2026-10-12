package models

import "sumeru/core/sdk"

type MailActivityType struct {
	sdk.Model `sumeru:"model=mail.activity.type"`

	Name       sdk.String  `sumeru:"required,unique,string=Name"`
	DelayCount sdk.Integer `sumeru:"string=Delay (days),default=0"`
	Icon       sdk.String  `sumeru:"string=Icon"`
}
