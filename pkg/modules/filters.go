package modules

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
)

// list of all filters
var Filters = map[string]func(*ics.Calendar, map[string]string) ([]int, error){
	"regex":      FilterRegex,
	"id":         FilterId,
	"timeframe":  FilterTimeframe,
	"duplicates": FilterDuplicates,
	"all":        FilterAll,
	"duration":   FilterDuration,
}

// This wrappter gets a function from the above filters map and calls it with the parameters and the passed calendar.
// parameters can be any dictionary. The function will then choose how to handle the parameters.
// Returns an array of indices of the events that are filtered out.
func CallFilter(filter func(*ics.Calendar, map[string]string) ([]int, error), cal *ics.Calendar, params map[string]string) ([]int, error) {
	return filter(cal, params)
}

// Filters by regex.
// Params: 'regex'
// 'target' is the property to search in. Default is 'summary'
// Returns the number of added entries. negative, if it removed entries.
func FilterRegex(cal *ics.Calendar, params map[string]string) ([]int, error) {
	var indices []int
	if params["regex"] == "" {
		return indices, fmt.Errorf("missing mandatory Parameter 'regex'")
	}
	if params["target"] == "" {
		params["target"] = "summary"
	}
	regex, _ := regexp.Compile(params["regex"])

	for i, component := range cal.Components { // iterate over events
		switch cal.Components[i].(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)
			var target string

			switch params["target"] {
			case "summary":
				if event.GetProperty(ics.ComponentPropertySummary) != nil {
					target = event.GetProperty(ics.ComponentPropertySummary).Value
				} else {
					continue
				}
			case "description":
				if event.GetProperty(ics.ComponentPropertyDescription) != nil {
					target = event.GetProperty(ics.ComponentPropertyDescription).Value
				} else {
					continue
				}
			case "location":
				if event.GetProperty(ics.ComponentPropertyLocation) != nil {
					target = event.GetProperty(ics.ComponentPropertyLocation).Value
				} else {
					continue
				}
			}

			if regex.MatchString(target) {
				// event matches regex
				indices = append(indices, i)
				log.Debug("Filtering event with id " + event.Id() + "\n")
			}
		default:
			// print type of component
			log.Debug("Unknown component type ignored: " + reflect.TypeOf(cal.Components[i]).String() + "\n")
		}
	}
	return indices, nil
}

// Filters by id.
// Params: 'id'
// An ID is not unique in the calendar. Repeating events can have duplicate IDs.
// Returns an Array of indices of the events that match the id.
func FilterId(cal *ics.Calendar, params map[string]string) ([]int, error) {
	var indices []int
	if params["id"] == "" {
		return indices, fmt.Errorf("missing mandatory Parameter 'id'")
	}

	for i, component := range cal.Components { // iterate over events
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)
			if event.Id() == params["id"] {
				indices = append(indices, i)
				log.Debug("Filter event with id " + params["id"] + " and index " + string(i) + "\n")
			}
		default:
			// print type of component
			log.Debug("Unknown component type ignored: " + reflect.TypeOf(cal.Components[i]).String() + "\n")
		}
	}
	return indices, nil
}

