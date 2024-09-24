package modules

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"

	"github.com/jm-lemmi/ical-relay/helpers"
)

// list of all actions
var Actions = map[string]func(*ics.Calendar, []int, map[string]string) error{
	"delete":       ActionDelete,
	"edit":         ActionEdit,
	"add-reminder": ActionAddReminder,
	"strip-info":   ActionStripInfo,
}

// This wrappter gets a function from the above action map and calls it with the indices and the passed calendar.
func CallAction(action func(*ics.Calendar, []int, map[string]string) error, cal *ics.Calendar, indices []int, params map[string]string) error {
	return action(cal, indices, params)
}

// Deletes events from the calendar.
func ActionDelete(cal *ics.Calendar, indices []int, params map[string]string) error {
	// sort indices in descending order, so that we can delete them without messing up the indices
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	log.Trace("Indices to delete: " + fmt.Sprint(indices))
	// delete events
	for _, i := range indices {
		cal.Components = helpers.RemoveFromICS(cal.Components, i)
	}
	return nil
}

// Edits events in the calendar.
// Params: 'overwrite': 'true' (default), 'false', 'fillempty'
// 'new-summary', 'new-description', 'new-location', 'new-start', 'new-end'
func ActionEdit(cal *ics.Calendar, indices []int, params map[string]string) error {
	// param defaults
	if params["overwrite"] == "" {
		params["overwrite"] = "true"
	}

	// parse move-time
	// should be ok to leave at empty string check
	if params["move-time"] != "" && (params["new-start"] != "" || params["new-end"] != "") {
		return fmt.Errorf("two exclusive params were given: 'move-time' and 'new-start'/'new-end'")
	}

	// edit events
	for _, i := range indices {
		switch cal.Components[i].(type) {
		case *ics.VEvent:
			event := cal.Components[i].(*ics.VEvent)

			log.Debug("Changing event with id " + event.Id())
			if _, ok := params["new-summary"]; ok {
				if event.GetProperty(ics.ComponentPropertySummary) == nil {
					params["overwrite"] = "true"
					// if the summary is not set, we need to create it
				}
				switch params["overwrite"] {
				case "false":
					event.SetSummary(event.GetSummary() + "; " + params["new-summary"])
				case "fillempty":
					if event.GetProperty(ics.ComponentPropertySummary).Value == "" {
						event.SetSummary(params["new-summary"])
					}
				case "true":
					event.SetSummary(params["new-summary"])
				}
				log.Debug("Changed summary to " + event.GetProperty(ics.ComponentPropertySummary).Value)
			}
			if _, ok := params["new-description"]; ok {
				if event.GetProperty(ics.ComponentPropertyDescription) == nil {
					params["overwrite"] = "true"
					// if the description is not set, we need to create it
				}
				switch params["overwrite"] {
				case "false":
					event.SetDescription(event.GetDescription() + "; " + params["new-description"])
				case "fillempty":
					if event.GetProperty(ics.ComponentPropertyDescription).Value == "" {
						event.SetDescription(params["new-description"])
					}
				case "true":
					event.SetDescription(params["new-description"])
				}
				log.Debug("Changed description to " + event.GetProperty(ics.ComponentPropertyDescription).Value)
			}
			if _, ok := params["new-location"]; ok {
				if event.GetProperty(ics.ComponentPropertyLocation) == nil {
					params["overwrite"] = "true"
					// if the description is not set, we need to create it
				}
				switch params["overwrite"] {
				case "false":
					event.SetLocation(event.GetLocation() + "; " + params["new-location"])
				case "fillempty":
					if event.GetProperty(ics.ComponentPropertyLocation).Value == "" {
						event.SetLocation(params["new-location"])
					}
				case "true":
					event.SetLocation(params["new-location"])
				}
				log.Debug("Changed location to " + event.GetProperty(ics.ComponentPropertyLocation).Value)
			}
			if _, ok := params["new-start"]; ok {
				start, err := time.Parse(time.RFC3339, params["new-start"])
				if err != nil {
					return fmt.Errorf("invalid start time: %s", err.Error())
				}
				event.SetStartAt(start)
				log.Debug("Changed start to " + params["new-start"])
			}
			if _, ok := params["new-end"]; ok {
				end, err := time.Parse(time.RFC3339, params["new-end"])
				if err != nil {
					return fmt.Errorf("invalid end time: %s", err.Error())
				}
				event.SetEndAt(end)
				log.Debug("Changed end to " + params["new-end"])
			}

			if _, ok := params["move-time"]; ok {
				dur, err := time.ParseDuration(params["move-time"])
				if err != nil {
					return fmt.Errorf("invalid duration: %s", err.Error())
				}
				start, _ := event.GetStartAt()
				log.Debug("Starttime is " + start.String())
				end, _ := event.GetEndAt()
				event.SetStartAt(start.Add(dur))
				log.Debug("Changed start to " + start.Add(dur).String())
				event.SetEndAt(end.Add(dur))
				log.Debug("Changed start and end by " + dur.String())
			}
			// adding edited event back to calendar
			cal.Components[i] = event

		default:
			// print type of component
			log.Debug("Unknown component type ignored: " + reflect.TypeOf(cal.Components[i]).String() + "\n")
		}
	}
	return nil
}

