package main

import (
	"time"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

// This function is a wrapper for removeByRegexSummaryAndTime, where the time is any time
func removeByRegexSummary(cal *ics.Calendar, regex regex) int {
	return removeByRegexSummaryAndTime(cal, regex, time.Time{}, time.Unix(1<<63-1-int64((1969*365+1969/4-1969/100+1969/400)*24*60*60), 999999999))
	// this is the maximum time that can be represented in the time.Time struct
}

// This function is a helper for removeByRegexSummaryAndTime. It removes the first element from the calendar matching the regex
// It returns true, if it removed an element and false, if it didnt
func removeSingleByRegexSummaryAndTime(cal *ics.Calendar, regex regex, start time.Time, end time.Time) bool {
	for i, component := range cal.Components { // iterate over events
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)
			date, _ := event.GetStartAt()
			if date.After(start) && end.After(date) {
				// event is in time range
				if regex.MatchString(event.GetProperty(ics.ComponentPropertySummary).Value) {
					// event matches regex
					remove(cal.Components, i)
					log.Debug("Excluding event '" + event.GetProperty(ics.ComponentPropertySummary).Value + "' with id " + event.Id() + "\n")
					return true
				}
			}
		}
	}
	return false
}

// This function is used to remove the events that are in the time range and match the regex string.
// It returns the number of events removed.
func removeByRegexSummaryAndTime(cal *ics.Calendar, regex regex, start time.Time, end time.Time) int {
	var count int
	loop := true
	for {
		loop = removeSingleByRegexSummaryAndTime(cal, regex, start, end)
		if loop {
			count++
		} else {
			break
		}
	}
	return count
}

// This function removes the event with an id matching the string.
// Only removes one event
// It returns the number of events removed.
func removeById(cal *ics.Calendar, id string) int {
	var count int
	for i, component := range cal.Components { // iterate over events
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)
			if event.Id() == id {
				remove(cal.Components, i)
				count++
				log.Debug("Excluding event with id " + id + "\n")
			}
		}
	}
	return count
}

// This function adds all events from cal2 to cal1.
// All other properties, such as TZ are retained from cal1.
func addEvents(cal1 *ics.Calendar, cal2 *ics.Calendar) int {
	var count int
	for _, event := range cal2.Events() {
		cal1.AddVEvent(event)
		count++
	}
	return count
}

func remove(slice []ics.Component, s int) []ics.Component {
	return append(slice[:s], slice[s+1:]...)
}
