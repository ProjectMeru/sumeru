package models

import (
	"sumeru/core/sdk"
)

type SysAttachment struct {
	sdk.Model `sumeru:"model=sys.attachment"`

	Name       sdk.String                `sumeru:"required,string=Name"`
	ResModel   sdk.String                `sumeru:"index,column=model,string=Model"`
	ResID      sdk.Integer               `sumeru:"index,string=Record"`
	Mimetype   sdk.String                `sumeru:"string=MIME Type"`
	FileSize   sdk.Integer               `sumeru:"string=Size"`
	Datas      sdk.Text                  `sumeru:"string=Data"`
	StoreFname sdk.String                `sumeru:"string=Stored Filename"`
	CreateDate sdk.DateTime              `sumeru:"string=Created"`
	CompanyID  sdk.Many2One[CoreCompany] `sumeru:"string=Company"`
}
