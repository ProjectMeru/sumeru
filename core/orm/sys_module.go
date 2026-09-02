package orm

import (
	"sumeru/core/modelmeta"
)

type SysModule struct {
	modelmeta.ModelMeta `sumeru:"model=sys.module"`

	Name        modelmeta.String `sumeru:"required,unique"`
	DisplayName modelmeta.String
	Author      modelmeta.String
	Version     modelmeta.String
	Description modelmeta.Text
	Icon        modelmeta.String
	CategoryID  modelmeta.Many2One[SysModuleCategory] `sumeru:"relation=sys.module.category"`
	State       modelmeta.String   `sumeru:"required"`
	Application modelmeta.Boolean `sumeru:"required"`
	Active      modelmeta.Boolean `sumeru:"required"`
	LastError   modelmeta.Text
}
