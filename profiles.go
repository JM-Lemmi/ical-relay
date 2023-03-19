package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"

	ics "github.com/arran4/golang-ical"
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
	for name, this_profile := range conf.Profiles {
		if this_profile.Public {
			// FIXME: any name with "/" will break the URL
			viewUrl, err := router.Get("monthlyView").URL("profile", name)
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
	}
	// sort profiles by name
	sort.SliceStable(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	return profiles
}

func getProfileCalendar(profile profile, profileName string) (*ics.Calendar, error) {
	var calendar *ics.Calendar

	// get the base calendar to which to apply rules
	if profile.Source == "" {
		calendar = ics.NewCalendar()
	} else {
		response, err := http.Get(profile.Source)
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
	}

	// apply rules
	for i, rule := range profile.Rules {
		log.Debug("Executing Rule ", i)

		var indices []int

		// run filters
		for _, filter := range rule.Filters {
			filter_name, ok := filters[filter["type"]]
			if !ok {
				return nil, fmt.Errorf("filter type '%s' doesn't exist", filter["type"])
			}
			local_indices, err := callFilter(filter_name, calendar, filter)
			if err != nil {
				return nil, err
			}

			if rule.Operator == "and" || rule.Operator == "" {
				indices = intersect.SimpleGeneric(indices, local_indices)
			} else if rule.Operator == "or" {
				indices = append(indices, local_indices...)
			} else {
				return nil, fmt.Errorf("Unknown operator '%s'", rule.Operator)
			}
		}

		// run action
		action_name, ok := actions[rule.Action["type"]]
		if !ok {
			return nil, fmt.Errorf("action type '%s' doesn't exist", rule.Action["type"])
		}
		err := callAction(action_name, calendar, indices, rule.Action)
		if err != nil {
			return nil, err
		}
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
			writeCalFile(calendar, historyFilename)
		}

		// load history file
		log.Debug("Loading history file")
		historyCal, err := loadCalFile(historyFilename)
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
		addEvents(calendar, historyCal)
		if err != nil {
			log.Errorln(err)
			return calendar, fmt.Errorf("Error executing immutable past (adding): %s", err.Error())
		}

		//saving history file
		log.Debug("Saving history file")
		err = writeCalFile(calendar, historyFilename)
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
	indices, err := callFilter(filters["timeframe"], cal, map[string]string{"timeframe": timeframe})
	if err != nil {
		return err
	}
	err = callAction(actions["delete"], cal, indices, map[string]string{})
	if err != nil {
		return err
	}
	return nil
}
