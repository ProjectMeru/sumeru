package models

import (
	"sumeru/core/sdk"
)

type DigestDigest struct {
	sdk.Model `sumeru:"model=digest.digest"`

	Name         sdk.String `sumeru:"required,string=Name"`
	Periodicity  sdk.Selection[string] `sumeru:"string=Periodicity,selection=daily:Daily,weekly:Weekly,monthly:Monthly,default=weekly"`
	NextRunDate  sdk.Date   `sumeru:"string=Next Run Date"`
	Active       sdk.Boolean `sumeru:"string=Active,default=true"`
	KpiIDs       sdk.One2Many[DigestKpi] `sumeru:"string=KPIs"`
}
