package models

import (
	"sumeru/core/sdk"
)

type SysFieldAccess struct {
	sdk.Model `sumeru:"model=sys.field.access"`

	Name      sdk.String              `sumeru:"required,unique,string=Name"`
	ResModel  sdk.String              `sumeru:"required,index,column=model,string=Model"`
	FieldName sdk.String              `sumeru:"required,string=Field"`
	GroupID   sdk.Many2One[CoreGroup] `sumeru:"string=Group"`
	PermRead  sdk.Boolean             `sumeru:"string=Read,default=true"`
	PermWrite sdk.Boolean             `sumeru:"string=Write,default=true"`
}
