package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/alexflint/go-arg"
	ics "github.com/arran4/golang-ical"
	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
)

var version = "2.0.0-beta.6.3"

func main() {
	log.Info("Welcome to ical-tools, version " + version)

	// CLI Flags
	// subcommands are defined in the respective files
	var args struct {
		Compare      *compareCmd `arg:"subcommand:compare" help:"Compare two ical files"`
		Verbose      bool        `arg:"-v,--verbose" help:"verbosity level Debug"`
		Superverbose bool        `arg:"--superverbose" help:"verbosity level Trace"`
	}
	arg.MustParse(&args)

	// set log level
	if args.Verbose {
		log.SetLevel(log.DebugLevel)
	}
	if args.Superverbose {
		log.SetLevel(log.TraceLevel)
	}

	// starting subcommands
	switch {
	case args.Compare != nil:
		cmdCompare(*args.Compare)
	}
}

func getSource(source string) (*ics.Calendar, error) {
	var calendar *ics.Calendar
	var err error

	switch strings.Split(source, "://")[0] {
	case "http", "https":
		response, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP error: %s", response.Status)
		}
		if err != nil {
			return nil, err
		}
		calendar, err = ics.ParseCalendar(response.Body)
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
