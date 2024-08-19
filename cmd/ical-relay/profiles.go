package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	ics "github.com/arran4/golang-ical"
	"github.com/jm-lemmi/ical-relay/database"
	"github.com/jm-lemmi/ical-relay/helpers"
	"github.com/jm-lemmi/ical-relay/modules"
	"github.com/juliangruber/go-intersect/v2" // requires go1.18
	log "github.com/sirupsen/logrus"
)

type profileMetadata struct {
	Name    string
	ViewURL string
	IcalURL string
}

func getProfilesMetadata() []profileMetadata {
	profiles := make([]profileMetadata, 0)
	for _, name := range dataStore.GetPublicProfileNames() {
		// FIXME: any name with "/" will break the URL
		viewUrl, err := router.Get("calendarView").URL("profile", name)
		if err != nil {
			log.Errorln(err)
			continue
		}
		icalUrl, err := router.Get("profile").URL("profile", name)
		if err != nil {
			log.Errorln(err)
			continue
		}
		profiles = append(profiles, profileMetadata{
			Name:    name,
			ViewURL: viewUrl.String(),
			IcalURL: icalUrl.String(),
		})
	}
	// sort profiles by name
	sort.SliceStable(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	return profiles
}

func getProfileCalendar(profile database.Profile, profileName string) (*ics.Calendar, error) {
	var calendar *ics.Calendar

	// get all sources
	if len(profile.Sources) == 0 {
		log.Debug("No sources, creating empty calendar")
		calendar = ics.NewCalendar()
	} else {
		// loop over sources and combine
		var ncalendar *ics.Calendar
		var err error

		for i, s := range profile.Sources {
			if i == 0 {
				// first source gets assigned to base calendar
				log.Debug("Loading source ", s, " as base calendar")
				calendar, err = getSource(s)
				if err != nil {
					return nil, err
				}
			} else {
				// all other calendars only load events
				log.Debug("Loading source ", s, " as additional calendar")
				ncalendar, err = getSource(s)
				if err != nil {
					return nil, err
				}
				helpers.AddEvents(calendar, ncalendar)
			}
		}
	}

	// apply rules
	for i, rule := range profile.Rules {
		log.Debug("Executing Rule ", i)

		var indices []int

		// run filters
		for _, filter := range rule.Filters {
			filter_name, ok := modules.Filters[filter["type"]]
			if !ok {
				return nil, fmt.Errorf("filter type '%s' doesn't exist", filter["type"])
			}
			local_indices, err := modules.CallFilter(filter_name, calendar, filter)
			if err != nil {
				return nil, err
			}

			log.Trace("Filter operator: ", rule.Operator)
			if rule.Operator == "and" {
				indices = intersect.SimpleGeneric(indices, local_indices)
			} else if rule.Operator == "or" || rule.Operator == "" {
				indices = append(indices, local_indices...)
			} else {
				return nil, fmt.Errorf("Unknown operator '%s'", rule.Operator)
			}
		}
		log.Trace("Indices after all filters: ", indices)

		// run action
		action_name, ok := modules.Actions[rule.Action["type"]]
		if !ok {
			return nil, fmt.Errorf("action type '%s' doesn't exist", rule.Action["type"])
		}
		err := modules.CallAction(action_name, calendar, indices, rule.Action)
		if err != nil {
			return nil, err
		}
		log.Trace("Finished action!")
	}

	// immutable past:
	historyFilename := conf.Server.StoragePath + "calstore/" + profileName + "-past.ics"
	if profile.ImmutablePast {
		// check if file exists, if not download for the first time
		if _, err := os.Stat(historyFilename); os.IsNotExist(err) {
			log.Info("History file does not exist, saving for the first time")
			historyCal := calendar
			err := ImmutablePastDelete(historyCal, "after")
			if err != nil {
				log.Errorln(err)
				return calendar, fmt.Errorf("Error executing immutable past (first-run): %s", err.Error())
			}
			helpers.WriteCalFile(calendar, historyFilename)
		}

		// load history file
		log.Debugf("Loading history file %s", historyFilename)
		historyCal, err := helpers.LoadCalFile(historyFilename)
		if err != nil {
			log.Errorln(err)
			return calendar, fmt.Errorf("Error loading history file: %s", err.Error())
		}
		log.Debug("Removing future from history file")
		// delete events from historyCal that are in the future
		err = ImmutablePastDelete(historyCal, "after")
		if err != nil {
			log.Errorln(err)
			return calendar, fmt.Errorf("Error executing immutable past (setup): %s", err.Error())
		}

		// delete events from calendar that are in the past
		log.Debug("Removing past from calendar")
		err = ImmutablePastDelete(calendar, "before")
		if err != nil {
			log.Errorln(err)
			return calendar, fmt.Errorf("Error executing immutable past (delete): %s", err.Error())
		}
		// combine calendars
		log.Debug("Combining calendars")
		helpers.AddEvents(calendar, historyCal)
		if err != nil {
			log.Errorln(err)
			return calendar, fmt.Errorf("Error executing immutable past (adding): %s", err.Error())
		}

		//saving history file
		log.Debugf("Saving history file %s", historyFilename)
		err = helpers.WriteCalFile(calendar, historyFilename)
		if err != nil {
			log.Errorln(err)
			return calendar, fmt.Errorf("Error saving history file: %s", err.Error())
		}
	}
	// it may be neccesary to run delete-duplicates here to avoid duplicates from the history file

	return calendar, nil
}

// Delete Helper funtion for immutable past.
// Will delete events from the calendar either before or after now.
// timeframes: "before": delete up till now, "after" delete everything after now
func ImmutablePastDelete(cal *ics.Calendar, timeframe string) error {
	indices, err := modules.CallFilter(modules.Filters["timeframe"], cal, map[string]string{timeframe: "now"})
	if err != nil {
		return err
	}
	err = modules.CallAction(modules.Actions["delete"], cal, indices, map[string]string{})
	if err != nil {
		return err
	}
	return nil
}

func getSource(source string) (*ics.Calendar, error) {
	var calendar *ics.Calendar
	var err error

	switch strings.Split(source, "://")[0] {
	case "http", "https":
		response, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP error: %s", response.Status)
		}
		if err != nil {
			return nil, err
		}
		calendar, err = ics.ParseCalendar(response.Body)
		if err != nil {
			return nil, err
		}
	case "file":
		calendar, err = helpers.LoadCalFile(strings.Split(source, "://")[1])
		if err != nil {
			return nil, err
		}
	case "profile":
		profileName := strings.Split(source, "://")[1]
		if !dataStore.ProfileExists(profileName) {
			return nil, fmt.Errorf("Profile does not exist: %s", profileName)
		}
		calendar, err = getProfileCalendar(dataStore.GetProfileByName(profileName), profileName)
		if err != nil {
			return nil, err
		}
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(strings.Split(source, "://")[1])
		if err != nil {
			return nil, err
		}

		calendar, err = ics.ParseCalendar(bytes.NewReader(decoded))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown source type '%s'", strings.Split(source, "://")[0])
	}
	return calendar, nil
}
