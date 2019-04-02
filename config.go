package main

import (
	"regexp"

	"github.com/BurntSushi/toml"
)

type regex struct {
	regexp.Regexp
}

type profile struct {
	RegEx  []string
	Public bool
}

type serverConfig struct {
	Addr string
}

// Config represents configuration for the application
type Config struct {
	URL      string
	Profiles map[string]profile
	Server   serverConfig
}

func (r regex) UnmarshalText(text []byte) error {
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

	return tmpConfig, nil
}
