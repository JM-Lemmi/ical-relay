package helpers

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/mail"
	"os"

	ics "github.com/arran4/golang-ical"
)

func ReadCalURL(url string) (*ics.Calendar, error) {
	// download file
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	// parse original calendar
	return ics.ParseCalendar(response.Body)
}

func WriteCalFile(cal *ics.Calendar, filename string) error {
	// write file
	return ioutil.WriteFile(filename, []byte(cal.Serialize()), 0600)
}

func LoadCalFile(filename string) (*ics.Calendar, error) {
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

func DirectoryExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	// check if it's a directory
	return info.IsDir()
}

func PrettyPrint(e ics.VEvent) string {
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

// returns the MD5 hash of a string
func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// https://stackoverflow.com/a/66624104
func ValidMail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// returns true, if a is in list b
func StringInSlice(a string, list []string) bool {
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
func Remove(slice []interface{}, s int) []interface{} {
	return append(slice[:s], slice[s+1:]...)
}
func RemoveFromICS(slice []ics.Component, s int) []ics.Component {
	return append(slice[:s], slice[s+1:]...)
}
func RemoveFromMapString(slice []map[string]string, s int) []map[string]string {
	return append(slice[:s], slice[s+1:]...)
}

// warning: if you iterate over []ics.IANAProperty forward, this remove will lead to mistakes. Iterate backwards instead!
func RemoveProperty(slice []ics.IANAProperty, s int) []ics.IANAProperty {
	return append(slice[:s], slice[s+1:]...)
}

// This function adds all events from cal2 to cal1.
// All other properties, such as TZ are retained from cal1.
func AddEvents(cal1 *ics.Calendar, cal2 *ics.Calendar) int {
	var count int
	for _, event := range cal2.Events() {
		cal1.AddVEvent(event)
		count++
	}
	return count
}

// Fixes calendars where the timezone is just set once instead of in every VEvent
func FixTimezone(cal *ics.Calendar) {
	var property *ics.CalendarProperty
	for _, prop := range cal.CalendarProperties {
		if prop.IANAToken == "TIMEZONE" || prop.IANAToken == "X-WR-TIMEZONE" {
			property = &prop
			break
		}
	}
	// no default timezone for the calendar found
	if property == nil {
		return
	}
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
}
