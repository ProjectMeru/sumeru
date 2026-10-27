package orm

import (
	"sumeru/core/modelmeta"
)

type SysRule struct {
	modelmeta.ModelMeta `sumeru:"model=sys.rule"`

	Name        modelmeta.String `sumeru:"required,unique"`
	ResModel    modelmeta.String `sumeru:"required,column=model"`
	DomainForce modelmeta.Text
	Active      modelmeta.Boolean
	PermRead    modelmeta.Boolean
	PermWrite   modelmeta.Boolean
	PermCreate  modelmeta.Boolean
	PermUnlink  modelmeta.Boolean
	GroupIDs    modelmeta.Many2Many[CoreGroup] `sumeru:"string=Groups,table=sys_rule_group_rel,left=rule_id,right=group_id"`
}
