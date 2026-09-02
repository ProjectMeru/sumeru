package orm

import (
	"sumeru/core/modelmeta"
)

type SysAccess struct {
	modelmeta.ModelMeta `sumeru:"model=sys.access"`

	Name       modelmeta.String              `sumeru:"required,unique"`
	ResModel   modelmeta.String              `sumeru:"required,column=model"`
	GroupID    modelmeta.Many2One[CoreGroup] `sumeru:"string=Group"`
	PermRead   modelmeta.Boolean
	PermWrite  modelmeta.Boolean
	PermCreate modelmeta.Boolean
	PermUnlink modelmeta.Boolean
}
