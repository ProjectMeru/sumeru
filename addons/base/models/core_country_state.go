package models

import (
	"sumeru/core/sdk"
)

type CoreCountryState struct {
	sdk.Model `sumeru:"model=core.country.state"`

	Name      sdk.String                `sumeru:"required,string=State"`
	Code      sdk.String                `sumeru:"string=Code"`
	CountryID sdk.Many2One[CoreCountry] `sumeru:"required,index,string=Country"`
	Active    sdk.Boolean               `sumeru:"string=Active,default=true"`
}
