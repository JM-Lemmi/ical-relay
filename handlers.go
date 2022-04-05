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
	var addedEvents int
	var excludedEvents int

	//exclude regex
	for _, excludeRe := range profile.RegEx {
		excludedEvents = removeByRegexSummaryAndTime(calendar, excludeRe, profile.From, profile.Until)
	}

	// legacy loop until #23 is resolved
	for _, event := range calendar.Events() {
		calendar.AddVEvent(event)
	}

	// read additional ical files
	if _, err := os.Stat("addical.ics"); err == nil { //this if is legacy until #20 is resolved, keep only content
		count, err := addEventsFile(calendar, "addical.ics")
		if err != nil {
			requestLogger.Errorln(err)
		}
		addedEvents = addedEvents + count
	}

	// read additional ical urls
	if len(profile.AddURL) != 0 { // this if is legacy until #20 is resolved, keep only content
		for _, url := range profile.AddURL {
			count, err := addEventsURL(calendar, url)
			if err != nil {
				http.Error(w, fmt.Sprint(err), 500)
			}
			addedEvents += count
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
