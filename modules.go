package main

import (
	"time"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

// This function is a wrapper for removeByRegexSummaryAndTime, where the time is any time
func removeByRegexSummary(cal *ics.Calendar, regex regex) int {
	return removeByRegexSummaryAndTime(cal, regex, time.Time{}, time.Unix(1<<63-1-int64((1969*365+1969/4-1969/100+1969/400)*24*60*60), 999999999))
}

// This function is used to remove the events that are in the time range and match the regex string.
// It returns the number of events removed.
func removeByRegexSummaryAndTime(cal *ics.Calendar, regex regex, start time.Time, end time.Time) int {
	var count int
	for i, event := range cal.Events() { // iterate over events
		date, _ := event.GetStartAt()
		if date.After(start) && end.After(date) {
			// event is in time range
			if regex.MatchString(event.GetProperty(ics.ComponentPropertySummary).Value) {
				// event matches regex
				remove(cal.Components, i)
				count++
				log.Debug("Excluding event '" + event.GetProperty(ics.ComponentPropertySummary).Value + "' with id " + event.Id() + "\n")
			}
		}
	}
	return count
}

func removeById(cal *ics.Calendar, id string) int {
	var count int
	for i, event := range cal.Events() { // iterate over events
		if event.Id() == id {
			remove(cal.Components, i)
			count++
			log.Debug("Excluding event with id " + id + "\n")
		}
	}
	return count
}

func remove(slice []ics.Component, s int) []ics.Component {
	return append(slice[:s], slice[s+1:]...)
}
