package models

import (
	"sumeru/core/sdk"
)

type CalendarEvent struct {
	sdk.Model `sumeru:"model=calendar.event"`

	Name        sdk.String                   `sumeru:"required,string=Meeting Subject"`
	Start       sdk.DateTime                 `sumeru:"required,string=Start"`
	Stop        sdk.DateTime                 `sumeru:"required,string=Stop"`
	Allday      sdk.Boolean                  `sumeru:"string=All Day,default=false"`
	PartnerIDs  sdk.Many2Many[CorePartner]   `sumeru:"string=Attendees,table=calendar_event_partner_rel,left=event_id,right=partner_id"`
	UserID      sdk.Many2One[CoreUser]       `sumeru:"string=Organizer"`
	ResModel    sdk.String                   `sumeru:"index,column=model,string=Linked Document Model"`
	ResID       sdk.Integer                  `sumeru:"index,string=Linked Document"`
	Description sdk.Text                     `sumeru:"string=Description"`
	Active      sdk.Boolean                  `sumeru:"string=Active,default=true"`
}
