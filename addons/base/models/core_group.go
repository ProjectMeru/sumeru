package models

import (
	"sumeru/core/sdk"
)

type CoreGroup struct {
	sdk.Model `sumeru:"model=core.group"`

	Name       sdk.String                      `sumeru:"required,unique,string=Name"`
	CategoryID sdk.Many2One[SysModuleCategory] `sumeru:"index,string=Application"`
	Sequence   sdk.Integer                     `sumeru:"string=Sequence"`
}
