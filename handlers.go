package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	ics "github.com/arran4/golang-ical"
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
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "profile": vars["profile"]})
	requestLogger.Infoln("New Request!")

	// load profile
	profileName := vars["profile"]
	profile, ok := conf.Profiles[profileName]
	if !ok {
		errorMsg := fmt.Sprintf("Profile '%s' doesn't exist", profileName)
		requestLogger.Infoln(errorMsg)
		http.Error(w, errorMsg, 404)
		return
	}
	// request original ical
	var calendar *ics.Calendar
	if profile.Source == "" {
		calendar = ics.NewCalendar()
	} else {
		response, err := http.Get(profile.Source)
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
		calendar, err = ics.ParseCalendar(response.Body)
		if err != nil {
			requestLogger.Errorln(err)
		}
	}

	origlen := len(calendar.Events())
	var addedEvents int

	for i := range profile.Modules {
		log.Debug("Requested Module: " + profile.Modules[i]["name"])
		module, ok := modules[profile.Modules[i]["name"]]
		if !ok {
			requestLogger.Warnf(fmt.Sprintf("Module '%s' doesn't exist", profile.Modules[i]["name"]))
			continue
		}
		count, err := callModule(module, profile.Modules[i], calendar)
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, fmt.Sprintf("Error executing module: %s", err.Error()), 500)
			return
		}
		addedEvents += count
	}

	// make sure new calendar has all events but excluded and added
	eventCountDiff := origlen + addedEvents - len(calendar.Events())
	if eventCountDiff == 0 {
		requestLogger.Infoln("Output validation successfull; event counts match")
	} else {
		requestLogger.Warnf("This shouldn't happen, event count diff: %d", eventCountDiff)
	}
	requestLogger.Debugf("Added %d events", addedEvents)
	// return new calendar
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.ics", profileName))
	fmt.Fprint(w, calendar.Serialize())
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
