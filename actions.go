package main

import (
	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

// list of all actions
var actions = map[string]func(*ics.Calendar, []int) error{
	"delete":       actionDelete,
	"edit":         actionEdit,
	"add-reminder": actionAddReminder,
}

// This wrappter gets a function from the above action map and calls it with the indices and the passed calendar.
func callAction(action func(*ics.Calendar, []int) error, indices []int, cal *ics.Calendar) error {
	return action(cal, indices)
}

func actionAddReminder(cal *ics.Calendar, indices []int) error {
	// add reminder to events
	for i := range indices {
		switch cal.Components[i].(type) {
		case *ics.VEvent:
			event := cal.Components[i].(*ics.VEvent)
			event.AddAlarm()
			event.Alarms()[0].SetTrigger(("-PT" + params["time"]))
			event.Alarms()[0].SetAction("DISPLAY")
			cal.Components[i] = event
			log.Debug("Added reminder to event " + event.Id())
		}
	}
	return nil
}
