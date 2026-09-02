package orm

import (
	"sumeru/core/modelmeta"
)

type SysMenu struct {
	modelmeta.ModelMeta `sumeru:"model=sys.menu"`

	Name         modelmeta.String            `sumeru:"required"`
	ParentID     modelmeta.Many2One[SysMenu] `sumeru:"string=Parent Menu"`
	ActionID     modelmeta.Integer
	Action       modelmeta.String
	Sequence     modelmeta.Integer
	WebIcon      modelmeta.String
	Module       modelmeta.String `sumeru:"index"`
	AccessGroups modelmeta.String
}
