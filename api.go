package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

func checkAuthoriziation(token string, profileName string) bool {
	if contains(conf.Profiles[profileName].Tokens, token) || checkSuperAuthorization(token) {
		return true
	} else {
		return false
	}
}

func checkSuperAuthorization(token string) bool {
	if contains(conf.Server.SuperTokens, token) {
		return true
	} else {
		return false
	}
}

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

func NotifyRecipientApiHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": r.Method + " " + r.URL.Path})
	requestLogger.Infoln("New API-Request!")

	notifier := mux.Vars(r)["notifier"]
	if !conf.notifierExists(notifier) {
		requestLogger.Errorln("Notifier does not exist")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Error: Notifier does not exist\n")
		return
	}

	mail := r.URL.Query().Get("mail")
	if !validMail(mail) {
		requestLogger.Errorln("Invalid mail address")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Invalid mail address\n")
		return
	}

	switch r.Method {
	case http.MethodPost:
		err := conf.addNotifyRecipient(notifier, mail)
		if err != nil {
			requestLogger.Errorln(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error: "+err.Error()+"\n")
			return
		} else {
			requestLogger.Infoln("Added " + mail + " to " + notifier)
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Added "+mail+" to "+notifier+"\n")
		}
	case http.MethodDelete:
		err := conf.removeNotifyRecipient(notifier, mail)
		if err != nil {
			requestLogger.Errorln(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error: "+err.Error()+"\n")
			return
		} else {
			requestLogger.Infoln("Removed " + mail + " from " + notifier)
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Removed "+mail+" from "+notifier+"\n")
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func checkAuthorizationApiHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": r.URL.Path})
	requestLogger.Infoln("New API-Request!")

	token := r.Header.Get("Authorization")
	profileName := vars["profile"]

	_, ok := conf.Profiles[profileName]
	if !ok {
		requestLogger.Infoln("Profile " + profileName + " not found!")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Profile "+profileName+" not found!\n")
		return
	}

	if checkAuthoriziation(token, profileName) {
		requestLogger.Infoln("Authorization successful!")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Authorized!\n")
	} else {
		requestLogger.Infoln("Authorization not successful!")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized!\n")
	}
}

func checkSuperAuthorizationApiHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": r.URL.Path})
	requestLogger.Infoln("New API-Request!")

	token := r.Header.Get("Authorization")

	if checkSuperAuthorization(token) {
		requestLogger.Infoln("Authorization successful!")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Authorized!\n")
	} else {
		requestLogger.Infoln("Authorization not successful!")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized!\n")
		return
	}
}
