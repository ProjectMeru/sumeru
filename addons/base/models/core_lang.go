package models

import (
	"sumeru/core/sdk"
)

type CoreLang struct {
	sdk.Model `sumeru:"model=core.lang"`

	Code         sdk.String  `sumeru:"required,unique,index,string=Code"`
	Name         sdk.String  `sumeru:"required,string=Name"`
	Active       sdk.Boolean `sumeru:"string=Active,default=true"`
	IsoCode      sdk.String  `sumeru:"string=ISO Code"`
	DecimalPoint sdk.String  `sumeru:"string=Decimal Separator,default=."`
	ThousandsSep sdk.String  `sumeru:"string=Thousands Separator,default=,"`
	Grouping     sdk.String  `sumeru:"string=Grouping"`
}
