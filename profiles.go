package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"

	ics "github.com/arran4/golang-ical"
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
	origlen := len(calendar.Events())
	var addedEvents int

	for _, module_request := range profile.Rules {
		log.Debug("Requested module: ", module_request["name"])
		module, ok := modules[module_request["name"]]
		if !ok {
			return nil, fmt.Errorf("module '%s' doesn't exist", module_request["name"])
		}
		count, err := callModule(module, module_request, calendar)
		if err != nil {
			return nil, err
		}
		addedEvents += count
	}

	// immutable past:
	historyFilename := conf.Server.StoragePath + "calstore/" + profileName + "-past.ics"
	if profile.ImmutablePast {
		// check if file exists, if not download for the first time
		if _, err := os.Stat(historyFilename); os.IsNotExist(err) {
			log.Info("History file does not exist, saving for the first time")
			historyCal := calendar
			_, err := moduleDeleteTimeframe(historyCal, map[string]string{"after": "now"})
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
		_, err = moduleDeleteTimeframe(historyCal, map[string]string{"after": "now"})
		if err != nil {
			log.Errorln(err)
			return calendar, fmt.Errorf("Error executing immutable past (setup): %s", err.Error())
		}

		// delete events from calendar that are in the past
		log.Debug("Removing past from calendar")
		count, err := moduleDeleteTimeframe(calendar, map[string]string{"before": "now"})
		if err != nil {
			log.Errorln(err)
			return calendar, fmt.Errorf("Error executing immutable past (delete): %s", err.Error())
		}
		addedEvents += count
		// combine calendars
		log.Debug("Combining calendars")
		count = addEvents(calendar, historyCal)
		if err != nil {
			log.Errorln(err)
			return calendar, fmt.Errorf("Error executing immutable past (adding): %s", err.Error())
		}
		addedEvents += count

		//saving history file
		log.Debug("Saving history file")
		err = writeCalFile(calendar, historyFilename)
		if err != nil {
			log.Errorln(err)
			return calendar, fmt.Errorf("Error saving history file: %s", err.Error())
		}
	}
	// it may be neccesary to run delete-duplicates here to avoid duplicates from the history file

	// make sure new calendar has all events but excluded and added
	eventCountDiff := origlen + addedEvents - len(calendar.Events())
	if eventCountDiff != 0 {
		log.Warnf("Calendar has %d events after applying rules, but should have %d", len(calendar.Events()), origlen+addedEvents)
	}
	log.Debugf("Added %d events", addedEvents)
	return calendar, nil
}
