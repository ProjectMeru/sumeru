package applog

import "time"

func ParseLogTimezone(input string) (*time.Location, string) { return parseLogTimezone(input) }

func EffectiveLocationForTest() *time.Location { return effectiveLocation() }

func SetLogLocationForTest(loc *time.Location) { logLocation = loc }
