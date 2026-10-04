package models

import (
	"sumeru/core/sdk"
)

// CoreCurrency is the platform currency catalog. Rates are kept on
// core.currency.rate (see core_currency_rate.go).
type CoreCurrency struct {
	sdk.Model `sumeru:"model=core.currency"`

	Name   sdk.String  `sumeru:"required,unique,index,string=Currency"`
	Symbol sdk.String  `sumeru:"required,string=Symbol"`
	Active sdk.Boolean `sumeru:"string=Active,default=true"`
}
