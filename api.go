package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		requestLogger.Warnln("Notifier does not exist")
		if !conf.profileExists(notifier) {
			requestLogger.Errorln("Profile does not exist either.")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Error: Profile and Notifier does not exist\n")
			return
		} else {
			requestLogger.Infoln("Profile exists, but not the notifier. Creating notifier...")
			conf.addNotifierFromProfile(notifier)
		}
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

func calendarEntryApiHandler(w http.ResponseWriter, r *http.Request) {
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

	if !checkAuthoriziation(token, profileName) {
		requestLogger.Warnln("Authorization not successful!")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized!\n")
		return
	}

	id := r.URL.Query().Get("id")

	switch r.Method {
	case http.MethodGet:
		// TODO: Implement
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprint(w, "Not implemented yet!\n")
	case http.MethodPost:
		var entry map[string]interface{}

		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &entry)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		module := map[string]string{"name": "edit-byid", "id": id, "overwrite": "true"}

		_, ok := entry["summary"]
		if ok {
			module["new-summary"] = entry["summary"].(string)
		}

		_, ok = entry["location"]
		if ok {
			module["new-location"] = entry["location"].(string)
		}

		_, ok = entry["start"]
		if ok {
			module["new-start"] = entry["start"].(string)
		}

		_, ok = entry["end"]
		if ok {
			module["new-end"] = entry["end"].(string)
		}

		_, ok = entry["description"]
		if ok {
			module["new-description"] = entry["description"].(string)
		}

		conf.addModule(profileName, module)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Added Module edit-byid to profile "+profileName+"\n")
	case http.MethodPut:
		// TODO: Implement
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprint(w, "Not implemented yet!\n")
	case http.MethodDelete:
		module := map[string]string{"name": "delete-byid", "id": id}
		conf.addModule(profileName, module)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Added Module to delete entry with id "+id+"\n")
	}
}

func modulesApiHandler(w http.ResponseWriter, r *http.Request) {
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

	if !checkAuthoriziation(token, profileName) {
		requestLogger.Warnln("Authorization not successful!")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized!\n")
		return
	}

	switch r.Method {
	case http.MethodGet:
		// TODO: Implement
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprint(w, "Not implemented yet!\n")
	case http.MethodPost:
		var module map[string]string

		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &module)
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if module["name"] == "" {
			requestLogger.Errorln("No module name given!")
			http.Error(w, "No module name given!", http.StatusBadRequest)
			return
		}

		if !checkSuperAuthorization(token) {
			requestLogger.Debugln("Running in low-privilege mode!")
			if !contains(lowPrivModules, module["name"]) {
				requestLogger.Warnln("Module " + module["name"] + " not allowed in low-privilege mode!")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Module "+module["name"]+" not allowed in low-privilege mode!\n")
				return
			}
		}

		conf.addModule(profileName, module)
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
