package orm

import (
	"sumeru/core/modelmeta"
)

type SysActionWindow struct {
	modelmeta.ModelMeta `sumeru:"model=sys.action.window"`

	Name         modelmeta.String `sumeru:"required"`
	CoreModel    modelmeta.String `sumeru:"required"`
	ViewMode     modelmeta.String
	Domain       modelmeta.String
	Context      modelmeta.String
	Help         modelmeta.String
	Target       modelmeta.String  `sumeru:"string=Target,default=current"`
	ResID        modelmeta.Integer `sumeru:"string=Record"`
	Limit        modelmeta.Integer `sumeru:"string=Limit"`
	ViewID       modelmeta.String  `sumeru:"string=View"`
	SearchViewID modelmeta.String  `sumeru:"string=Search View"`
	Groups       modelmeta.String  `sumeru:"string=Groups"`
}
