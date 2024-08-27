package main

import (
	_ "embed"
	"strconv"

	"html/template"
	"net/http"
	"os"

	"github.com/jm-lemmi/ical-relay/datastore"
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
var dataStore datastore.DataStore

var router *mux.Router

func main() {
	log.Info("Welcome to ical-relay, version " + version)

	// CLI Flags
	var args struct {
		Notifier   string `help:"Run notifier with given ID"`
		ConfigPath string `arg:"-c,--config" help:"Configuration path" default:"config.yml"`
		// TODO: add data path
		Verbose      bool   `arg:"-v,--verbose" help:"verbosity level Debug"`
		SuperVerbose bool   `arg:"--superverbose" help:"verbosity level Trace"`
		ImportData   string `arg:"--import-data" help:"Import Data from Data.yml into DB"`
		DisableTele  bool   `arg:"--disable-telemetry" help:"Disables reporting its own existence"`
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

	if !helpers.DirectoryExists(conf.Server.StoragePath + "calstore/") {
		log.Info("Creating calstore directory")
		err = os.MkdirAll(conf.Server.StoragePath+"calstore/", 0750)
		if err != nil {
			log.Fatalf("Error creating calstore: %v", err)
		}
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
		datastore.Connect(conf.Server.DB.User, conf.Server.DB.Password, conf.Server.DB.Host, conf.Server.DB.DbName)
		dataStore = datastore.DatabaseDataStore{}

		if args.ImportData != "" {
			err := datastore.ImportToDB(args.ImportData) // TODO
			if err != nil {
				log.Fatalf("Error importing data: %v", err)
			}
		}

		// setup routes
		initHandlersProfile()
		initHandlersApi()

		if !conf.Server.DisableFrontend {
			htmlTemplates = template.Must(template.ParseGlob(conf.Server.TemplatePath + "*.html")) // TODO: fail more gracefully than segfault

			initHandlersFrontend()
		}
	} else {
		log.Warn("Running in lite mode. No changes will be saved.")
		dataStore, err = datastore.ParseDataFile(conf.Server.StoragePath + "data.yml")
		if err != nil {
			log.Fatalf("Error loading data file: %v", err)
		}

		// setup routes
		initHandlersProfile()
	}

	// Telemetry
	if !args.DisableTele {
		// in own thread, to avoid hanging up the startup, if telemetry fails for some reason
		go func() {
			_, err := http.Get("https://ical-relay.telemetry.julian-lemmerich.de/ping?name=" + helpers.GetMD5Hash(conf.Server.Name+conf.Server.URL) + "&litemode=" + strconv.FormatBool(conf.Server.LiteMode) + "&version=" + version)
			if err == nil {
				log.Trace("Sent telemetry successfully")
			} else {
				log.Tracef("Sending telemetry failed: %s", err)
			}
		}()
	}

	// start server
	address := conf.Server.Addr
	log.Info("Starting server on " + address)
	log.Fatal(http.ListenAndServe(address, router))
}
