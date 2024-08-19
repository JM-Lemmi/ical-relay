package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	ics "github.com/arran4/golang-ical"
	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
)

type addCmd struct {
	Event string `arg:"positional" help:"Event to add to the calendar (- for STDIN, include protocol, e.g. file:// or http://)"`
	Base  string `arg:"positional" help:"Base calendar to add to (include protocol, e.g. file://)"`
}

func cmdAdd(addArgs addCmd) {
	log.Debugf("Adding event %s to %s", addArgs.Event, addArgs.Base)

	var add *ics.Calendar
	var err error
	if addArgs.Event == "-" {
		reader := bufio.NewReader(os.Stdin)
		add, err = ics.ParseCalendar(reader)
		if err == errors.New("malformed calendar; expected a vcalendar") {
			// probably a single VEVENT
			// TODO: do better parsing
			log.Fatal(fmt.Errorf("error reading input. Single Events are not supported yet"))
		} else if err != nil {
			err = fmt.Errorf("error reading input event: %s", err)
			log.Fatal(err)
		}
	} else {
		add, err = getSource(addArgs.Event)
		if err != nil {
			log.Fatal(err)
		}
	}

	base, err := getSource(addArgs.Base)
	if err != nil {
		log.Fatal(err)
	}

	helpers.AddEvents(base, add)

	fmt.Println(base.Serialize())
}
