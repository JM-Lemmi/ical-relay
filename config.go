package main

import (
	"io/ioutil"
	"regexp"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type regex struct {
	regexp.Regexp
}

type profile struct {
	Source  string              `yaml:"source"`
	Public  bool                `yaml:"public"`
	Modules []map[string]string `yaml:"modules,omitempty"`
}

type serverConfig struct {
	Addr     string    `yaml:"addr"`
	LogLevel log.Level `yaml:"loglevel"`
}

// Config represents configuration for the application
type Config struct {
	Profiles map[string]profile `yaml:"profiles"`
	Server   serverConfig       `yaml:"server"`
}

func (r *regex) UnmarshalText(text []byte) error {
	tmpRe, err := regexp.Compile("(?i)" + string(text))
	r.Regexp = *tmpRe
	return err
}

// ParseConfig reads config from path and returns a Config struct
func ParseConfig(path string) (Config, error) {
	var tmpConfig Config

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, &tmpConfig)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
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
