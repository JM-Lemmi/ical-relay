package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexflint/go-arg"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var version = "2.0.0-beta.6.3"

var configPath string
var conf Config

var router *mux.Router

func main() {
	log.Info("Welcome to ical-relay, version " + version)

	// CLI Flags
	var args struct {
		Notifier     string `help:"Run notifier with given ID"`
		ConfigPath   string `arg:"--config" help:"Configuration path" default:"config.yml"`
		Verbose      bool   `arg:"-v,--verbose" help:"verbosity level Debug"`
		Superverbose bool   `arg:"--superverbose" help:"verbosity level Trace"`
		ImportData   bool   `help:"Import Data from Config into DB"`
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

	if len(conf.Server.DB.Host) > 0 {
		// connect to DB
		if conf.Server.DB.Host == "Special:EMBEDDED" {
			log.Info("Starting embedded postgres server (this will take a while on the first run)...")
			if conf.Server.DB.User == "" {
				conf.Server.DB.User = "postgres"
			}
			if conf.Server.DB.Password == "" {
				conf.Server.DB.Password = "postgres"
			}
			postgres := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
				Username(conf.Server.DB.User).
				Password(conf.Server.DB.Password).
				Database(conf.Server.DB.DbName).
				Version(embeddedpostgres.V15).
				Logger(log.StandardLogger().Writer()).
				BinariesPath(conf.Server.StoragePath + "db/runtime").
				DataPath(conf.Server.StoragePath + "db/data").
				Locale("C").
				Port(5432)) //todo: support non default port
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
			go func() {
				sigs := <-sigs
				log.Info("Caught ", sigs)
				err := postgres.Stop()
				if err != nil {
					log.Fatal("Could not properly shutdown embedded postgres server: ", err)
				}
				os.Exit(0)
			}()
			err := postgres.Start()
			if err != nil {
				log.Fatal("Could not start embedded postgres server: ", err)
			}
			conf.Server.DB.Host = "localhost"
		}
		connect()
		fmt.Printf("%#v\n", db)

		if args.ImportData {
			conf.importToDB()
		}
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
