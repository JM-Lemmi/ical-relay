package main

import (
	"flag"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var version = "2.0.0-beta.6.3"

var configPath string
var conf Config

var router *mux.Router

func main() {
	log.Info("Welcome to ical-relay, version " + version)

	var notifier string
	flag.StringVar(&notifier, "notifier", "", "Run notifier with given ID")
	flag.StringVar(&configPath, "config", "config.yml", "Path to config file")
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Enables verbose debug output")
	var superverbose bool
	flag.BoolVar(&superverbose, "vv", false, "Enable super verbose trace output")
	importData := flag.Bool("import-data", false, "Whether to import data")
	ephemeral := flag.Bool("ephemeral", false, "Enable ephemeral mode. Running only in Memory, no Database needed.")
	flag.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	if superverbose {
		log.SetLevel(log.TraceLevel)
	}

	// load config
	var err error
	conf, err = ParseConfig(configPath)
	if err != nil {
		os.Exit(1)
	}

	if !verbose && !superverbose {
		// only set the level from config, if not set by flags
		log.SetLevel(conf.Server.LogLevel)
	}
	log.Debug("Debug log is enabled") // only shows if Debug is actually enabled
	log.Trace("Trace log is enabled") // only shows if Trace is actually enabled

	log.Trace("Trace log is enabled") // only shows if Trace is actually enabled
	log.Tracef("%+v\n", conf)

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

	if !*ephemeral {
		if len(conf.Server.DB.Host) > 0 {
			// connect to DB
			connect()
			log.Traceln("%#v", db)

			if *importData {
				conf.importToDB()
			}
		} else {
			log.Fatal("No database configured. Did you mean to start in ephemeral mode?")
		}
	} else {
		log.Warn("Running in ephemeral-mode. Changes to the config will not persist!!")
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
