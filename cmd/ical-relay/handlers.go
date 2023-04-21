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

var templatePath string
var htmlTemplates *template.Template

type eventData map[string]interface{}
type calendarDataByDay map[string][]eventData

func initHandlers() {
	router.HandleFunc("/", indexHandler)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(conf.Server.TemplatePath+"static/"))))
	router.HandleFunc("/view/{profile}", calendarViewHandler).Name("calendarView")
	router.HandleFunc("/view/{profile}/edit/{uid}", editViewHandler).Name("editView")
	router.HandleFunc("/view/{profile}/edit", rulesViewHandler).Name("rulesView")
	router.HandleFunc("/notifier/{notifier}/subscribe", notifierSubscribeHandler).Name("notifierSubscribe")
	router.HandleFunc("/notifier/{notifier}/unsubscribe", notifierUnsubscribeHandler).Name("notifierUnsubscribe")
	router.HandleFunc("/settings", settingsHandler).Name("settings")
	router.HandleFunc("/howto-users", howtoUsersHandler).Name("howtoUsers")
	router.HandleFunc("/profiles/{profile}", profileHandler).Name("profile")
	router.HandleFunc("/api/calendars", calendarlistApiHandler)
	router.HandleFunc("/api/checkSuperAuth", checkSuperAuthorizationApiHandler)
	router.HandleFunc("/api/profiles/{profile}/checkAuth", checkAuthorizationApiHandler).Name("apiCheckAuth")
	router.HandleFunc("/api/reloadconfig", reloadConfigApiHandler)
	router.HandleFunc("/api/notifier/{notifier}/recipient", NotifyRecipientApiHandler).Name("notifier")
	router.HandleFunc("/api/profiles/{profile}/calentry", calendarEntryApiHandler).Name("calentry")
	router.HandleFunc("/api/profiles/{profile}/rules", rulesApiHandler).Name("rules")
	router.HandleFunc("/api/profiles/{profile}/newentryjson", newentryjsonApiHandler).Name("newentryjson")
	router.HandleFunc("/api/profiles/{profile}/newentryfile", newentryfileApiHandler).Name("newentryfile")
}

func getGlobalTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"Profiles":          getProfilesMetadata(),
		"Version":           version,
		"ApplicationName":   conf.Server.Name,
		"FaviconPath":       conf.Server.FaviconPath,
		"ImprintLink":       conf.Server.Imprint,
		"PrivacyPolicyLink": conf.Server.PrivacyPolicy,
		"Router":            router,
	}
}

func tryRenderErrorOrFallback(w http.ResponseWriter, r *http.Request, statusCode int, err error, fallback string) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r)})
	requestLogger.Errorln(err)
	w.WriteHeader(statusCode)
	data := getGlobalTemplateData()
	data["Error"] = err
	err = htmlTemplates.ExecuteTemplate(w, "error.html", data)
	if err != nil {
		requestLogger.Errorln(err)
		http.Error(w, fallback, statusCode)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r)})
	requestLogger.Infoln("index request")

	err := htmlTemplates.ExecuteTemplate(w, "index.html", getGlobalTemplateData())
	if err != nil {
		tryRenderErrorOrFallback(w, r, http.StatusInternalServerError, err, "Internal Server Error")
		return
	}
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r)})
	requestLogger.Infoln("settings request")
	err := htmlTemplates.ExecuteTemplate(w, "settings.html", getGlobalTemplateData())
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
	calendar, err := getProfileCalendar(profile, vars["profile"])
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
	data := getGlobalTemplateData()
	data["ProfileName"] = profileName
	data["Event"] = event
	htmlTemplates.ExecuteTemplate(w, "edit.html", data)
}

func rulesViewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "profile": vars["profile"]})
	requestLogger.Infoln("rules view request")
	profileName := vars["profile"]
	profile, ok := conf.Profiles[profileName]
	if !ok {
		err := fmt.Errorf("profile '%s' doesn't exist", profileName)
		tryRenderErrorOrFallback(w, r, http.StatusNotFound, err, err.Error())
		return
	}
	data := getGlobalTemplateData()
	data["Rules"] = profile.Rules
	data["ProfileName"] = profileName
	htmlTemplates.ExecuteTemplate(w, "rules.html", data)
}

func calendarViewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "profile": vars["profile"]})
	requestLogger.Infoln("calendar view request")
	profileName := vars["profile"]
	profile, ok := conf.Profiles[profileName]
	if !ok {
		err := fmt.Errorf("profile '%s' doesn't exist", profileName)
		tryRenderErrorOrFallback(w, r, http.StatusNotFound, err, err.Error())
		return
	}
	calendar, err := getProfileCalendar(profile, vars["profile"])
	if err != nil {
		tryRenderErrorOrFallback(w, r, http.StatusInternalServerError, err, "Internal Server Error")
		return
	}
	allEvents := getEventsByDay(calendar, profileName)
	data := getGlobalTemplateData()
	data["ProfileName"] = profileName
	data["Events"] = allEvents
	data["ImmutablePast"] = profile.ImmutablePast
	htmlTemplates.ExecuteTemplate(w, "calendar.html", data)
}

func getEventsByDay(calendar *ics.Calendar, profileName string) calendarDataByDay {
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
		edit_url, err := router.Get("editView").URL("profile", profileName, "uid", event.GetProperty("UID").Value)
		if err != nil {
			log.Errorln(err)
		}
		day := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)
		data := eventData{
			"title":    event.GetProperty("SUMMARY").Value,
			"start":    startTime,
			"end":      endTime,
			"id":       event.GetProperty("UID").Value,
			"edit_url": edit_url.String(),
		}
		description := event.GetProperty("DESCRIPTION")
		if description != nil {
			data["description"] = description.Value
		}
		if event.GetProperty("LOCATION") != nil {
			data["location"] = event.GetProperty("LOCATION").Value
		}
		calendarDataByDay[day.Format("2006-01-02")] = append(calendarDataByDay[day.Format("2006-01-02")], data)
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

	// load params
	time := r.URL.Query().Get("reminder")
	if time != "" {
		profile.Rules = append(profile.Rules, Rule{
			Filters: []map[string]string{
				{"type": "all"},
			},
			Action: map[string]string{
				"type": "reminder",
				"time": time,
			},
		})
	}

	calendar, err := getProfileCalendar(profile, vars["profile"])
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

func notifierSubscribeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "notifier": vars["notifier"]})
	requestLogger.Infoln("New Request!")
	notifier, ok := vars["notifier"]
	if !ok {
		err := fmt.Errorf("profile '%s' doesn't exist", vars["notifier"])
		requestLogger.Errorln(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// load params
	data := getGlobalTemplateData()
	data["mail"] = r.URL.Query().Get("mail")
	data["notifier"] = notifier
	data["ProfileName"] = notifier // this is vor the nav header to not break

	htmlTemplates.ExecuteTemplate(w, "subscribe.html", data)
}

func notifierUnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r), "notifier": vars["notifier"]})
	requestLogger.Infoln("New Request!")
	notifier, ok := vars["notifier"]
	if !ok {
		err := fmt.Errorf("profile '%s' doesn't exist", vars["notifier"])
		requestLogger.Errorln(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// load params
	data := getGlobalTemplateData()
	data["mail"] = r.URL.Query().Get("mail")
	data["notifier"] = notifier
	data["ProfileName"] = notifier // this is vor the nav header to not break

	htmlTemplates.ExecuteTemplate(w, "unsubscribe.html", data)
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func howtoUsersHandler(w http.ResponseWriter, r *http.Request) {
	requestLogger := log.WithFields(log.Fields{"client": GetIP(r)})
	requestLogger.Infoln("New Request!")
	data := getGlobalTemplateData()
	htmlTemplates.ExecuteTemplate(w, "howto-users.html", data)
}
