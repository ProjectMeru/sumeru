package models

import "sumeru/core/sdk"

type SysPropertyDefinition struct {
	sdk.Model `sumeru:"model=sys.property.definition"`

	ResModel  sdk.String  `sumeru:"required,index,column=model,string=Model"`
	FieldName sdk.String  `sumeru:"required,index,string=Field"`
	Key       sdk.String  `sumeru:"required,index,string=Key"`
	Label     sdk.String  `sumeru:"string=Label"`
	ValueType sdk.String  `sumeru:"required,string=Type,selection=char:Text,integer:Integer,boolean:Boolean,date:Date,selection:Selection,many2one:Many2one"`
	Selection sdk.Text    `sumeru:"string=Selection"`
	Relation  sdk.String  `sumeru:"string=Relation"`
	Required  sdk.Boolean `sumeru:"string=Required,default=false"`
	Sequence  sdk.Integer `sumeru:"string=Sequence,default=10"`
}
