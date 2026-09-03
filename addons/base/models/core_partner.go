package models

import (
	"sumeru/core/sdk"
)

type CorePartner struct {
	sdk.Model `sumeru:"model=core.partner"`

	Name                        sdk.String                     `sumeru:"required,string=Name"`
	Image                       sdk.Text                       `sumeru:"string=Image"`
	Email                       sdk.String                     `sumeru:"string=Email"`
	Phone                       sdk.String                     `sumeru:"string=Phone"`
	Street                      sdk.String                     `sumeru:"string=Street"`
	CityID                      sdk.Many2One[CoreCity]         `sumeru:"string=City"`
	StateID                     sdk.Many2One[CoreCountryState] `sumeru:"string=State"`
	CountryID                   sdk.Many2One[CoreCountry]      `sumeru:"string=Country"`
	Comment                     sdk.Text                       `sumeru:"string=Notes"`
	IsCompany                   sdk.Boolean                    `sumeru:"string=Is a Company"`
	Active                      sdk.Boolean                    `sumeru:"string=Active,default=true"`
	PropertyAccountReceivableID sdk.Many2One[sdk.Any]          `sumeru:"string=Receivable Account,comodel=account.account"`
	PropertyAccountPayableID    sdk.Many2One[sdk.Any]          `sumeru:"string=Payable Account,comodel=account.account"`
	Color                       sdk.Integer                    `sumeru:"string=Color Index"`
	Gender                      sdk.String                     `sumeru:"string=Gender,selection=male:Male,female:Female,other:Other"`
}
