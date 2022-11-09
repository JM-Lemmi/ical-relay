package main

import (
	"io/ioutil"
	"net/http"
	"net/mail"
	"os"

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
	return ioutil.WriteFile(filename, []byte(cal.Serialize()), 0600)
}

func loadCalFile(filename string) (*ics.Calendar, error) {
	var cal *ics.Calendar
	// read file
	file, err := os.Open(filename)
	if err != nil {
		return cal, err
	}
	// parse original calendar
	cal, err = ics.ParseCalendar(file)
	if err != nil {
		return cal, err
	}
	return cal, nil
}

func directoryExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	// check if it's a directory
	return info.IsDir()
}

func prettyPrint(e ics.VEvent) string {
	var output string
	output += e.GetProperty(ics.ComponentPropertySummary).Value + "\n"

	start, err := e.GetStartAt()
	if err != nil {
		start, _ = e.GetAllDayStartAt()
		output += start.Format("02. Jan 2006") + " - "
	} else {
		output += start.Format("Mon 02. Jan 2006, 15:04") + " - "
	}
	end, err := e.GetEndAt()
	if err != nil {
		end, _ = e.GetAllDayEndAt()
		output += end.Format("02. Jan 2006") + "\n"
	} else {
		output += end.Format("15:04") + "\n"
	}

	if e.GetProperty(ics.ComponentPropertyLocation) != nil {
		output += "Location: " + e.GetProperty(ics.ComponentPropertyLocation).Value + "\n"
	}
	if e.GetProperty(ics.ComponentPropertyDescription) != nil {
		output += "Description: " + e.GetProperty(ics.ComponentPropertyDescription).Value + "\n"
	}

	return output
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// https://stackoverflow.com/a/66624104
func validMail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
