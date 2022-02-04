package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type templateData struct {
	Name string
	URL  string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusNoContent)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"request": uuid.New().String()})
	requestLogger.Infoln("Client-addr:", r.RemoteAddr)
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
	response, err := http.Get(profile.URL)
	if err != nil {
		requestLogger.Errorln(err)
		http.Error(w, fmt.Sprintf("Error requesting original URL: %s", err.Error()), 500)
		return
	}
	if response.StatusCode != 200 {
		requestLogger.Errorf("Unexpected status '%s' from original URL\n", response.Status)
		resp, err := ioutil.ReadAll(response.Body)
		if err != nil {
			requestLogger.Errorln(err)
		}
		requestLogger.Debugf("Full response body: %s\n", resp)
		http.Error(w, fmt.Sprintf("Error response from original URL: Status %s", response.Status), 500)
		return
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
	for _, component := range calendar.Components {
		if len(component.SubComponents()) > 1 {
			if len(component.UnknownPropertiesIANAProperties()) > 1 {
				if component.UnknownPropertiesIANAProperties()[0].IANAToken == "TZID" {
					newCalendar.Components = append(newCalendar.Components, component)
				}
			}
		}
	}
	excludedEvents := 0
	for _, event := range calendar.Events() {
		// extract summary and time from original event
		summary := event.GetProperty(ics.ComponentPropertySummary).Value
		date, _ := event.GetStartAt()
		id := event.Id()
		// check if one of the profiles regex's matches summary
		exclude := false
		for _, excludeRe := range profile.RegEx {
			if date.After(profile.From) && profile.Until.After(date) {
				if excludeRe.MatchString(summary) || excludeRe.MatchString(id) {
					exclude = true
					break
				}
			}
		}
		if !exclude {
			// add event to new calendar
			if !profile.PassID {
				// overwrite uid to prevent conflicts with original ical stream
				h := md5.New()
				h.Write([]byte(event.Id()))
				h.Write([]byte(profile.URL))
				id = fmt.Sprintf("%x@%s", h.Sum(nil), "ical-relay")
				event.SetProperty(ics.ComponentPropertyUniqueId, id)
			}
			newCalendar.AddVEvent(event)
		} else {
			excludedEvents++
			requestLogger.Debugf("Excluding event '%s' with id %s\n", summary, id)
		}
	}

	// read additional ical files
	addedEvents := 0
	if _, err := os.Stat("addical.ics"); err == nil {
		addicsfile, _ := os.Open("addical.ics")
		addics, _ := ics.ParseCalendar(addicsfile)
		for _, event := range addics.Events() {
			newCalendar.AddVEvent(event)
			addedEvents++
		}
	}

	// read additional ical url
	if len(profile.AddURL) != 0 {
		for _, url := range profile.AddURL {
			response, err := http.Get(url)
			if err != nil {
				requestLogger.Errorln(err)
				http.Error(w, fmt.Sprintf("Error requesting additional URL: %s", err.Error()), 500)
				return
			}
			if response.StatusCode != 200 {
				requestLogger.Errorf("Unexpected status '%s' from additional URL\n", response.Status)
				resp, err := ioutil.ReadAll(response.Body)
				if err != nil {
					requestLogger.Errorln(err)
				}
				requestLogger.Debugf("Full response body: %s\n", resp)
				http.Error(w, fmt.Sprintf("Error response from additional URL: Status %s", response.Status), 500)
				return
			}
			// parse aditional calendar
			addcal, err := ics.ParseCalendar(response.Body)
			if err != nil {
				requestLogger.Errorln(err)
			}
			// add to new calendar
			for _, event := range addcal.Events() {
				newCalendar.AddVEvent(event)
				addedEvents++
			}
		}
	}

	// make sure new calendar has all events but excluded and added
	eventCountDiff := len(newCalendar.Events()) + excludedEvents - addedEvents - len(calendar.Events())
	if eventCountDiff == 0 {
		requestLogger.Debugf("Output validation successfull; event counts match")
	} else {
		requestLogger.Warnf("This shouldn't happen, event count diff: %d", eventCountDiff)
	}
	requestLogger.Debugf("Excluded %d events and added %d events", excludedEvents, addedEvents)
	// return new calendar
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.ics", profileName))
	fmt.Fprint(w, newCalendar.Serialize())
}

func profileViewHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"request": uuid.New().String()})
	// load profile
	vars := mux.Vars(r)
	profileName := vars["profile"]
	// load template file from box
	templateString, err := templateBox.String("profile.html")
	if err != nil {
		requestLogger.Errorln(err)
	}

	viewTemplate, err := template.New("profile").Parse(templateString)
	if err != nil {
		requestLogger.Errorln(err)
	}
	profileURL, err := router.Get("profile").URL("profile", profileName)
	if err != nil {
		requestLogger.Errorln(err)
	}
	viewTemplate.Execute(w, templateData{Name: profileName, URL: profileURL.String()})
}
