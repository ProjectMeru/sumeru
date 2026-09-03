package models

import (
	"sumeru/core/sdk"
)

type SysWorkflowTransition struct {
	sdk.Model `sumeru:"model=sys.workflow.transition"`

	Name      sdk.String              `sumeru:"required,string=Name"`
	ResModel  sdk.String              `sumeru:"required,index,column=model,string=Model"`
	FromState sdk.String              `sumeru:"string=From"`
	ToState   sdk.String              `sumeru:"required,string=To"`
	GroupID   sdk.Many2One[CoreGroup] `sumeru:"string=Required Group"`
	Active    sdk.Boolean             `sumeru:"string=Active,default=true"`
}
