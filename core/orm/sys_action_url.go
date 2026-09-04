package orm

import (
	"sumeru/core/modelmeta"
)

type SysActionURL struct {
	modelmeta.ModelMeta `sumeru:"model=sys.action.url"`

	Name modelmeta.String `sumeru:"required,unique"`
	URL  modelmeta.String `sumeru:"required"`
}
