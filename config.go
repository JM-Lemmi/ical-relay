package main

import (
	"regexp"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type regex struct {
	regexp.Regexp
}

type profile struct {
	RegEx  []regex
	Public bool
}

type serverConfig struct {
	Addr     string
	LogLevel log.Level
}

// Config represents configuration for the application
type Config struct {
	URL      string
	Profiles map[string]profile
	Server   serverConfig
}

func (r *regex) UnmarshalText(text []byte) error {
	tmpRe, err := regexp.Compile("(?i)" + string(text))
	r.Regexp = *tmpRe
	return err
}

// ParseConfig reads config from path and returns a Config struct
func ParseConfig(path string) (Config, error) {
	var tmpConfig Config

	if _, err := toml.DecodeFile(path, &tmpConfig); err != nil {
		return tmpConfig, err
	}

	// defaults
	if tmpConfig.Server.Addr == "" {
		tmpConfig.Server.Addr = ":8080"
	}
	if tmpConfig.Server.LogLevel == 0 {
		tmpConfig.Server.LogLevel = log.InfoLevel
	}

	return tmpConfig, nil
}
