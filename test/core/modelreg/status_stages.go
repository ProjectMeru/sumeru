package modelreg

import "sumeru/core/modelmeta"

// StatusStage constants are the stages for the status tag registration test.
type StatusStage string

const (
	StatusStageDraft  StatusStage = "draft"
	StatusStageActive StatusStage = "active"
)

type StatusBarModel struct {
	modelmeta.ModelMeta `sumeru:"model=test.status.bar"`
	State               modelmeta.Selection[StatusStage] `sumeru:"status,required,string=Status,default=draft"`
}

type BadStatusModel struct {
	modelmeta.ModelMeta `sumeru:"model=test.status.bad"`
	Name                modelmeta.String `sumeru:"status"`
}
