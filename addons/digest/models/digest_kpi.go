package models

import (
	"sumeru/core/sdk"
)

type DigestKpi struct {
	sdk.Model `sumeru:"model=digest.kpi"`

	Name        sdk.String              `sumeru:"required,string=Name"`
	DigestID    sdk.Many2One[DigestDigest] `sumeru:"required,index,string=Digest"`
	ComputeCode sdk.String              `sumeru:"required,string=Compute Code"`
}
