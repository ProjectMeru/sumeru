package models

import "sumeru/core/sdk"

// PartnerContacts extends core.partner with contacts-specific fields.
type PartnerContacts struct {
	sdk.Model `sumeru:"inherit=core.partner"`

	CustomerRank sdk.Integer `sumeru:"string=Customer Rank,default=0"`
}
