package models

import (
	"sumeru/core/sdk"
)

type CoreCity struct {
	sdk.Model `sumeru:"model=core.city"`

	Name      sdk.String                     `sumeru:"required,string=City"`
	StateID   sdk.Many2One[CoreCountryState] `sumeru:"index,string=State"`
	CountryID sdk.Many2One[CoreCountry]      `sumeru:"required,index,string=Country"`
	Active    sdk.Boolean                    `sumeru:"string=Active,default=true"`
}
