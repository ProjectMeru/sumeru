package models

import (
	"sumeru/core/sdk"
)

type CoreUserAPIKey struct {
	sdk.Model `sumeru:"model=core.user.apikey"`

	UserID     sdk.Many2One[CoreUser] `sumeru:"required,index,string=User"`
	Name       sdk.String             `sumeru:"required,string=Name"`
	KeyPrefix  sdk.String             `sumeru:"required,string=Prefix"`
	KeyHash    sdk.String             `sumeru:"required,string=Hash"`
	Active     sdk.Boolean            `sumeru:"string=Active,default=true"`
	CreateDate sdk.DateTime           `sumeru:"required,string=Created"`
}
