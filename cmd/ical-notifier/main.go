package main

import (
	_ "embed"

	"github.com/alexflint/go-arg"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

var version = "2.0.0-beta.7.3"

var db sqlx.DB

func main() {
	log.Info("Welcome to ical-notifier, version " + version)

	// CLI Flags
	var args struct {
		Notifier     string `help:"Run notifier with given ID"`
		Verbose      bool   `arg:"-v,--verbose" help:"verbosity level Debug"`
		Superverbose bool   `arg:"--superverbose" help:"verbosity level Trace"`
		Database     string `arg:"-d,--database" arg:"required" help:"Database connection string"` // postgresql://ical_relay:password@localhost:5234/ical_relay
	}
	arg.MustParse(&args)

	if args.Verbose {
		log.SetLevel(log.DebugLevel)
	}
	if args.Superverbose {
		log.SetLevel(log.TraceLevel)
	}

	log.Debug("Debug log is enabled") // only shows if Debug is actually enabled
	log.Trace("Trace log is enabled") // only shows if Trace is actually enabled

	// connect to DB
	dbConn, err := sqlx.Connect("postgres", args.Database)
	if err != nil {
		log.Fatalf("Connection to db failed: %s", err)
		panic(err)
	}
	log.Debug("Connected to db")
	db = *dbConn
	log.Tracef("%#v", db)

	// check db version / init tables
	// TODO: if we share database with ical_relay there will be a version but we still want to init the database
	var dbVersion int
	err = db.Get(&dbVersion, `SELECT MAX(version) FROM schema_upgrades`)
	if err != nil {
		log.Info("Initially creating tables...")
		initTables()
	}

	// TODO

	// DETECTION
	// get all notifiers to iterate
	// get source
	// compare to history on file
	// write output to db

	// NOTIFY
	// iterate over all notifiers
	// trigger notifier code (send mail, write rss file)
}

//go:embed db.sql
var dbInitScript string

func initTables() {
	_, err := db.Exec(dbInitScript)
	if err != nil {
		log.Panic("Failed to execute db init script", err)
	}
}
