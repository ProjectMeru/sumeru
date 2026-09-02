package orm

import (
	"sumeru/core/modelmeta"
)

type SysModelData struct {
	modelmeta.ModelMeta `sumeru:"model=sys.model.data"`

	Module   modelmeta.String  `sumeru:"required"`
	Name     modelmeta.String  `sumeru:"required,unique"`
	ResModel modelmeta.String  `sumeru:"required,column=model"`
	CoreID   modelmeta.Integer `sumeru:"required"`
}
