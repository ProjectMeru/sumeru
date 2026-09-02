package orm

import (
	"sumeru/core/modelmeta"
)

type SysField struct {
	modelmeta.ModelMeta `sumeru:"model=sys.field"`

	Name      modelmeta.String             `sumeru:"required"`
	ModelID   modelmeta.Many2One[SysModel] `sumeru:"required,string=Model"`
	CoreModel modelmeta.String
	FieldType modelmeta.String
	Relation  modelmeta.String
	Label     modelmeta.String
	Required  modelmeta.Boolean
	Readonly  modelmeta.Boolean
	Index     modelmeta.Boolean
	Manual    modelmeta.Boolean             `sumeru:"string=Manual field"`
}
