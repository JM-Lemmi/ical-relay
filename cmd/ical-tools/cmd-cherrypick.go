package main

import (
	"fmt"

	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
)

type cherrypickCmd struct {
	Base string `arg:"positional" help:"Base calendar to cherrypick from (include protocol, e.g. file:// or http://)"`
	Id   string `arg:"positional" help:"ID of the event to cherrypick"`
}

func cmdCherrypick(cherrypickArgs cherrypickCmd) {
	log.Debugf("Cherrypicking event %s from %s", cherrypickArgs.Id, cherrypickArgs.Base)

	base, err := getSource(cherrypickArgs.Base)
	if err != nil {
		log.Fatal(err)
	}

	event, err := helpers.GetEventFromCalByID(base, cherrypickArgs.Id)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(event.Serialize())
}
