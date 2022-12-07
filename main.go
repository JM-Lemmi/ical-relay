package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var version = "2.0.0-beta.1"

var configPath string
var conf Config

var router *mux.Router

func main() {
	log.Info("Welcome to ical-relay, version " + version)

	var notifier string
	flag.StringVar(&notifier, "notifier", "", "Run notifier with given ID")
	flag.StringVar(&configPath, "config", "config.yml", "Path to config file")
	flag.Parse()

	// load config
	var err error
	conf, err = ParseConfig(configPath)
	if err != nil {
		os.Exit(1)
	}

	log.SetLevel(conf.Server.LogLevel)
	log.Debug("Debug log is enabled") // only shows if Debug is actually enabled

	// run notifier if specified
	if notifier != "" {
		log.Debug("Notifier mode called. Running: " + notifier)
		err := RunNotifier(notifier)
		if err != nil {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	} else {
		log.Debug("Server mode.")
	}

	// setup routes
	router = mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("templates/static/"))))
	router.HandleFunc("/view/{profile}/monthly", monthlyViewHandler).Name("monthlyView")
	router.HandleFunc("/view/{profile}/edit/{uid}", editViewHandler).Name("editView")
	router.HandleFunc("/view/{profile}/edit", modulesViewHandler).Name("modulesView")
	router.HandleFunc("/notifier/{notifier}/subscribe", notifierSubscribeHandler).Name("notifierSubscribe")
	router.HandleFunc("/notifier/{notifier}/unsubscribe", notifierUnsubscribeHandler).Name("notifierUnsubscribe")
	router.HandleFunc("/settings", settingsHandler).Name("settings")
	router.HandleFunc("/profiles/{profile}", profileHandler).Name("profile")
	router.HandleFunc("/api/calendars", calendarlistApiHandler)
	router.HandleFunc("/api/checkSuperAuth", checkSuperAuthorizationApiHandler)
	router.HandleFunc("/api/profiles/{profile}/checkAuth", checkAuthorizationApiHandler).Name("apiCheckAuth")
	router.HandleFunc("/api/reloadconfig", reloadConfigApiHandler)
	router.HandleFunc("/api/notifier/{notifier}/recipient", NotifyRecipientApiHandler).Name("notifier")
	router.HandleFunc("/api/profiles/{profile}/calentry", calendarEntryApiHandler).Name("calentry")
	router.HandleFunc("/api/profiles/{profile}/modules", modulesApiHandler).Name("modules")

	// start notifiers
	NotifierStartup()

	// start server
	address := conf.Server.Addr
	log.Info("Starting server on " + address)
	log.Fatal(http.ListenAndServe(address, router))
}
