package orm

import (
	"sumeru/core/modelmeta"
)

type SysSavedSearch struct {
	modelmeta.ModelMeta `sumeru:"model=swc.saved.search"`

	Name      modelmeta.String             `sumeru:"required"`
	UserID    modelmeta.Many2One[CoreUser] `sumeru:"required,index,string=User"`
	ResModel  modelmeta.String             `sumeru:"required,column=model,index"`
	ActionID  modelmeta.Integer            `sumeru:"index"`
	Search    modelmeta.String             `sumeru:"column=search_query"`
	Filter    modelmeta.String             `sumeru:"column=filter_csv"`
	Domain    modelmeta.Text               `sumeru:"column=domain_json"`
	GroupBy   modelmeta.String             `sumeru:"column=group_by"`
	IsDefault modelmeta.Boolean            `sumeru:"column=is_default"`
	IsShared  modelmeta.Boolean            `sumeru:"column=is_shared,index"`
}
