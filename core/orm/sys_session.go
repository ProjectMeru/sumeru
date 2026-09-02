package orm

import (
	"sumeru/core/modelmeta"
)

type SysSession struct {
	modelmeta.ModelMeta `sumeru:"model=sys.session"`

	Sid       modelmeta.String             `sumeru:"required,unique,index"`
	UserID    modelmeta.Many2One[CoreUser] `sumeru:"required,index,string=User"`
	ExpiresAt modelmeta.DateTime           `sumeru:"required"`
}
