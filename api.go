package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	ics "github.com/arran4/golang-ical"
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

// Path: /api/calendars
func calendarlistApiHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": "/api/calendars"})
	requestLogger.Infoln("New API-Request!")

	var callist []string = conf.getPublicCalendars()

	w.Header().Set("Content-Type", "application/json")
	caljson, _ := json.Marshal(callist)
	fmt.Fprint(w, string(caljson)+"\n")
}

// Path: /api/profiles
func profileApiHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": r.Method + " " + r.URL.Path})
	requestLogger.Infoln("New API-Request!")

	token := r.Header.Get("Authorization")
	if !checkSuperAuthorization(token) {
		requestLogger.Warnln("Authorization not successful!")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized!\n")
		return
	}

	type profileJson struct {
		Name          string   `json:"name"`
		Sources       []string `json:"sources"`
		Public        bool     `json:"public"`
		ImmutablePast bool     `json:"immutable_past"`
	}

	switch r.Method {
	case http.MethodPost:
		// Create new profile
		var newProfile profileJson

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&newProfile)
		if err != nil {
			requestLogger.Errorln(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error decoding json: "+err.Error()+"\n")
			return
		}

		if conf.profileExists(newProfile.Name) {
			requestLogger.Errorln("Profile already exists!")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: Profile already exists!\n")
			return
		}

		conf.addProfile(newProfile.Name, newProfile.Sources, newProfile.Public, newProfile.ImmutablePast)

		requestLogger.Infoln("Created new profile: " + newProfile.Name)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Created new profile: "+newProfile.Name+"\n")

	case http.MethodPatch:
		// Update profile
		var newProfile profileJson

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&newProfile)
		if err != nil {
			requestLogger.Errorln(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error decoding json: "+err.Error()+"\n")
			return
		}

		if !conf.profileExists(newProfile.Name) {
			requestLogger.Errorln("Profile doenst exist!")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: Profile doesnt exist!\n")
			return
		}

		conf.editProfile(newProfile.Name, newProfile.Sources, newProfile.Public, newProfile.ImmutablePast)

		requestLogger.Infoln("Edited new profile: " + newProfile.Name)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Edited new profile: "+newProfile.Name+"\n")

	case http.MethodDelete:
		// Delete profile
		profileName := r.URL.Query().Get("name")
		if !conf.profileExists(profileName) {
			requestLogger.Errorln("Profile does not exist!")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: Profile does not exist!\n")
			return
		}

		conf.deleteProfile(profileName)

		requestLogger.Infoln("Deleted profile: " + profileName)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Deleted profile: "+profileName+"\n")
	}
}

// Path: /api/reloadconfig
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

// Path: /api/notifier/{notifier}/recipient
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

// Path: /api/profiles/{profile}/calentry
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

		rule := Rule{
			Filters: []map[string]string{
				{
					"type": "id",
					"id":   id,
				},
			},
			Action: map[string]string{
				"type":      "edit",
				"overwrite": "true",
			},
		}

		_, ok := entry["summary"]
		if ok {
			rule.Action["new-summary"] = entry["summary"].(string)
		}

		_, ok = entry["location"]
		if ok {
			rule.Action["new-location"] = entry["location"].(string)
		}

		_, ok = entry["start"]
		if ok {
			rule.Action["new-start"] = entry["start"].(string)
		}

		_, ok = entry["end"]
		if ok {
			rule.Action["new-end"] = entry["end"].(string)
		}

		_, ok = entry["description"]
		if ok {
			rule.Action["new-description"] = entry["description"].(string)
		}

		conf.addRule(profileName, rule)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Added Rule with filter-type 'id' to profile "+profileName+"\n")
	case http.MethodPut:
		// TODO: Implement
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprint(w, "Not implemented yet!\n")
	case http.MethodDelete:
		rule := Rule{
			Filters: []map[string]string{
				{
					"type": "id",
					"id":   id,
				},
			},
			Action: map[string]string{"type": "delete"},
		}
		conf.addRule(profileName, rule)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Added Rule to delete entry with id "+id+"\n")
	}
}

// Path /api/profiles/{profile}/newentryjson
func newentryjsonApiHandler(w http.ResponseWriter, r *http.Request) {
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
	case http.MethodPost:
		var cal ics.Calendar

		// read json from body to calendar struct
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &cal)
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// convert calendar to base64
		source := "base64://" + base64.StdEncoding.EncodeToString([]byte(cal.Serialize()))

		// create source
		conf.addSource(profileName, source)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Wrong method!\n")
	}
}

// Path: /api/profiles/{profile}/newentryfile
func newentryfileApiHandler(w http.ResponseWriter, r *http.Request) {
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
	case http.MethodPost, http.MethodPut:

		// read file from multipart form and convert to base64
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// read file into buffer and convert to base 64
		for infile, _ := range r.MultipartForm.File {
			file, err := r.MultipartForm.File[infile][0].Open()
			if err != nil {
				requestLogger.Errorln(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(file)
			file.Close()
			b64file := base64.StdEncoding.EncodeToString(buf.Bytes())

			// create source
			source := "base64://" + b64file
			log.Debug("Adding source " + source)
			conf.addSource(profileName, source)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Wrong method!\n")
	}
}

// Path: /api/profiles/{profile}/rules
func rulesApiHandler(w http.ResponseWriter, r *http.Request) {
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
		var rule Rule

		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &rule)
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// TODO implement
		if !checkRuleIntegrity(rule) {
			requestLogger.Errorln("Rule is invalid!")
			http.Error(w, "Rule is invalid!", http.StatusBadRequest)
			return
		}

		conf.addRule(profileName, rule)
	case http.MethodDelete:
		id := r.URL.Query().Get("id")

		if id == "" {
			requestLogger.Errorln("No id given!")
			http.Error(w, "No id given!", http.StatusBadRequest)
			return
		}

		idint, err := strconv.Atoi(id)
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		conf.removeRuleFromProfile(profileName, idint)
	}
}

// Path: /api/profiles/{profile}/checkAuth
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

// Path: /api/checkSuperAuth
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
