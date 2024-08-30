package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jm-lemmi/ical-relay/datastore"
	"github.com/jm-lemmi/ical-relay/helpers"

	log "github.com/sirupsen/logrus"
)

func checkAuthorization(tokenString string, profileName string) bool {
	if !dataStore.ProfileExists(profileName) {
		log.Errorf("profile '%s' doesn't exist", profileName)
		return false
	}
	for _, token := range dataStore.GetProfileByName(profileName).Tokens {
		if token.Token == tokenString {
			return true
		}
	}
	return checkSuperAuthorization(tokenString)
}

func checkSuperAuthorization(token string) bool {
	if helpers.Contains(conf.Server.SuperTokens, token) {
		return true
	} else {
		return false
	}
}

// Path: /api/calendars
func calendarlistApiHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": "/api/calendars"})
	requestLogger.Infoln("New API-Request!")

	var callist []string
	token := r.Header.Get("Authorization")
	if !checkSuperAuthorization(token) {
		callist = dataStore.GetPublicProfileNames()
	} else {
		callist = dataStore.GetAllProfileNames()
	}

	w.Header().Set("Content-Type", "application/json")
	caljson, err := json.Marshal(callist)
	if err != nil {
		requestLogger.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(caljson)
}

// Path: /api/profiles/{profile}
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

	profileName := mux.Vars(r)["profile"]

	type profileJson struct {
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

		if dataStore.ProfileExists(profileName) {
			requestLogger.Errorln("Profile already exists!")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: Profile already exists!\n")
			return
		}

		dataStore.AddProfile(profileName, newProfile.Sources, newProfile.Public, newProfile.ImmutablePast)

		requestLogger.Infoln("Created new profile: " + profileName)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Created new profile: "+profileName+"\n")

	// TODO MethodPatch, to only edit singe aspects

	case http.MethodPut:
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

		if !dataStore.ProfileExists(profileName) {
			requestLogger.Errorln("Profile doesnt exist!")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: Profile doesnt exist!\n")
			return
		}

		dataStore.EditProfile(profileName, newProfile.Sources, newProfile.Public, newProfile.ImmutablePast)

		requestLogger.Infoln("Edited profile: " + profileName)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Edited profile: "+profileName+"\n")

	case http.MethodDelete:
		// Delete profile
		if !dataStore.ProfileExists(profileName) {
			requestLogger.Errorln("Profile does not exist!")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: Profile does not exist!\n")
			return
		}

		dataStore.DeleteProfile(profileName)

		requestLogger.Infoln("Deleted profile: " + profileName)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Deleted profile: "+profileName+"\n")
	}
}

