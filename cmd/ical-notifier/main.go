package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
	ics "github.com/arran4/golang-ical"
	"github.com/jm-lemmi/ical-relay/datastore"
	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
)

//go:generate ../../.github/scripts/generate-version.sh
//go:embed VERSION
var version string // If you are here due to a compile error, run go generate

var configPath string
var conf Config
var dataStore datastore.DataStore

func main() {
	log.Info("Welcome to ical-notifier, version " + version)

	// CLI Flags
	var args struct {
		Notifier     string `help:"Run notifier with given ID"`
		Verbose      bool   `arg:"-v,--verbose" help:"verbosity level Debug"`
		Superverbose bool   `arg:"--superverbose" help:"verbosity level Trace"`
		ConfigPath   string `arg:"-c,--config" help:"Configuration path" default:"config.yml"`
		// TODO: add data path
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
		log.SetLevel(conf.General.LogLevel)
	}

	log.Debug("Debug log is enabled") // only shows if Debug is actually enabled
	log.Trace("Trace log is enabled") // only shows if Trace is actually enabled

	log.Tracef("%+v\n", conf)

	if !helpers.DirectoryExists(conf.General.StoragePath + "notifystore/") {
		log.Info("Creating notifystore directory")
		err = os.MkdirAll(conf.General.StoragePath+"notifystore/", 0750)
		if err != nil {
			log.Fatalf("Error creating notifystore: %v", err)
		}
	}

	// load data

	if !conf.General.LiteMode {
		// RUNNING FULL MODE
		log.Debug("Running in full mode.")
		if conf.General.DB.Host == "" {
			log.Fatal("DB configuration missing")
		}

		// connect to DB
		datastore.Connect(conf.General.DB.User, conf.General.DB.Password, conf.General.DB.Host, conf.General.DB.DbName)
		dataStore = datastore.DatabaseDataStore{}

	} else {
		log.Warn("Running in lite mode. No connection with ical-relay assumed. No dynamic unsubscription possible.")
		// TODO: add dynamic data path
		dataStore, err = datastore.ParseDataFile(conf.General.StoragePath + "data.yml")
		if err != nil {
			log.Fatalf("Error loading data file: %v", err)
		}
	}

	// APPLICATION LOGIC
	// get all notifiers to iterate

	for notifierName, notifier := range dataStore.GetNotifiers() {
		err = notifyChanges(notifierName, notifier)
		if err != nil {
			log.Error("Failed to run notifier ", notifierName, ": ", err)
		} else {
			log.Info("Notifier ", notifierName, " ran successfully")
		}
	}
}

// direct copy from ../ical-relay/profiles.go
// but with switch option "profile" removed
func getSource(source string) (*ics.Calendar, error) {
	var calendar *ics.Calendar
	var err error

	switch strings.Split(source, "://")[0] {
	case "http", "https":
		calendar, err = helpers.ReadCalURL(source)
		if err != nil {
			return nil, err
		}
	case "file":
		calendar, err = helpers.LoadCalFile(strings.Split(source, "://")[1])
		if err != nil {
			return nil, err
		}
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(strings.Split(source, "://")[1])
		if err != nil {
			return nil, err
		}

		calendar, err = ics.ParseCalendar(bytes.NewReader(decoded))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown source type '%s'", strings.Split(source, "://")[0])
	}
	return calendar, nil
}
