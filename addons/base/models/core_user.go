package models

import (
	"sumeru/core/sdk"
)

type CoreUser struct {
	sdk.Model `sumeru:"model=core.user"`

	Login          sdk.String                 `sumeru:"required,unique,index,string=Login"`
	Password       sdk.String                 `sumeru:"string=Password"`
	Name           sdk.String                 `sumeru:"string=Name"`
	Image          sdk.Text                   `sumeru:"string=Image"`
	ImageCrop      sdk.Text                   `sumeru:"string=Image Crop"`
	Active         sdk.Boolean                `sumeru:"string=Active,default=true"`
	Email          sdk.String                 `sumeru:"string=Email"`
	Phone          sdk.String                 `sumeru:"string=Work Phone"`
	Mobile         sdk.String                 `sumeru:"string=Mobile"`
	CompanyID      sdk.Many2One[CoreCompany]  `sumeru:"index,string=Company"`
	CompanyIds     sdk.Many2Many[CoreCompany] `sumeru:"string=Companies,table=core_user_company_rel,left=user_id,right=company_id"`
	Lang           sdk.String                 `sumeru:"string=Language,default=en_US"`
	Tz             sdk.String                 `sumeru:"string=Timezone"`
	Signature      sdk.Text                   `sumeru:"string=Email Signature"`
	UserType       sdk.String                 `sumeru:"string=User Type,default=internal,selection=internal:Internal,portal:Portal,public:Public"`
	TotpSecret     sdk.String                 `sumeru:"string=TOTP Secret"`
	TotpEnabled    sdk.Boolean                `sumeru:"string=2FA Enabled,default=false"`
	PasswordMinLen sdk.Integer                `sumeru:"string=Min Password Length,default=8"`
	PinnedApps     sdk.Text                   `sumeru:"string=Pinned Apps"`
}
