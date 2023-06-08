package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/alexflint/go-arg"
	ics "github.com/arran4/golang-ical"
	"github.com/jm-lemmi/ical-relay/compare"
	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
)

var version = "2.0.0-beta.6.3"

func main() {
	log.Info("Welcome to ical-tools, version " + version)

	// CLI Flags
	// the type xxCmd is used to define subcommands
	type CompareCmd struct {
		Base     string `arg:"positional" help:"Base calendar to compare against (include protocol, e.g. file:// or http://)"`
		Incoming string `arg:"positional" help:"Incoming calendar to compare against (include protocol, e.g. file:// or http://)"`
	}
	var args struct {
		Compare      *CompareCmd `arg:"subcommand:compare" help:"Compare two ical files"`
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
		log.Debug("Comparing ical files %s and %s", args.Compare.Base, args.Compare.Incoming)

		base, err := getSource(args.Compare.Base)
		if err != nil {
			log.Fatal(err)
		}
		incoming, err := getSource(args.Compare.Incoming)
		if err != nil {
			log.Fatal(err)
		}

		added, deleted, changed := compare.Compare(base, incoming)

		if len(added)+len(deleted)+len(changed) == 0 {
			log.Info("No changes detected.")
		} else {
			log.Debug("Changes detected: " + fmt.Sprint(len(added)) + " added, " + fmt.Sprint(len(deleted)) + " deleted, " + fmt.Sprint(len(changed)) + " changed")

			var body string

			if len(added) > 0 {
				for _, event := range added {
					body += "Added:\n\n" + helpers.PrettyPrint(event) + "\n\n"
				}
			}
			if len(deleted) > 0 {
				for _, event := range deleted {
					body += "Deleted:\n\n" + helpers.PrettyPrint(event) + "\n\n"
				}
			}
			if len(changed) > 0 {
				for _, event := range changed {
					body += "Changed (displaying new version):\n\n" + helpers.PrettyPrint(event) + "\n\n"
				}
			}

			fmt.Println(body)
		}
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
