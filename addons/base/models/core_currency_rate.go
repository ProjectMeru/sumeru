package models

import (
	"sumeru/core/sdk"
)

// CoreCurrencyRate holds the value of one unit of CurrencyID in the platform
// reference currency (the currency whose rate is 1.0), effective from DateFrom.
//
// Rates are global in this first iteration; verticals (e.g. accounting) can
// layer company-scoped rates and rate providers on top of this primitive.
// Lookups resolve the newest DateFrom <= the conversion date (see
// sdk.ConvertCurrency).
type CoreCurrencyRate struct {
	sdk.Model `sumeru:"model=core.currency.rate"`

	CurrencyID sdk.Many2One[CoreCurrency] `sumeru:"required,index,string=Currency"`
	Rate       sdk.Float64                `sumeru:"required,string=Rate"`
	DateFrom   sdk.Date                   `sumeru:"required,index,string=Date From"`
	Active     sdk.Boolean                `sumeru:"string=Active,default=true"`
}
