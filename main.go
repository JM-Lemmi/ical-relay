package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var version = "1.2.0-beta.3"

var configPath = "config.yml"
var conf Config

var router *mux.Router

func main() {
	if len(os.Args) >= 2 {
		configPath = os.Args[1]
	}

	log.Info("Welcome to ical-relay, version " + version)
	log.Debug("Debug log is enabled") // only shows if Debug is actually enabled

	// load config
	var err error
	conf, err = ParseConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(conf.Server.LogLevel)

	// setup routes
	router = mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/profiles/{profile}", profileHandler).Name("profile")
	router.HandleFunc("/api/calendars", calendarlistApiHandler)
	router.HandleFunc("/api/reloadconfig", reloadConfigApiHandler)
	router.HandleFunc("/api/notifier/{notifier}/addrecipient", addNotifyRecipientApiHandler).Name("notifier")

	log.Debug("Starting notifiers")
	NotifierStartup()

	// listen and serve
	address := conf.Server.Addr
	log.Fatal(http.ListenAndServe(address, router))
}
