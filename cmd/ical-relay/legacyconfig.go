package main

import (
	"os"

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

type ConfigV2 struct {
	Version   int                  `yaml:"version"`
	Server    serverConfig         `yaml:"server"`
	Profiles  map[string]profileV2 `yaml:"profiles,omitempty"`
	Notifiers map[string]notifier  `yaml:"notifiers,omitempty"`
}

type profileV2 struct {
	Name          string   `db:"name"`
	Sources       []string `yaml:"sources,omitempty"`
	Public        bool     `yaml:"public" db:"public"`
	ImmutablePast bool     `yaml:"immutable-past,omitempty" db:"immutable_past"`
	Tokens        []string `yaml:"admin-tokens"`
	NTokens       []token  `yaml:"admin-tokens-storage-v2,omitempty"`
	Rules         []Rule   `yaml:"rules,omitempty"`
}

type ConfigVersion struct {
	Version int `yaml:"version"`
}

// CONFIG MANAGEMENT FUNCTIONS

// ParseConfig reads config from path and returns a Config struct
// This function is used to parse the legacy config file
// It is only used by the config.go file as part of the initial config parsing
// This may be removed in the future
func LegacyParseConfig(path string) (Config, error) {
	var tmpConfig Config
	var oldConfig LegacyConfig
	var configVersion ConfigVersion

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error Reading Config: %v ", err)
	}

	err = yaml.Unmarshal(yamlFile, &configVersion)
	if err != nil {
		configVersion = ConfigVersion{Version: -1}
	}

	if configVersion.Version == 2 {
		var configV2 ConfigV2

		yamlFile, err = os.ReadFile(path)
		if err != nil {
			log.Fatalf("Error Reading Config: %v ", err)
		}

		err = yaml.Unmarshal(yamlFile, &configV2)
		if err != nil {
			log.Fatalf("Error Unmarshalling Config: %v", err)
		}

		tmpConfig.Server = configV2.Server
		tmpConfig.Notifiers = configV2.Notifiers
		tmpConfig.Version = 3
		tmpConfig.Profiles = make(map[string]profile)

		for p := range configV2.Profiles {
			var tokens []token
			for _, tokenString := range configV2.Profiles[p].Tokens {
				tokens = append(tokens, token{
					Token: tokenString,
				})
			}
			for _, nToken := range configV2.Profiles[p].NTokens {
				tokens = append(tokens, nToken)
			}

			tmpConfig.Profiles[p] = profile{
				Sources:       configV2.Profiles[p].Sources,
				Public:        configV2.Profiles[p].Public,
				ImmutablePast: configV2.Profiles[p].ImmutablePast,
				Tokens:        tokens,
				Rules:         configV2.Profiles[p].Rules,
			}
		}
		return tmpConfig, nil
	}

	if configVersion.Version != -1 {
		log.Fatalf("Unknown configVersion %d", configVersion)
	}

	//Attempting to resurrect unversioned config:

	yamlFile, err = os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error Reading Config: %v ", err)
	}

	err = yaml.Unmarshal(yamlFile, &oldConfig)
	if err != nil {
		log.Fatalf("Error Unmarshalling Config: %v", err)
	}

	// transfer the old config to the new one
	tmpConfig.Server = oldConfig.Server
	tmpConfig.Notifiers = oldConfig.Notifiers
	tmpConfig.Version = 3
	tmpConfig.Profiles = make(map[string]profile)

	for p := range oldConfig.Profiles {
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

		var tokens []token
		for _, tokenString := range oldConfig.Profiles[p].Tokens {
			tokens = append(tokens, token{
				Token: tokenString,
			})
		}

		tmpConfig.Profiles[p] = profile{
			Sources:       sources,
			Public:        oldConfig.Profiles[p].Public,
			ImmutablePast: oldConfig.Profiles[p].ImmutablePast,
			Tokens:        tokens,
			Rules:         rules,
		}
	}

	return tmpConfig, nil
}
