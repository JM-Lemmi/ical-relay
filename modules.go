package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

var modules = map[string]func(*ics.Calendar, map[string]string) (int, error){
	"delete-bysummary-regex": moduleDeleteSummaryRegex,
	"delete-byid":            moduleDeleteId,
	"add-url":                moduleAddURL,
	"add-file":               moduleAddFile,
	"delete-timeframe":       moduleDeleteTimeframe,
	"delete-duplicates":      moduleDeleteDuplicates,
}

// This wrappter gets a function from the above modules map and calls it with the parameters and the passed calendar.
// parameters can be any dictionary. The function will then choose how to handle the parameters.
// Returns the number of added entries. negative, if it removed entries.
func callModule(module func(*ics.Calendar, map[string]string) (int, error), params map[string]string, cal *ics.Calendar) (int, error) {
	return module(cal, params)
}

// This modules delete all events whose summary match the regex and are in the time range from the calendar.
// Parameters:
// - 'regex', mandatory: regular expression to remove.
// - 'from' & 'until', optional parameters: If timeframe is not given, all events matching the regex are removed.
//   Currenty if only either "from" or "until" is set, the timeframe will be ignored. TODO
// Returns the number of events removed. This number should always be negative.
func moduleDeleteSummaryRegex(cal *ics.Calendar, params map[string]string) (int, error) {
	var count int
	if params["regex"] == "" {
		return 0, fmt.Errorf("Missing mandatory Parameter 'regex'")
	}
	regex, _ := regexp.Compile(params["regex"])
	if params["from"] != "" && params["until"] != "" {
		from, _ := time.Parse(time.RFC3339, params["from"])
		until, _ := time.Parse(time.RFC3339, params["until"])
		count = removeByRegexSummaryAndTime(cal, *regex, from, until)
	} else {
		count = removeByRegexSummary(cal, *regex)
	}
	if count > 0 {
		return count, fmt.Errorf("This number should not be positive!")
	}
	return count, nil
}

// This function is a wrapper for removeByRegexSummaryAndTime, where the time is any time
func removeByRegexSummary(cal *ics.Calendar, regex regexp.Regexp) int {
	return removeByRegexSummaryAndTime(cal, regex, time.Time{}, time.Unix(1<<63-1-int64((1969*365+1969/4-1969/100+1969/400)*24*60*60), 999999999))
	// this is the maximum time that can be represented in the time.Time struct
}

// This function is used to remove the events that are in the time range and match the regex string.
// It returns the number of events removed. (always negative)
func removeByRegexSummaryAndTime(cal *ics.Calendar, regex regexp.Regexp, start time.Time, end time.Time) int {
	var count int
	for i := len(cal.Components) - 1; i >= 0; i-- { // iterate over events
		switch cal.Components[i].(type) {
		case *ics.VEvent:
			event := cal.Components[i].(*ics.VEvent)
			date, _ := event.GetStartAt()
			if date.After(start) && end.After(date) {
				// event is in time range
				if regex.MatchString(event.GetProperty(ics.ComponentPropertySummary).Value) {
					// event matches regex
					cal.Components = remove(cal.Components, i)
					log.Debug("Excluding event '" + event.GetProperty(ics.ComponentPropertySummary).Value + "' with id " + event.Id() + "\n")
					count--
				} else {
					log.Debug("Event '" + event.GetProperty(ics.ComponentPropertySummary).Value + "' with id " + event.Id() + " does not match regex" + regex.String() + "\n")
				}
			}
		default:
			// print type of component
			log.Debug("Component type: " + reflect.TypeOf(cal.Components[i]).String() + "\n")
		}
	}
	return count
}

// This module deletes an Event with the given id.
// Parameters: "id" mandatory
// Returns the number of events removed.
func moduleDeleteId(cal *ics.Calendar, params map[string]string) (int, error) {
	var count int
	if params["id"] == "" {
		return 0, fmt.Errorf("Missing mandatory Parameter 'id'")
	}
	for i, component := range cal.Components { // iterate over events
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)
			if event.Id() == params["id"] {
				cal.Components = remove(cal.Components, i)
				count--
				log.Debug("Excluding event with id " + params["id"] + "\n")
				break
			}
		}
	}
	return count, nil
}

// This function adds all events from cal2 to cal1.
// All other properties, such as TZ are retained from cal1.
func addEvents(cal1 *ics.Calendar, cal2 *ics.Calendar) int {
	var count int
	for _, event := range cal2.Events() {
		cal1.AddVEvent(event)
		count++
	}
	return count
}

func moduleAddURL(cal *ics.Calendar, params map[string]string) (int, error) {
	if params["url"] == "" {
		return 0, fmt.Errorf("Missing mandatory Parameter 'url'")
	}
	// put all params starting with header- into header map
	header := make(map[string]string)
	for k, v := range params {
		if strings.HasPrefix(k, "header-") {
			header[strings.TrimPrefix(k, "header-")] = v
		}
	}

	return addEventsURL(cal, params["url"], header)
}

