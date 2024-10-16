package compare

import (
	"reflect"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

func Compare(cal1 *ics.Calendar, cal2 *ics.Calendar) ([]ics.VEvent, []ics.VEvent, []ics.VEvent, []ics.VEvent) {
	// Compare the two calendars
	// cal1 is the old calendar, cal2 is the new calendar
	// Returns array of arrays. Added, Deleted, Changed_Old and Changed_New Events
	// The Old and New Events in the Changed arrays are on the same position.

	var added []ics.VEvent
	var deleted []ics.VEvent
	var changed_old []ics.VEvent
	var changed_new []ics.VEvent

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
			if !reflect.DeepEqual(event1.GetProperty("Summary"), event2.GetProperty("Summary")) || !reflect.DeepEqual(event1.GetProperty(ics.ComponentPropertyDtStart), event2.GetProperty(ics.ComponentPropertyDtStart)) || !reflect.DeepEqual(event1.GetProperty(ics.ComponentPropertyDtEnd), event2.GetProperty(ics.ComponentPropertyDtEnd)) || !reflect.DeepEqual(event1.GetProperty(ics.ComponentPropertyDescription), event2.GetProperty(ics.ComponentPropertyDescription)) || !reflect.DeepEqual(event1.GetProperty(ics.ComponentPropertyLocation), event2.GetProperty(ics.ComponentPropertyLocation)) {
				log.Debug("Event changed: ", event1.Id())
				changed_old = append(changed_old, *event1)
				changed_new = append(changed_new, *event2)
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

	return added, deleted, changed_old, changed_new
}
