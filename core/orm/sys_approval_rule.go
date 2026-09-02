package orm

import (
	"sumeru/core/modelmeta"
)

type SysApprovalRule struct {
	modelmeta.ModelMeta `sumeru:"model=sys.approval.rule"`

	ResModel        modelmeta.String              `sumeru:"required,column=model"`
	GroupID         modelmeta.Many2One[CoreGroup] `sumeru:"required,string=Group"`
	FromState       modelmeta.String
	ToState         modelmeta.String `sumeru:"required"`
	RequireApproval modelmeta.Boolean
}
