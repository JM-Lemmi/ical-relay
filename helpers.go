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

// returns true, if a is in list b
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// https://stackoverflow.com/a/37335777/9397749
// removes the element at index i from ics.Component slice
// warning: if you iterate over []ics.Component forward, this removeFromICS will lead to mistakes. Iterate backwards instead!
func remove(slice []interface{}, s int) []interface{} {
	return append(slice[:s], slice[s+1:]...)
}
func removeFromICS(slice []ics.Component, s int) []ics.Component {
	return append(slice[:s], slice[s+1:]...)
}
func removeFromMapString(slice []map[string]string, s int) []map[string]string {
	return append(slice[:s], slice[s+1:]...)
}

// warning: if you iterate over []ics.IANAProperty forward, this remove will lead to mistakes. Iterate backwards instead!
func removeProperty(slice []ics.IANAProperty, s int) []ics.IANAProperty {
	return append(slice[:s], slice[s+1:]...)
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
