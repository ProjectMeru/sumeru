package orm

import (
	"sumeru/core/modelmeta"
)

type SysModuleCategory struct {
	modelmeta.ModelMeta `sumeru:"model=sys.module.category"`

	Name      modelmeta.String  `sumeru:"required,unique,string=Name"`
	Sequence  modelmeta.Integer `sumeru:"string=Sequence"`
	Exclusive modelmeta.Boolean `sumeru:"string=Exclusive,default=false"`
}
