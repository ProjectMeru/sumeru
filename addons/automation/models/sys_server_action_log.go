package models

import "sumeru/core/sdk"

type SysServerActionLog struct {
	sdk.Model `sumeru:"model=sys.server.action.log"`

	ActionID  sdk.Many2One[SysServerAction] `sumeru:"required,index,string=Action"`
	EventName sdk.String                    `sumeru:"required,index,string=Event"`
	ResModel  sdk.String                    `sumeru:"index,column=model,string=Model"`
	ResID     sdk.Integer                   `sumeru:"index,string=Record"`
	RanAt     sdk.DateTime                  `sumeru:"required,string=Ran At"`
	DedupeKey sdk.String                    `sumeru:"required,unique,string=Dedupe Key"`
	Detail    sdk.Text                      `sumeru:"string=Detail"`
}
