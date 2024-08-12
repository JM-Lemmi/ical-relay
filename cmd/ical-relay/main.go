package main

import (
	_ "embed"

	"html/template"
	"net/http"
	"os"

	"github.com/jm-lemmi/ical-relay/database"
	"github.com/jm-lemmi/ical-relay/helpers"

	"github.com/alexflint/go-arg"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

//go:generate ../../.github/scripts/generate-version.sh
//go:embed VERSION
var version string // If you are here due to a compile error, run go generate

var configPath string
var conf Config
var dataStore database.DataStore

var router *mux.Router

func main() {
	log.Info("Welcome to ical-relay, version " + version)

	// CLI Flags
	var args struct {
		Notifier     string `help:"Run notifier with given ID"`
		ConfigPath   string `arg:"--config" help:"Configuration path" default:"config.yml"`
		Verbose      bool   `arg:"-v,--verbose" help:"verbosity level Debug"`
		SuperVerbose bool   `arg:"--superverbose" help:"verbosity level Trace"`
		ImportData   bool   `arg:"--import-data" help:"Import Data from Config into DB"`
	}
	arg.MustParse(&args)

	configPath = args.ConfigPath

	if args.Verbose {
		log.SetLevel(log.DebugLevel)
	}
	if args.SuperVerbose {
		log.SetLevel(log.TraceLevel)
	}

	// load config
	var err error
	conf, err = ParseConfig(configPath)
	if err != nil {
		os.Exit(1)
	}

	if !args.Verbose && !args.SuperVerbose {
		// only set the level from config, if not set by flags
		log.SetLevel(conf.Server.LogLevel)
	}
	log.Debug("Debug log is enabled") // only shows if Debug is actually enabled
	log.Trace("Trace log is enabled") // only shows if Trace is actually enabled

	log.Tracef("%+v\n", conf)

	if !helpers.DirectoryExists(conf.Server.StoragePath + "notifystore/") {
		log.Info("Creating notifystore directory")
		err = os.MkdirAll(conf.Server.StoragePath+"notifystore/", 0750)
		if err != nil {
			log.Fatalf("Error creating notifystore: %v", err)
		}
	}
	if !helpers.DirectoryExists(conf.Server.StoragePath + "calstore/") {
		log.Info("Creating calstore directory")
		err = os.MkdirAll(conf.Server.StoragePath+"calstore/", 0750)
		if err != nil {
			log.Fatalf("Error creating calstore: %v", err)
		}
	}

	// run notifier if specified
	if args.Notifier != "" {
		log.Debug("Notifier mode called. Running: " + args.Notifier)
		err := RunNotifier(args.Notifier)
		if err != nil {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	} else {
		log.Debug("Server mode.")
	}

	// setup router. Will be configured depending on FULL or LITE mode
	router = mux.NewRouter()

	if !conf.Server.LiteMode {
		// RUNNING FULL MODE
		log.Debug("Running in full mode.")
		if conf.Server.DB.Host == "" {
			log.Fatal("DB configuration missing")
		}

		// connect to DB
		database.Connect(conf.Server.DB.User, conf.Server.DB.Password, conf.Server.DB.Host, conf.Server.DB.DbName)
		log.Tracef("%#v", database.Db)
		dataStore = database.DatabaseDataStore{}

		if args.ImportData {
			conf.importToDB()
		}

		// setup routes
		initHandlersProfile()
		initHandlersApi()

		if !conf.Server.DisableFrontend {
			htmlTemplates = template.Must(template.ParseGlob(conf.Server.TemplatePath + "*.html")) // TODO: fail more gracefully than segfault

			initHandlersFrontend()
		}

		// start cleanup
		CleanupStartup()
	} else {
		log.Warn("Running in lite mode. No changes will be saved.")
		dataStore = conf

		// setup routes
		initHandlersProfile()
	}

	// start server
	address := conf.Server.Addr
	log.Info("Starting server on " + address)
	log.Fatal(http.ListenAndServe(address, router))
}