func addEventsURL(cal *ics.Calendar, url string, headers map[string]string) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Errorln(err)
		return 0, fmt.Errorf("Error requesting additional URL: %s", err.Error())
	}
	if response.StatusCode != 200 {
		log.Errorf("Unexpected status '%s' from additional URL\n", response.Status)
		resp, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Errorln(err)
		}
		log.Debugf("Full response body: %s\n", resp)
		return 0, fmt.Errorf("Error response from additional URL: Status %s", response.Status)
	}
	// parse aditional calendar
	addcal, err := ics.ParseCalendar(response.Body)
	if err != nil {
		log.Errorln(err)
	}
	// add to new calendar
	return addEvents(cal, addcal), nil
}

func addMultiURL(cal *ics.Calendar, urls []string, header map[string]string) (int, error) {
	var count int
	for _, url := range urls {
		c, err := addEventsURL(cal, url, header)
		if err != nil {
			return count, err
		}
		count += c
	}
	return count, nil
}

func moduleAddFile(cal *ics.Calendar, params map[string]string) (int, error) {
	if params["filename"] == "" {
		return 0, fmt.Errorf("Missing mandatory Parameter 'file'")
	}
	return addEventsFile(cal, params["filename"])
}

func addEventsFile(cal *ics.Calendar, filename string) (int, error) {
	if _, err := os.Stat(filename); err != nil {
		return 0, fmt.Errorf("File %s not found", filename)
	}
	addicsfile, _ := os.Open(filename)
	addics, _ := ics.ParseCalendar(addicsfile)
	return addEvents(cal, addics), nil
}

func addMultiFile(cal *ics.Calendar, filenames []string) (int, error) {
	var count int
	for _, filename := range filenames {
		c, err := addEventsFile(cal, filename)
		if err != nil {
			return count, err
		}
		count += c
	}
	return count, nil
}

// Removes all Events in a passed Timeframe.
// Parameters: either "start" or "end" mandatory
// Format is RFC3339: "2006-01-02T15:04:05-07:00"
// Returns the number of events removed. (always negative)
func moduleDeleteTimeframe(cal *ics.Calendar, params map[string]string) (int, error) {
	var count int
	var after time.Time
	var before time.Time
	var err error
	if params["after"] == "" && params["before"] == "" {
		return 0, fmt.Errorf("Missing both Parameters 'start' or 'end'. One has to be present")
	}
	if params["after"] == "" {
		after = time.Time{}
	} else {
		after, err = time.Parse(time.RFC3339, params["start"])
		if err != nil {
			return 0, fmt.Errorf("Invalid start time: %s", err.Error())
		}
	}
	if params["end"] == "" {
		before = time.Unix(1<<63-1-int64((1969*365+1969/4-1969/100+1969/400)*24*60*60), 999999999)
	} else {
		before, err = time.Parse(time.RFC3339, params["end"])
		if err != nil {
			return 0, fmt.Errorf("Invalid end time: %s", err.Error())
		}
	}

	// remove events
	for i := len(cal.Components) - 1; i >= 0; i-- { // iterate over events
		switch cal.Components[i].(type) {
		case *ics.VEvent:
			event := cal.Components[i].(*ics.VEvent)
			date, _ := event.GetStartAt()
			if date.After(after) && before.After(date) {
				cal.Components = remove(cal.Components, i)
				count++
				log.Debug("Excluding event with id " + event.Id() + "\n")
			}
		}
	}

	return count, nil
}

// This Module deletes duplicate Events.
// No Parameters
// Returns the number of events removed. (always negative)
// Duplicates are defined by the same Summary and the same start and end time. Only the event that is latest in the file will be kept.
// TODO smarter: if the description or other components differ, it should combine them
func moduleDeleteDuplicates(cal *ics.Calendar, params map[string]string) (int, error) {
	var count int
	var uniques []string
	for i := len(cal.Components) - 1; i >= 0; i-- { // iterate over events backwards
		switch cal.Components[i].(type) {
		case *ics.VEvent:
			event := cal.Components[i].(*ics.VEvent)
			start, _ := event.GetStartAt()
			end, _ := event.GetEndAt()
			identifier := start.String() + end.String() + event.GetProperty(ics.ComponentPropertySummary).Value
			if stringInSlice(identifier, uniques) {
				cal.Components = remove(cal.Components, i)
				count--
				log.Debug("Excluding event with id " + event.Id() + "\n")
			} else {
				uniques = append(uniques, identifier)
			}
		}
	}
	return count, nil
}

// removes the element at index i from ics.Component slice
// warning: if you iterate over []ics.Component forward, this remove will lead to mistakes. Iterate backwards instead!
func remove(slice []ics.Component, s int) []ics.Component {
	return append(slice[:s], slice[s+1:]...)
}

// returns true, if a is in list b
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}