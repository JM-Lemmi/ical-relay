package main

import (
	"fmt"
	"net/http"
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

func getProfileCalendar(profile profile) (*ics.Calendar, error) {
	var calendar *ics.Calendar

	// get the base calendar to which to apply modules
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

	// apply modules
	origlen := len(calendar.Events())
	var addedEvents int

	for _, module_request := range profile.Modules {
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

	// make sure new calendar has all events but excluded and added
	eventCountDiff := origlen + addedEvents - len(calendar.Events())
	if eventCountDiff != 0 {
		log.Warnf("Calendar has %d events after applying modules, but should have %d", len(calendar.Events()), origlen+addedEvents)
	}
	log.Debugf("Added %d events", addedEvents)
	return calendar, nil
}
