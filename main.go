package main

import (
	"flag"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var version = "2.0.0-beta.3.2"

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

	// setup template path
	htmlTemplates = template.Must(template.ParseGlob(conf.Server.TemplatePath + "*.html"))

	// setup routes
	router = mux.NewRouter()
	initHandlers()

	// start notifiers
	NotifierStartup()
	// start cleanup
	CleanupStartup()

	// start server
	address := conf.Server.Addr
	log.Info("Starting server on " + address)
	log.Fatal(http.ListenAndServe(address, router))
}
