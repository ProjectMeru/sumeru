package models

import (
	"sumeru/core/sdk"
)

// UomCategory groups units of measure that can be converted between each
// other (e.g. Weight: kg, g, ton).
type UomCategory struct {
	sdk.Model `sumeru:"model=uom.category"`

	Name sdk.String `sumeru:"required,unique,index,string=Category"`
}

// UomUom is a unit of measure. Factor expresses how many reference units one
// unit of this UoM contains (e.g. in Weight with kg as reference: kg=1,
// g=0.001, ton=1000). Conversions inside a category divide the factors and
// round to the target UoM's Rounding precision (see sdk.ConvertUom).
type UomUom struct {
	sdk.Model `sumeru:"model=uom.uom"`

	Name       sdk.String                `sumeru:"required,index,string=Unit of Measure"`
	CategoryID sdk.Many2One[UomCategory] `sumeru:"required,index,string=Category"`
	Factor     sdk.Float64               `sumeru:"required,string=Ratio,default=1"`
	Rounding   sdk.Float64               `sumeru:"string=Rounding Precision,default=0.01"`
	Active     sdk.Boolean               `sumeru:"string=Active,default=true"`
}
