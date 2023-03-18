package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

var modules = map[string]func(*ics.Calendar, map[string]string) (int, error){
	"add-url":      moduleAddURL,
	"add-file":     moduleAddFile,
	"save-to-file": moduleSaveToFile,
	"add-reminder": moduleAddAllReminder,
}

// These modules are allowed to be edited by the module admin. This is a security measure to prevent SSRF and LFI attacks.
var lowPrivModules = []string{
	"delete-bysummary-regex",
	"delete-byid",
	"delete-timeframe",
	"delete-duplicates",
	"edit-byid",
	"edit-bysummary-regex",
}

// This wrappter gets a function from the above modules map and calls it with the parameters and the passed calendar.
// parameters can be any dictionary. The function will then choose how to handle the parameters.
// Returns the number of added entries. negative, if it removed entries.
func callModule(module func(*ics.Calendar, map[string]string) (int, error), params map[string]string, cal *ics.Calendar) (int, error) {
	return module(cal, params)
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
		return 0, fmt.Errorf("missing mandatory Parameter 'url'")
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
		return 0, fmt.Errorf("error requesting additional URL: %s", err.Error())
	}
	if response.StatusCode != 200 {
		log.Warnf("Unexpected status '%s' from additional URL '%s'", response.Status, url)
		resp, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Errorln(err)
		}
		log.Debugf("Full response body: %s\n", resp)
		return 0, nil // we are not returning an error here, to just ignore URLs that are unavailible. TODO: make this configurable
	}
	// parse aditional calendar
	addcal, err := ics.ParseCalendar(response.Body)
	if err != nil {
		log.Errorln(err)
	}
	// add to new calendar
	return addEvents(cal, addcal), nil
}

// This module saves the current calendar to a file.
// Parameters: "file" mandatory: full path of file to save
func moduleSaveToFile(cal *ics.Calendar, params map[string]string) (int, error) {
	if params["file"] == "" {
		return 0, fmt.Errorf("missing mandatory Parameter 'file'")
	}
	err := ioutil.WriteFile(params["file"], []byte(cal.Serialize()), 0600)
	if err != nil {
		log.Errorln(err)
		return 0, fmt.Errorf("error writing to file: %s", err.Error())
	}
	return 0, nil
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
		return 0, fmt.Errorf("missing mandatory Parameter 'filename'")
	}
	return addEventsFile(cal, params["filename"])
}

func addEventsFile(cal *ics.Calendar, filename string) (int, error) {
	if _, err := os.Stat(filename); err != nil {
		return 0, fmt.Errorf("file %s not found", filename)
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

func moduleAddAllReminder(cal *ics.Calendar, params map[string]string) (int, error) {
	// add reminder to calendar
	for i := len(cal.Components) - 1; i >= 0; i-- {
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
	return 0, nil
}

// removes the element at index i from ics.Component slice
// warning: if you iterate over []ics.Component forward, this remove will lead to mistakes. Iterate backwards instead!
func remove(slice []ics.Component, s int) []ics.Component {
	return append(slice[:s], slice[s+1:]...)
}

// removes the element at index i from ics.Component slice
// warning: if you iterate over []ics.IANAProperty forward, this remove will lead to mistakes. Iterate backwards instead!
func removeProperty(slice []ics.IANAProperty, s int) []ics.IANAProperty {
	return append(slice[:s], slice[s+1:]...)
}
