package main

import (
	"net/http"
	"os"
	"flag"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var version = "1.2.0-beta.5"

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
		log.Fatal(err)
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
	router.HandleFunc("/profiles/{profile}", profileHandler).Name("profile")
	router.HandleFunc("/api/calendars", calendarlistApiHandler)
	router.HandleFunc("/api/profiles/{profile}/checkAuth", checkAuthorizationApiHandler)
	router.HandleFunc("/api/reloadconfig", reloadConfigApiHandler)
	router.HandleFunc("/api/notifier/{notifier}/addrecipient", addNotifyRecipientApiHandler).Name("notifier")
	router.HandleFunc("/api/notifier/{notifier}/removerecipient", removeNotifyRecipientApiHandler).Name("notifier")

	// start notifiers
	NotifierStartup()

	// start server
	address := conf.Server.Addr
	log.Fatal(http.ListenAndServe(address, router))
}
