package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var htmlTemplates = template.Must(template.ParseGlob("templates/*.html"))

type eventData map[string]interface{}
type calendarDataByDay map[string][]eventData

func tryRenderErrorOrFallback(w http.ResponseWriter, r *http.Request, statusCode int, err error, fallback string) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r)})
	requestLogger.Errorln(err)
	w.WriteHeader(statusCode)
	err = htmlTemplates.ExecuteTemplate(w, "error.html", map[string]interface{}{
		"Error":    err.Error(),
		"Profiles": getProfilesMetadata(),
	})
	if err != nil {
		requestLogger.Errorln(err)
		http.Error(w, fallback, statusCode)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r)})
	requestLogger.Infoln("index request")

	err := htmlTemplates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Profiles": getProfilesMetadata(),
	})
	if err != nil {
		tryRenderErrorOrFallback(w, r, http.StatusInternalServerError, err, "Internal Server Error")
		return
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r)})
	requestLogger.Infoln("login request")
	err := htmlTemplates.ExecuteTemplate(w, "login.html", map[string]interface{}{
		"Profiles": getProfilesMetadata(),
	})
	if err != nil {
		tryRenderErrorOrFallback(w, r, http.StatusInternalServerError, err, "Internal Server Error")
		return
	}
}

func editViewHandler(w http.ResponseWriter, r *http.Request) {
	// simple dummy handler for now
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "profile": vars["profile"]})
	requestLogger.Infoln("edit view request")
	profileName := vars["profile"]
	profile, ok := conf.Profiles[profileName]
	if !ok {
		err := fmt.Errorf("profile '%s' doesn't exist", profileName)
		requestLogger.Errorln(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// find event by uid in profile
	uid := vars["uid"]
	calendar, err := getProfileCalendar(profile)
	if err != nil {
		requestLogger.Errorln(err)
		tryRenderErrorOrFallback(w, r, http.StatusInternalServerError, err, err.Error())
		return
	}
	var event *ics.VEvent
	for _, e := range calendar.Events() {
		if e.GetProperty("UID").Value == uid {
			event = e
			break
		}
	}
	htmlTemplates.ExecuteTemplate(w, "edit.html", map[string]interface{}{
		"ProfileName": profileName,
		"Event":       event,
		"Profiles":    getProfilesMetadata(),
	})
}

func monthlyViewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "profile": vars["profile"]})
	requestLogger.Infoln("montly view request")
	profileName := vars["profile"]
	profile, ok := conf.Profiles[profileName]
	if !ok {
		err := fmt.Errorf("profile '%s' doesn't exist", profileName)
		tryRenderErrorOrFallback(w, r, http.StatusNotFound, err, err.Error())
		return
	}
	calendar, err := getProfileCalendar(profile)
	if err != nil {
		tryRenderErrorOrFallback(w, r, http.StatusInternalServerError, err, "Internal Server Error")
		return
	}
	allEvents := getEventsByDay(calendar)

	htmlTemplates.ExecuteTemplate(w, "monthly.html", map[string]interface{}{
		"ProfileName": profileName,
		"AllEvents":   allEvents,
		"Profiles":    getProfilesMetadata(),
	})
}

func getEventsByDay(calendar *ics.Calendar) calendarDataByDay {
	calendarDataByDay := make(calendarDataByDay)
	for _, event := range calendar.Events() {
		startTime, err := event.GetStartAt()
		if err != nil {
			log.Errorln(err)
			continue
		}
		endTime, err := event.GetEndAt()
		if err != nil {
			log.Errorln(err)
			continue
		}
		day := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)
		calendarDataByDay[day.Format("2006-01-02")] = append(calendarDataByDay[day.Format("2006-01-02")], eventData{
			"title":    event.GetProperty("SUMMARY").Value,
			"location": event.GetProperty("LOCATION").Value,
			"start":    startTime,
			"end":      endTime,
			"id":       event.GetProperty("UID").Value,
		})
	}
	return calendarDataByDay
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "profile": vars["profile"]})
	requestLogger.Infoln("New Request!")
	profile, ok := conf.Profiles[vars["profile"]]
	if !ok {
		err := fmt.Errorf("profile '%s' doesn't exist", vars["profile"])
		requestLogger.Errorln(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	calendar, err := getProfileCalendar(profile)
	if err != nil {
		requestLogger.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// return new calendar
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.ics", vars["profile"]))
	fmt.Fprint(w, calendar.Serialize())
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
