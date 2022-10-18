package main

import (
	"io/ioutil"
	"net/http"

	ics "github.com/arran4/golang-ical"
)

func readCalURL(url string) (*ics.Calendar, error) {
	// download file
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	// parse original calendar
	return ics.ParseCalendar(response.Body)
}

func writeCalFile(cal *ics.Calendar, filename string) error {
	// write file
	return ioutil.WriteFile(filename, []byte(cal.Serialize()), 0644)
}