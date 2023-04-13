package main

import (
	"fmt"
	"io/ioutil"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

var modules = map[string]func(*ics.Calendar, map[string]string) (int, error){
	"save-to-file": moduleSaveToFile,
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

// removes the element at index i from ics.Component slice
// warning: if you iterate over []ics.IANAProperty forward, this remove will lead to mistakes. Iterate backwards instead!
func removeProperty(slice []ics.IANAProperty, s int) []ics.IANAProperty {
	return append(slice[:s], slice[s+1:]...)
}
