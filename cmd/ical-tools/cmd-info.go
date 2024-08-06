package main

import (
	"fmt"

	"github.com/jm-lemmi/ical-relay/helpers"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

type eventinfoCmd struct {
	Source   string `arg:"positional" help:"Source calendar to get event info from (include protocol, e.g. file:// or http://)"`
	Id       string `arg:"--id" help:"Get events by ID"`
	Summary  string `arg:"--summary" help:"Get events by summary"`
	Location string `arg:"--location" help:"Get events by location"`
}

func cmdEventinfo(eventinfoArgs eventinfoCmd) {
	log.Debugf("Getting event info for event in calendar %s", eventinfoArgs.Source)
	log.Trace("Event info args: ", eventinfoArgs)

	calendar, err := getSource(eventinfoArgs.Source)
	if err != nil {
		log.Fatal(err)
	}

	found := false

	// TODO: use module filters instead
	if eventinfoArgs.Id != "" {
		for _, component := range calendar.Components {
			if event, ok := component.(*ics.VEvent); ok {
				if event.Id() == eventinfoArgs.Id {
					fmt.Println(helpers.PrettyPrint(*event))
					found = true
				}
			}
		}
	}
	if eventinfoArgs.Summary != "" {
		for _, component := range calendar.Components {
			if event, ok := component.(*ics.VEvent); ok {
				if event.GetSummary() == eventinfoArgs.Summary {
					fmt.Println(helpers.PrettyPrint(*event))
					found = true
				}
			}
		}
	}
	if eventinfoArgs.Location != "" {
		for _, component := range calendar.Components {
			if event, ok := component.(*ics.VEvent); ok {
				if event.GetLocation() == eventinfoArgs.Location {
					fmt.Println(helpers.PrettyPrint(*event))
					found = true
				}
			}
		}
	}

	if !found {
		log.Fatal("Event not found")
	}

}
