package models

import (
	"sumeru/core/sdk"
)

// PlatformPricingSample is a reference row for SUM-PLAT-18 monetary fields and conversion.
type PlatformPricingSample struct {
	sdk.Model `sumeru:"model=platform.pricing.sample"`

	Name       sdk.String               `sumeru:"required,string=Label"`
	CurrencyID sdk.Many2One[CoreCurrency] `sumeru:"required,index,string=Currency"`
	Amount     sdk.Money                `sumeru:"required,string=Amount,currency=currency_id"`
}
