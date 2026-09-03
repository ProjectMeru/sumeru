package models

import (
	"sumeru/core/sdk"
)

type SysAudit struct {
	sdk.Model `sumeru:"model=sys.audit"`

	UserID     sdk.Many2One[CoreUser] `sumeru:"index,string=User"`
	Action     sdk.String             `sumeru:"required,index,string=Action"`
	ResModel   sdk.String             `sumeru:"index,column=model,string=Model"`
	ResID      sdk.Integer            `sumeru:"string=Record"`
	BeforeJson sdk.Text               `sumeru:"string=Before"`
	AfterJson  sdk.Text               `sumeru:"string=After"`
	Detail     sdk.Text               `sumeru:"string=Detail"`
	CreateDate sdk.DateTime           `sumeru:"required,index,string=When"`
}
