package main

import (
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
	origlen := len(calendar.Events())
	excludedEvents := 0
	for i, component := range calendar.Components {
		// this is a hack to get around the fact that the ical library doesn't provide a way to remove an event
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)
			// extract summary and time from original event
			summary := event.GetProperty(ics.ComponentPropertySummary).Value
			date, _ := event.GetStartAt()
			id := event.Id()
			// check if one of the profiles regex's matches summary
			if date.After(profile.From) && profile.Until.After(date) {
				for _, excludeRe := range profile.RegEx {
					if excludeRe.MatchString(summary) || excludeRe.MatchString(id) {
						// remove event from calendar
						remove(calendar.Components, i)
						excludedEvents++
						requestLogger.Debugf("Excluding event '%s' with id %s\n", summary, id)
						continue
					}
				}
			}
		default:
			continue
		}
	}

	// overwrite uid to prevent conflicts with original ical stream
	if !profile.PassID {
		// deprecated in v0.4.1
		log.Error("PassID is deprecated")
	}

	// read additional ical files
	addedEvents := 0
	if _, err := os.Stat("addical.ics"); err == nil {
		addicsfile, _ := os.Open("addical.ics")
		addics, _ := ics.ParseCalendar(addicsfile)
		for _, event := range addics.Events() {
			calendar.AddVEvent(event)
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
				calendar.AddVEvent(event)
				addedEvents++
			}
		}
	}

	// make sure new calendar has all events but excluded and added
	eventCountDiff := origlen + excludedEvents - addedEvents - len(calendar.Events())
	if eventCountDiff == 0 {
		requestLogger.Debugf("Output validation successfull; event counts match")
	} else {
		requestLogger.Warnf("This shouldn't happen, event count diff: %d", eventCountDiff)
	}
	requestLogger.Debugf("Excluded %d events and added %d events", excludedEvents, addedEvents)
	// return new calendar
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.ics", profileName))
	fmt.Fprint(w, calendar.Serialize())
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

func remove(slice []ics.Component, s int) []ics.Component {
	return append(slice[:s], slice[s+1:]...)
}
