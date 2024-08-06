package main

import (
	"fmt"

	"github.com/jm-lemmi/ical-relay/compare"
	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
)

type compareCmd struct {
	Base     string `arg:"positional" help:"Base calendar to compare against (include protocol, e.g. file:// or http://)"`
	Incoming string `arg:"positional" help:"Incoming calendar to compare against (include protocol, e.g. file:// or http://)"`
}

func cmdCompare(compareArgs compareCmd) {
	log.Debugf("Comparing ical files %s and %s", compareArgs.Base, compareArgs.Incoming)

	base, err := getSource(compareArgs.Base)
	if err != nil {
		log.Fatal(err)
	}
	incoming, err := getSource(compareArgs.Incoming)
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
