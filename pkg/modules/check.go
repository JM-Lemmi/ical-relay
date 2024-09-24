package modules

import (
	ics "github.com/arran4/golang-ical"
)

func CheckXWRTimezone(cal *ics.Calendar) bool {
	for _, prop := range cal.CalendarProperties {
		if prop.IANAToken == "TIMEZONE" || prop.IANAToken == "X-WR-TIMEZONE" {
			return true
		}
	}
	return false
}
