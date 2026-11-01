package sdk

import "sumeru/core/orm"

// BindCompute registers fn as the value of field.
// Generated model code calls this. fn runs on read and on write.
func BindCompute(model, field string, fn orm.ComputeFunc) {
	orm.RegisterCompute(model, field, nil, fn)
}

// BindStatus registers fn for a status field. Generated model code calls this.
// fn runs when that field is written and its result is the stored status.
func BindStatus(model, field string, fn orm.ComputeFunc) {
	orm.RegisterStatus(model, field, fn)
}
