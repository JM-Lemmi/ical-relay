package main

import (
	"crypto/md5"
	"fmt"
	"net/http"

	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"request": uuid.New().String()})
	// load profile
	vars := mux.Vars(r)
	profileName := vars["profile"]
	profile, ok := conf.Profiles[profileName]
	if !ok {
		errorMsg := fmt.Sprintf("Profile '%s' doesn't exist", profileName)
		requestLogger.Infoln(errorMsg)
		http.Error(w, errorMsg, 404)
		return
	}
	// request original ical
	response, err := http.Get(conf.URL)
	if err != nil {
		// throw error
		requestLogger.Errorln(err)
	}
	// parse original calendar
	calendar, err := ics.ParseCalendar(response.Body)
	if err != nil {
		requestLogger.Errorln(err)
	}
	// create new calendar, excluding based on pattern
	newCalendar := ics.NewCalendar()
	newCalendar.CalendarProperties = []ics.CalendarProperty{
		ics.CalendarProperty{BaseProperty: ics.BaseProperty{IANAToken: "METHOD", Value: "PUBLISH"}},
		ics.CalendarProperty{BaseProperty: ics.BaseProperty{IANAToken: "VERSION", Value: "2.0"}},
		ics.CalendarProperty{BaseProperty: ics.BaseProperty{IANAToken: "PRODID", Value: "-//ical-relay//" + profileName}},
	}
	excludedEvents := 0
	for _, event := range calendar.Events() {
		// extract summary from original event
		summary := event.GetProperty(ics.ComponentPropertySummary).Value
		// check if one of the profiles regex's matches summary
		exclude := false
		for _, excludeRe := range profile.RegEx {
			if excludeRe.MatchString(summary) {
				exclude = true
				break
			}
		}
		if !exclude {
			// add event to new calendar
			// overwrite uid to prevent conflicts with original ical stream
			h := md5.New()
			h.Write([]byte(event.Id()))
			h.Write([]byte(conf.URL))
			id := fmt.Sprintf("%x@%s", h.Sum(nil), "ical-relay")
			newEvent := newCalendar.AddEvent(id)
			// exclude organizer property due to broken escaping
			for _, property := range event.Properties {
				if (property.IANAToken != string(ics.ComponentPropertyOrganizer)) && (property.IANAToken != string(ics.ComponentPropertyUniqueId)) {
					newEvent.Properties = append(newEvent.Properties, property)
				}
			}
			sequenceProperty := ics.IANAProperty{BaseProperty: ics.BaseProperty{IANAToken: "SEQUENCE", Value: "0"}}
			newEvent.Properties = append(newEvent.Properties, sequenceProperty)
		} else {
			excludedEvents++
			requestLogger.Debugf("Excluding event with summary '%s'\n", summary)
		}
	}
	// make sure new calendar has all events but excluded
	eventCountDiff := len(newCalendar.Events()) + excludedEvents - len(calendar.Events())
	if eventCountDiff == 0 {
		requestLogger.Debugf("Output validation successfull; event counts match")
	} else {
		requestLogger.Warnf("This shouldn't happen, event count diff: %d", eventCountDiff)
	}
	requestLogger.Debugf("Excluded %d events", excludedEvents)
	// return new calendar
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.ics", profileName))
	fmt.Fprint(w, newCalendar.Serialize())
}
