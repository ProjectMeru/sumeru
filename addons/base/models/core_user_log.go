package models

import (
	"sumeru/core/sdk"
)

type CoreUserLog struct {
	sdk.Model `sumeru:"model=core.user.log"`

	UserID     sdk.Many2One[CoreUser] `sumeru:"index,string=User"`
	CreateDate sdk.DateTime           `sumeru:"required,string=When"`
	Ip         sdk.String             `sumeru:"string=IP"`
	Result     sdk.String             `sumeru:"required,string=Result,default=success"`
}