// Filter by timeframe.
// Parameters: either "after" or "before" mandatory
// Format is RFC3339: "2006-01-02T15:04:05Z"
// or "now" for current time
// TODO: implement RRULE compatibility from v1.3.1
func FilterTimeframe(cal *ics.Calendar, params map[string]string) ([]int, error) {
	var indices []int

	// parsing time parameters
	var after time.Time
	var before time.Time
	var err error
	if params["after"] == "" && params["before"] == "" {
		return indices, fmt.Errorf("missing both Parameters 'after' or 'before'. One has to be present")
	}
	if params["after"] == "" {
		log.Debug("No after time given. Using time 0.\n")
		after = time.Time{}
	} else if params["after"] == "now" {
		after = time.Now()
	} else {
		after, err = time.Parse(time.RFC3339, params["after"])
		if err != nil {
			return indices, fmt.Errorf("invalid start time: %s", err.Error())
		}
	}
	if params["before"] == "" {
		log.Debug("No end time given. Using max time\n")
		before = time.Unix(1<<63-1-int64((1969*365+1969/4-1969/100+1969/400)*24*60*60), 999999999)
	} else if params["before"] == "now" {
		before = time.Now()
	} else {
		before, err = time.Parse(time.RFC3339, params["before"])
		if err != nil {
			return indices, fmt.Errorf("invalid end time: %s", err.Error())
		}
	}
	log.Debugf("Filtering events between %s and %s\n", after.Format(time.RFC3339), before.Format(time.RFC3339))

	// actual filtering
	for i, component := range cal.Components { // iterate over events
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)
			if event.GetProperty(ics.ComponentPropertyRrule) != nil {
				// event is a repeating event
				// TODO: get all repeating dates (as its own function) and filter them
				// TODO: add to rruleIndices
				// TODO: break out of switch
			}
			date, _ := event.GetStartAt()
			if date.After(after) && before.After(date) {
				indices = append(indices, i)
				log.Debug("Filtered event with id " + event.Id() + "\n")
			}
		default:
			// print type of component
			log.Debug("Unknown component type ignored: " + reflect.TypeOf(cal.Components[i]).String() + "\n")
		}
	}
	return indices, nil
}

// Looks for duplicate events and returns the indices of duplicate events.
// Only the second and following events are returned, the first is not.
func FilterDuplicates(cal *ics.Calendar, params map[string]string) ([]int, error) {
	var indices []int
	var uniques []string
	for i, component := range cal.Components { // iterate over events
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)
			start, _ := event.GetStartAt()
			end, _ := event.GetEndAt()
			identifier := start.String() + end.String() + event.GetProperty(ics.ComponentPropertySummary).Value
			if helpers.StringInSlice(identifier, uniques) {
				indices = append(indices, i)
				log.Debug("Filter event with id " + event.Id() + "\n")
			} else {
				uniques = append(uniques, identifier)
			}
		default:
			// print type of component
			log.Debug("Unknown component type ignored: " + reflect.TypeOf(cal.Components[i]).String() + "\n")
		}
	}
	return indices, nil
}

// This function filters all events, so returns a list of all indices
func FilterAll(cal *ics.Calendar, params map[string]string) ([]int, error) {
	var indices []int
	for i := range cal.Events() {
		indices = append(indices, i)
	}
	return indices, nil
}

// This function filters events by duration.
// Parameters: "duration" in timeDuration format
// "operator" longer/shorter (default: longer). Longer filters events longer than the given duration.
func FilterDuration(cal *ics.Calendar, params map[string]string) ([]int, error) {
	var indices []int

	// param duration (mandatory)
	if params["duration"] == "" {
		return indices, fmt.Errorf("missing mandatory Parameter 'duration'")
	}
	duration, err := time.ParseDuration(params["duration"])
	if err != nil {
		return indices, fmt.Errorf("invalid duration: %s", err.Error())
	}

	// param operator (optional)
	if params["operator"] == "" {
		params["operator"] = "longer"
	}
	if params["operator"] != "longer" && params["operator"] != "shorter" {
		return indices, fmt.Errorf("invalid operator: %s", params["operator"])
	}

	for i, component := range cal.Components { // iterate over events
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)
			start, _ := event.GetStartAt()
			end, _ := event.GetEndAt()
			if params["operator"] == "longer" {
				if end.Sub(start) >= duration {
					indices = append(indices, i)
					log.Debug("Filtered event with id " + event.Id() + "\n")
				}
			} else if params["operator"] == "shorter" {
				if end.Sub(start) <= duration {
					indices = append(indices, i)
					log.Debug("Filtered event with id " + event.Id() + "\n")
				}
			}
		default:
			// print type of component
			log.Debug("Unknown component type ignored: " + reflect.TypeOf(cal.Components[i]).String() + "\n")
		}
	}
	log.Trace("Duration Filter indices: " + fmt.Sprint(indices) + "\n")
	return indices, nil
}