func ActionAddReminder(cal *ics.Calendar, indices []int, params map[string]string) error {
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

// Strips information from events, similar to Outlooks export feature.
// Params: 'mode': 'availibility' (only show freebusy availibility), 'limited' (show only summary)
func ActionStripInfo(cal *ics.Calendar, indices []int, params map[string]string) error {
	// strip info from events
	for _, i := range indices {
		switch cal.Components[i].(type) {
		case *ics.VEvent:
			event := cal.Components[i].(*ics.VEvent)
			// create new event and copy only the needed data over
			var newevent *ics.VEvent

			switch params["mode"] {
			case "availibility":
				// copies: start, end, uid and freebusy status
				newevent = ics.NewEvent(event.GetProperty(ics.ComponentPropertyUniqueId).Value)
				start, _ := event.GetStartAt()
				end, _ := event.GetEndAt()
				newevent.SetStartAt(start)
				newevent.SetEndAt(end)
				if event.GetProperty(ics.ComponentPropertyFreebusy) == nil && event.GetProperty(ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS")) == nil {
					// no freebusy or MS-freebusy status set, assume busy
					newevent.AddProperty(ics.ComponentPropertySummary, "Busy")
				} else if event.GetProperty(ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS")) != nil {
					// MS-freebusy status set
					newevent.AddProperty(ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS"), event.GetProperty(ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS")).Value)
					newevent.AddProperty(ics.ComponentPropertySummary, event.GetProperty(ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS")).Value)
				} else {
					// freebusy status set
					newevent.AddProperty(ics.ComponentPropertyFreebusy, event.GetProperty(ics.ComponentPropertyFreebusy).Value)
					newevent.AddProperty(ics.ComponentPropertySummary, event.GetProperty(ics.ComponentPropertyFreebusy).Value)
				}

			case "limited":
				// copies: summary, start, end, uid and freebusy status
				newevent = ics.NewEvent(event.GetProperty(ics.ComponentPropertyUniqueId).Value)
				newevent.AddProperty(ics.ComponentPropertySummary, event.GetProperty(ics.ComponentPropertySummary).Value)
				start, _ := event.GetStartAt()
				end, _ := event.GetEndAt()
				newevent.SetStartAt(start)
				newevent.SetEndAt(end)
				if event.GetProperty(ics.ComponentPropertyFreebusy) == nil && event.GetProperty(ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS")) == nil {
					// nothing happens here, we don't want to add a freebusy status
				} else if event.GetProperty(ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS")) != nil {
					// MS-freebusy status set
					newevent.AddProperty(ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS"), event.GetProperty(ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS")).Value)
				} else {
					// freebusy status set
					newevent.AddProperty(ics.ComponentPropertyFreebusy, event.GetProperty(ics.ComponentPropertyFreebusy).Value)
				}

			default:
				return fmt.Errorf("invalid mode: %s", params["mode"])
			}

			cal.Components[i] = newevent
			log.Debug("Stripped info with mode " + params["mode"] + " from event " + event.Id())
		}
	}
	return nil
}

// Fixes calendars where the timezone is set by TIMEZONE or X-WR-TIMEZONE property once instead of VTIMEZONE
// adds a correct VTIMEZONE to the calendar.
func ActionXWRTimezoneToVTimezone(cal *ics.Calendar) error {
	var property *ics.CalendarProperty
	for _, prop := range cal.CalendarProperties {
		if prop.IANAToken == "TIMEZONE" || prop.IANAToken == "X-WR-TIMEZONE" {
			property = &prop
			break
		}
	}
	// no default timezone for the calendar found
	if property == nil {
		return fmt.Errorf("no timezone from TIMEZONE or X-WR-TIMEZONE found in calendar")
	}

	// load VTIMEZONE
	tz, err := helpers.GetVTimezoneFromString(property.Value)
	if err != nil {
		return err
	}

	// add VTIMEZONE to calendar
	cal.AddVTimezone(&tz)

	// add TZID to all events, that don't have it yet
	for _, event := range cal.Events() {
		for _, prop := range event.Properties {
			if prop.IANAToken == "DTSTART" || prop.IANAToken == "DTEND" {
				// set timezone only if no timezone is already set
				if _, ok := prop.ICalParameters["TZID"]; !ok {
					prop.ICalParameters["TZID"] = []string{property.Value}
				}
			}
		}
	}

	return nil
}
