package models

import (
	"sumeru/core/sdk"
)

type CoreCompany struct {
	sdk.Model `sumeru:"model=core.company"`

	Name                     sdk.String  `sumeru:"required,unique,index,string=Company Name"`
	Street                   sdk.String  `sumeru:"string=Street"`
	Street2                  sdk.String  `sumeru:"string=Street 2"`
	City                     sdk.String  `sumeru:"string=City"`
	Zip                      sdk.String  `sumeru:"string=Zip"`
	State                    sdk.String  `sumeru:"string=State / Province"`
	Country                  sdk.String  `sumeru:"string=Country"`
	Email                    sdk.String  `sumeru:"string=Email"`
	Phone                    sdk.String  `sumeru:"string=Phone"`
	Mobile                   sdk.String  `sumeru:"string=Mobile"`
	Website                  sdk.String  `sumeru:"string=Website"`
	Vat                      sdk.String  `sumeru:"index,string=Tax ID"`
	CompanyRegistry          sdk.String  `sumeru:"string=Company ID"`
	InternalNotes            sdk.Text    `sumeru:"string=Internal Notes"`
	MailChatterEnabled       sdk.Boolean `sumeru:"string=Chatter,default=true"`
	MailActivityPanelEnabled sdk.Boolean `sumeru:"string=Activity panel,default=true"`
}
