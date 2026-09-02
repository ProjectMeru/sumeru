package models

import (
	"sumeru/core/sdk"
)

type MailActivityType struct {
	sdk.Model `sumeru:"model=mail.activity.type"`

	Name       sdk.String `sumeru:"required,string=Name"`
	Summary    sdk.String `sumeru:"string=Default Summary"`
	Category   sdk.String `sumeru:"string=Category"`
	DelayCount sdk.Integer `sumeru:"string=Schedule Delay,default=0"`
	DelayUnit  sdk.Selection[string] `sumeru:"string=Delay Unit,selection=days:Days,hours:Hours,weeks:Weeks,default=days"`
	Icon       sdk.String `sumeru:"string=Icon"`
	Active     sdk.Boolean `sumeru:"string=Active,default=true"`
}
