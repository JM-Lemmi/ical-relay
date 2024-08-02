package main

import (
	_ "embed"

	"html/template"
	"net/http"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

//go:generate ../../.github/scripts/generate-version.sh
//go:embed VERSION
var version string // If you are here due to a compile error, run go generate

var configPath string
var conf Config

var router *mux.Router

func main() {
	log.Info("Welcome to ical-relay, version " + version)

	// CLI Flags
	var args struct {
		Notifier     string `help:"Run notifier with given ID"`
		ConfigPath   string `arg:"-c,--config" help:"Configuration path" default:"config.yml"`
		Verbose      bool   `arg:"-v,--verbose" help:"verbosity level Debug"`
		Superverbose bool   `arg:"--superverbose" help:"verbosity level Trace"`
		ImportData   bool   `arg:"--import-data" help:"Import Data from Config into DB"`
		Ephemeral    bool   `arg:"-e" help:"Enable ephemeral mode. Running only in Memory, no Database needed."`
	}
	arg.MustParse(&args)

	configPath = args.ConfigPath

	if args.Verbose {
		log.SetLevel(log.DebugLevel)
	}
	if args.Superverbose {
		log.SetLevel(log.TraceLevel)
	}

	// load config
	var err error
	conf, err = ParseConfig(configPath)
	if err != nil {
		os.Exit(1)
	}

	if !args.Verbose && !args.Superverbose {
		// only set the level from config, if not set by flags
		log.SetLevel(conf.Server.LogLevel)
	}
	log.Debug("Debug log is enabled") // only shows if Debug is actually enabled
	log.Trace("Trace log is enabled") // only shows if Trace is actually enabled

	log.Tracef("%+v\n", conf)

	if !args.Ephemeral {
		if len(conf.Server.DB.Host) > 0 {
			// connect to DB
			connect()
			log.Tracef("%#v", db)

			if args.ImportData {
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

	// start cleanup
	CleanupStartup()

	// start server
	address := conf.Server.Addr
	log.Info("Starting server on " + address)
	log.Fatal(http.ListenAndServe(address, router))
}
