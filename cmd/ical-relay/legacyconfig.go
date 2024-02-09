package main

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// STRUCTS
// These structs only include the fields that differ in the legacy config, otherwise reuses the original config.
// This also means that if something breaking is changed in the new config, the old version must be preserved here.

type legacyProfile struct {
	Source        string              `yaml:"source"`
	Public        bool                `yaml:"public"`
	ImmutablePast bool                `yaml:"immutable-past,omitempty"`
	Tokens        []string            `yaml:"admin-tokens"`
	Modules       []map[string]string `yaml:"modules,omitempty"`
}

// Config represents configuration for the application
type LegacyConfig struct {
	Server    serverConfig             `yaml:"server"`
	Profiles  map[string]legacyProfile `yaml:"profiles,omitempty"`
	Notifiers map[string]notifier      `yaml:"notifiers,omitempty"`
}

// CONFIG MANAGEMENT FUNCTIONS

// ParseConfig reads config from path and returns a Config struct
// This function is used to parse the legacy config file
// It is only used by the config.go file as part of the initial config parsing
// This may be removed in the future
func LegacyParseConfig(path string) (Config, error) {
	var tmpConfig Config
	var oldConfig LegacyConfig

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error Reading Config: %v ", err)
		return tmpConfig, err
	}

	err = yaml.Unmarshal(yamlFile, &oldConfig)
	if err != nil {
		log.Fatalf("Error Unmarshalling Config: %v", err)
		return tmpConfig, err
	}

	// transfer the old config to the new one
	tmpConfig.Server = oldConfig.Server
	tmpConfig.Notifiers = oldConfig.Notifiers
	tmpConfig.Version = 2
	tmpConfig.Profiles = make(map[string]profile)

	for p, _ := range oldConfig.Profiles {
		// build new rules and sources from old modules
		var rules []Rule
		var sources []string
		sources = append(sources, oldConfig.Profiles[p].Source)
		for _, m := range oldConfig.Profiles[p].Modules {
			var r Rule

			switch m["name"] {
			case "delete-bysummary-regex":
				r = Rule{
					Filters: []map[string]string{{
						"type":  "regex",
						"regex": m["regex"],
					}},
					Action: map[string]string{
						"type": "delete",
					},
					Expiry: m["expiry"],
				}
				rules = append(rules, r)
			case "delete-byid":
				r = Rule{
					Filters: []map[string]string{{
						"type": "id",
						"id":   m["id"],
					}},
					Action: map[string]string{
						"type": "delete",
					},
					Expiry: m["expiry"],
				}
				rules = append(rules, r)
			case "add-url":
				sources = append(sources, m["url"])
			case "add-file":
				sources = append(sources, "file://"+m["filename"])
			case "delete-timeframe":
				r = Rule{
					Filters: []map[string]string{{
						"type":   "timeframe",
						"after":  m["after"],
						"before": m["before"],
					}},
					Action: map[string]string{
						"type": "delete",
					},
					Expiry: m["expiry"],
				}
				rules = append(rules, r)
			case "delete-duplicates":
				r = Rule{
					Filters: []map[string]string{{
						"type": "duplicates",
					}},
					Action: map[string]string{
						"type": "delete",
					},
					Expiry: m["expiry"],
				}
				rules = append(rules, r)
			case "edit-byid":
				r = Rule{
					Filters: []map[string]string{{
						"type": "id",
						"id":   m["id"],
					}},
					Action: map[string]string{
						"type":            "edit",
						"overwrite":       m["overwrite"],
						"new-summary":     m["new-summary"],
						"new-description": m["new-description"],
						"new-start":       m["new-start"],
						"new-end":         m["new-end"],
						"new-location":    m["new-location"],
					},
					Expiry: m["expiry"],
				}
				rules = append(rules, r)
			case "edit-bysummary-regex":
				r = Rule{
					Filters: []map[string]string{{
						"type":  "regex",
						"regex": m["regex"],
					}},
					Action: map[string]string{
						"type":            "edit",
						"overwrite":       m["overwrite"],
						"new-summary":     m["new-summary"],
						"new-description": m["new-description"],
						"new-start":       m["new-start"],
						"new-end":         m["new-end"],
						"new-location":    m["new-location"],
						"move-time":       m["move-time"],
					},
					Expiry: m["expiry"],
				}
				rules = append(rules, r)
			case "save-to-file":
				log.Warn("'save-to-file' module is deprecated and will be removed from the config.")
			default:
				log.Warnf("Unknown module '%s' in profile '%s'", m["name"], p)
			}
		}

		tmpConfig.Profiles[p] = profile{
			Sources:       sources,
			Public:        oldConfig.Profiles[p].Public,
			ImmutablePast: oldConfig.Profiles[p].ImmutablePast,
			Tokens:        oldConfig.Profiles[p].Tokens,
			Rules:         rules,
		}
	}

	return tmpConfig, nil
}
