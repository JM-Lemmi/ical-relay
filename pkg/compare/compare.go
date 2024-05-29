package compare

import (
	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

func Compare(cal1 *ics.Calendar, cal2 *ics.Calendar) ([]ics.VEvent, []ics.VEvent, []ics.VEvent) {
	// Compare the two calendars
	// Returns array of arrays. Added, Deleted, Changed Events

	var added []ics.VEvent
	var deleted []ics.VEvent
	var changed []ics.VEvent

	// Create a map of the events
	cal1Map := make(map[string]*ics.VEvent)
	cal2Map := make(map[string]*ics.VEvent)
	for _, event := range cal1.Events() {
		cal1Map[event.Id()] = event
	}
	for _, event := range cal2.Events() {
		cal2Map[event.Id()] = event
	}

	// Compare the two calendars
	for _, event1 := range cal1Map {
		if event2, ok := cal2Map[event1.Id()]; ok {
			// Event exists in both calendars
			if event1.GetProperty("Summary") != event2.GetProperty("Summary") || event1.GetProperty(ics.ComponentPropertyDtStart) != event2.GetProperty(ics.ComponentPropertyDtStart) || event1.GetProperty(ics.ComponentPropertyDtEnd) != event2.GetProperty(ics.ComponentPropertyDtEnd) || event1.GetProperty(ics.ComponentPropertyDescription) != event2.GetProperty(ics.ComponentPropertyDescription) || event1.GetProperty(ics.ComponentPropertyLocation) != event2.GetProperty(ics.ComponentPropertyLocation) {
				log.Debug("Event changed: ", event1.Id())
				changed = append(changed, *event1)
			}
		} else {
			// Event only exists in cal1
			log.Debug("Event deleted: ", event1.Id())
			deleted = append(deleted, *event1)
		}
	}
	for _, event2 := range cal2Map {
		if _, ok := cal1Map[event2.Id()]; !ok {
			// Event only exists in cal2
			log.Debug("Event added: ", event2.Id())
			added = append(added, *event2)
		}
	}

	return added, deleted, changed
}
