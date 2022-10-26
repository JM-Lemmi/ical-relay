package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var htmlTemplates = template.Must(template.ParseGlob("templates/*.html"))

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

func viewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "profile": vars["profile"]})
	requestLogger.Infoln("New Request!")
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
	htmlTemplates.ExecuteTemplate(w, "view.html", map[string]interface{}{
		"ProfileName": profileName,
		"Calendar":    calendar.Events(),
		"Profiles":    getProfilesMetadata(),
	})
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
