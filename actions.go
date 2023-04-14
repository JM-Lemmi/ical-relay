package main

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

// list of all actions
var actions = map[string]func(*ics.Calendar, []int, map[string]string) error{
	"delete":       actionDelete,
	"edit":         actionEdit,
	"add-reminder": actionAddReminder,
}

// This wrappter gets a function from the above action map and calls it with the indices and the passed calendar.
func callAction(action func(*ics.Calendar, []int, map[string]string) error, cal *ics.Calendar, indices []int, params map[string]string) error {
	return action(cal, indices, params)
}

// Deletes events from the calendar.
func actionDelete(cal *ics.Calendar, indices []int, params map[string]string) error {
	// sort indices in descending order, so that we can delete them without messing up the indices
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	log.Trace("Indices to delete: " + fmt.Sprint(indices))
	// delete events
	for _, i := range indices {
		cal.Components = removeFromICS(cal.Components, i)
	}
	return nil
}

// Edits events in the calendar.
// Params: 'overwrite': 'true' (default), 'false', 'fillempty'
// 'new-summary', 'new-description', 'new-location', 'new-start', 'new-end'
func actionEdit(cal *ics.Calendar, indices []int, params map[string]string) error {
	// param defaults
	if params["overwrite"] == "" {
		params["overwrite"] = "true"
	}

	// edit events
	for _, i := range indices {
		switch cal.Components[i].(type) {
		case *ics.VEvent:
			event := cal.Components[i].(*ics.VEvent)

			log.Debug("Changing event with id " + event.Id())
			if params["new-summary"] != "" {
				if event.GetProperty(ics.ComponentPropertySummary) == nil {
					params["overwrite"] = "true"
					// if the summary is not set, we need to create it
				}
				switch params["overwrite"] {
				case "false":
					event.SetProperty(ics.ComponentPropertySummary, event.GetProperty(ics.ComponentPropertySummary).Value+"; "+params["new-summary"])
				case "fillempty":
					if event.GetProperty(ics.ComponentPropertySummary).Value == "" {
						event.SetProperty(ics.ComponentPropertySummary, params["new-summary"])
					}
				case "true":
					event.SetProperty(ics.ComponentPropertySummary, params["new-summary"])
				}
				log.Debug("Changed summary to " + event.GetProperty(ics.ComponentPropertySummary).Value)
			}
			if params["new-description"] != "" {
				if event.GetProperty(ics.ComponentPropertyDescription) == nil {
					params["overwrite"] = "true"
					// if the description is not set, we need to create it
				}
				switch params["overwrite"] {
				case "false":
					event.SetProperty(ics.ComponentPropertyDescription, event.GetProperty(ics.ComponentPropertyDescription).Value+"; "+params["new-description"])
				case "fillempty":
					if event.GetProperty(ics.ComponentPropertyDescription).Value == "" {
						event.SetProperty(ics.ComponentPropertyDescription, params["new-description"])
					}
				case "true":
					event.SetProperty(ics.ComponentPropertyDescription, params["new-description"])
				}
				log.Debug("Changed description to " + event.GetProperty(ics.ComponentPropertyDescription).Value)
			}
			if params["new-location"] != "" {
				if event.GetProperty(ics.ComponentPropertyLocation) == nil {
					params["overwrite"] = "true"
					// if the description is not set, we need to create it
				}
				switch params["overwrite"] {
				case "false":
					event.SetProperty(ics.ComponentPropertyLocation, event.GetProperty(ics.ComponentPropertyLocation).Value+"; "+params["new-location"])
				case "fillempty":
					if event.GetProperty(ics.ComponentPropertyLocation).Value == "" {
						event.SetProperty(ics.ComponentPropertyLocation, params["new-location"])
					}
				case "true":
					event.SetProperty(ics.ComponentPropertyLocation, params["new-location"])
				}
				log.Debug("Changed location to " + event.GetProperty(ics.ComponentPropertyLocation).Value)
			}
			if params["new-start"] != "" {
				start, err := time.Parse(time.RFC3339, params["new-start"])
				if err != nil {
					return fmt.Errorf("invalid start time: %s", err.Error())
				}
				event.SetStartAt(start)
				log.Debug("Changed start to " + params["new-start"])
			}
			if params["new-end"] != "" {
				end, err := time.Parse(time.RFC3339, params["new-end"])
				if err != nil {
					return fmt.Errorf("invalid end time: %s", err.Error())
				}
				event.SetEndAt(end)
				log.Debug("Changed end to " + params["new-end"])
			}
			// adding edited event back to calendar
			cal.Components[i] = event
			return nil

		default:
			// print type of component
			log.Debug("Unknown component type ignored: " + reflect.TypeOf(cal.Components[i]).String() + "\n")
		}
	}
	return nil
}

func actionAddReminder(cal *ics.Calendar, indices []int, params map[string]string) error {
	// add reminder to events
	for _, i := range indices {
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
