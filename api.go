package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io/ioutil"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

func calendarlistApiHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": "/api/calendars"})
	requestLogger.Infoln("New API-Request!")

	var callist []string = conf.getPublicCalendars()

	w.Header().Set("Content-Type", "application/json")
	caljson, _ := json.Marshal(callist)
	fmt.Fprint(w, string(caljson)+"\n")
}

func reloadConfigApiHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": r.URL.Path})
	requestLogger.Infoln("New API-Request!")

	err := reloadConfig()
	if err != nil {
		requestLogger.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error: "+err.Error()+"\n")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Config reloaded!\n")
}

func addNotifyRecipientApiHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": r.URL.Path})
	requestLogger.Infoln("New API-Request!")

	body, _ := ioutil.ReadAll(r.Body)
	bodystring := string(body)

	err := conf.addNotifyRecipient(mux.Vars(r)["notifier"], bodystring)
	if err != nil {
		requestLogger.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error: "+err.Error()+"\n")
		return
	} else {
		requestLogger.Infoln("Added " + bodystring + " to " + mux.Vars(r)["notifier"])
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Added " + bodystring + " to " + mux.Vars(r)["notifier"] + "\n")
	}
}
