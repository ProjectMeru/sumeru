package models

import (
	"sumeru/core/sdk"
)

type SysServerAction struct {
	sdk.Model `sumeru:"model=sys.server.action"`

	Name          sdk.String  `sumeru:"required,string=Name"`
	ResModel      sdk.String  `sumeru:"index,column=model,string=Model"`
	EventName     sdk.String  `sumeru:"index,string=On Event"`
	TriggerOn     sdk.String  `sumeru:"string=Trigger On,default=event"`
	TriggerDomain sdk.Text    `sumeru:"string=Trigger Domain"`
	Code          sdk.Text    `sumeru:"string=Code / Notes"`
	Active        sdk.Boolean `sumeru:"string=Active,default=true"`
}