// Path: /api/notifier/{notifier}/recipient
func NotifyRecipientApiHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": r.Method + " " + r.URL.Path})
	requestLogger.Infoln("New API-Request!")

	// check if notifier exists; create it, if a matching profile exists
	notifier := mux.Vars(r)["notifier"]
	if !dataStore.NotifierExists(notifier) {
		requestLogger.Warnln("Notifier does not exist")
		if !dataStore.ProfileExists(notifier) {
			requestLogger.Errorln("Profile does not exist either.")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Error: Profile and Notifier does not exist\n")
			return
		} else {
			requestLogger.Infoln("Profile exists, but not the notifier. Creating notifier...")
			err := dataStore.AddNotifierFromProfile(notifier, conf.Server.URL)
			if err != nil {
				requestLogger.Errorln(err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "Error: "+err.Error()+"\n")
				return
			} else {
				requestLogger.Infoln("Notifier created.")
			}
		}
	}

	mail := r.URL.Query().Get("mail")
	if !helpers.ValidMail(mail) {
		requestLogger.Errorln("Invalid mail address")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Invalid mail address\n")
		return
	}

	switch r.Method {
	case http.MethodPost:
		err := dataStore.AddNotifyRecipient(notifier, datastore.Recipient{Recipient: mail, Type: "email"})
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
		err := dataStore.RemoveNotifyRecipient(notifier, datastore.Recipient{Recipient: mail, Type: "email"})
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

	if !dataStore.ProfileExists(profileName) {
		requestLogger.Infoln("Profile " + profileName + " not found!")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Profile "+profileName+" not found!\n")
		return
	}

	if !checkAuthorization(token, profileName) {
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

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Could not read the body in calendarEntryApiHandler -- failed with %s", err.Error())
		}
		err = json.Unmarshal(body, &entry)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rule := datastore.Rule{
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

		dataStore.AddRule(profileName, rule)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Added Rule with filter-type 'id' to profile "+profileName+"\n")
	case http.MethodPut:
		// TODO: Implement
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprint(w, "Not implemented yet!\n")
	case http.MethodDelete:
		rule := datastore.Rule{
			Filters: []map[string]string{
				{
					"type": "id",
					"id":   id,
				},
			},
			Action: map[string]string{"type": "delete"},
		}
		dataStore.AddRule(profileName, rule)
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

	if !dataStore.ProfileExists(profileName) {
		requestLogger.Infoln("Profile " + profileName + " not found!")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Profile "+profileName+" not found!\n")
		return
	}

	if !checkAuthorization(token, profileName) {
		requestLogger.Warnln("Authorization not successful!")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized!\n")
		return
	}

	switch r.Method {
	case http.MethodPost:
		var eventjson map[string]string

		// read json from body to calendar struct
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Could not read the body in newentryfileApiHandler -- failed with %s", err.Error())
		}
		err = json.Unmarshal(body, &eventjson)
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// create event
		urlarr := strings.Split(conf.Server.URL, "://")
		var url string
		if len(url) < 2 { // failsafe, if Server.URL is does not contain URIdentifier
			url = "localhost"
		} else {
			url = urlarr[1]
		}
		event := ics.NewEvent(uuid.New().String() + "@" + url)

		if eventjson["summary"] != "" {
			event.SetSummary(eventjson["summary"])
		}
		if eventjson["location"] != "" {
			event.SetLocation(eventjson["location"])
		}
		if eventjson["start"] != "" {
			start, err := time.Parse(time.RFC3339, eventjson["start"])
			if err != nil {
				requestLogger.Errorln(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			event.SetStartAt(start)
		} else {
			requestLogger.Errorln("No start time given!")
			http.Error(w, "No start time given!", http.StatusBadRequest)
			return
		}
		if eventjson["end"] != "" {
			end, err := time.Parse(time.RFC3339, eventjson["end"])
			if err != nil {
				requestLogger.Errorln(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			event.SetEndAt(end)
		} else {
			requestLogger.Errorln("No end time given!")
			http.Error(w, "No end time given!", http.StatusBadRequest)
			return
		}
		if eventjson["description"] != "" {
			event.SetDescription(eventjson["description"])
		}

		// create calendar
		cal := ics.NewCalendar()
		cal.AddVEvent(event)

		// convert calendar to base64
		source := "base64://" + base64.StdEncoding.EncodeToString([]byte(cal.Serialize()))

		// create source
		dataStore.AddSource(profileName, source)
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

	if !dataStore.ProfileExists(profileName) {
		requestLogger.Infoln("Profile " + profileName + " not found!")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Profile "+profileName+" not found!\n")
		return
	}

	if !checkAuthorization(token, profileName) {
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
		for _, infile := range r.MultipartForm.File {
			file, err := infile[0].Open()
			if err != nil {
				requestLogger.Errorln(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(file)
			file.Close()
			b64file := base64.StdEncoding.EncodeToString(buf.Bytes())

			// Check if file can be parsed
			_, parse_err := ics.ParseCalendar(bytes.NewReader(buf.Bytes()))

			if parse_err != nil {
				requestLogger.Errorln(parse_err)
				http.Error(w, parse_err.Error(), http.StatusUnprocessableEntity)
				return
			}

			// create source
			source := "base64://" + b64file
			log.Debug("Adding source " + source)
			dataStore.AddSource(profileName, source)
			ok(w, requestLogger)
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

	if !dataStore.ProfileExists(profileName) {
		requestLogger.Infoln("Profile " + profileName + " not found!")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Profile "+profileName+" not found!\n")
		return
	}

	if !checkAuthorization(token, profileName) {
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
		var rule datastore.Rule

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Could not read the body in rulesApiHandler -- failed with %s", err.Error())
		}
		err = json.Unmarshal(body, &rule)
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// TODO implement
		if !rule.CheckRuleIntegrity() {
			requestLogger.Errorln("Rule is invalid!")
			http.Error(w, "Rule is invalid!", http.StatusBadRequest)
			return
		}

		dataStore.AddRule(profileName, rule)
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

		dataStore.RemoveRule(profileName, datastore.Rule{Id: idint})
	}
}

// Path: /api/profiles/{profile}/checkAuth
func checkAuthorizationApiHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": r.URL.Path})
	requestLogger.Infoln("New API-Request!")

	token := r.Header.Get("Authorization")
	profileName := vars["profile"]

	if !dataStore.ProfileExists(profileName) {
		requestLogger.Infoln("Profile " + profileName + " not found!")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Profile "+profileName+" not found!\n")
		return
	}

	if checkAuthorization(token, profileName) {
		requestLogger.Infoln("Authorization successful!")
		ok(w, requestLogger)
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
		ok(w, requestLogger)
	} else {
		requestLogger.Infoln("Authorization not successful!")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized!\n")
		return
	}
}

// Path: /api/profiles/{profile}/tokens
func tokenEndpoint(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "api": r.Method + " " + r.URL.Path})
	requestLogger.Infoln("New API-Request!")

	if !checkSuperAuthorization(r.Header.Get("Authorization")) {
		requestLogger.Warnln("Attempted to open admin interface without being admin.")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Forbidden!\n")
		return
	}

	vars := mux.Vars(r)
	profileName := vars["profile"]

	if !dataStore.ProfileExists(profileName) {
		requestLogger.Infoln("Profile " + profileName + " not found!")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Profile "+profileName+" not found!\n")
		return
	}

	var bodyData map[string]interface{}
	if r.Method != http.MethodGet {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Could not read the body in tokenEndpoint -- failed with %s", err.Error())
		}
		err = json.Unmarshal(body, &bodyData)
		if err != nil {
			requestLogger.Errorf("Error while parsing json body in tokenEndpoint -- failed with %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		tokens, err := json.Marshal(dataStore.GetProfileByName(profileName).Tokens)
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(tokens)
	case http.MethodPut:
		note, noteExists := bodyData["note"].(string)
		var err error
		if noteExists {
			err = dataStore.CreateToken(profileName, &note)
		} else {
			err = dataStore.CreateToken(profileName, nil)
		}
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			ok(w, requestLogger)
		}
	case http.MethodPatch:
		note, noteExists := bodyData["note"].(string)
		var err error
		if noteExists {
			err = dataStore.ModifyTokenNote(profileName, bodyData["token"].(string), &note)
		} else {
			err = dataStore.ModifyTokenNote(profileName, bodyData["token"].(string), nil)
		}
		if err != nil {
			requestLogger.Errorln(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			ok(w, requestLogger)
		}
	case http.MethodDelete:
		err := dataStore.DeleteToken(profileName, bodyData["token"].(string))
		if err != nil {
			requestLogger.Errorln(err)
			if strings.Contains(err.Error(), "does not exist") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			ok(w, requestLogger)
		}
	default:
		requestLogger.Errorln("Invalid request Method")
		http.Error(w, "Invalid request Method", http.StatusBadRequest)
	}
}

func ok(w http.ResponseWriter, requestLogger *log.Entry) {
	_, err := w.Write([]byte(`"ok"`))
	if err != nil {
		requestLogger.Error("Failed to send data to client!")
	}
}
