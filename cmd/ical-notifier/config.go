package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// STRUCTS
// !! breaking changes need to keep the old version in legacyconfig.go !!

// Config represents static configuration for the application
type Config struct {
	Version   int                 `yaml:"version"`
	General   generalConfig       `yaml:"general"`
	Notifiers map[string]notifier `yaml:"notifiers,omitempty"`
}

type generalConfig struct {
	LiteMode    bool       `yaml:"lite"`          // if used in conjunction with ical-relay, this should be false.
	URL         string     `yaml:"url,omitempty"` // if used in conjunction with ical-relay, this should be the same as in ical-relay
	LogLevel    log.Level  `yaml:"loglevel"`
	StoragePath string     `yaml:"storagepath"` // should be the same as in ical-relay, if used in conjunction with ical-relay
	Name        string     `yaml:"name,omitempty"`
	DB          dbConfig   `yaml:"db,omitempty"`
	Mail        mailConfig `yaml:"mail,omitempty"`
}

type dbConfig struct {
	Host     string `yaml:"host"`
	DbName   string `yaml:"db-name"`
	User     string `yaml:"user"`
	Password string `yaml:"password,omitempty"`
}

type mailConfig struct {
	SMTPServer string `yaml:"smtp_server"`
	SMTPPort   int    `yaml:"smtp_port"`
	Sender     string `yaml:"sender"`
	SMTPUser   string `yaml:"smtp_user,omitempty"`
	SMTPPass   string `yaml:"smtp_pass,omitempty"`
}

type notifier struct {
	Source     string              `yaml:"source"`
	Recipients map[string][]string `yaml:"recipients"`
}

// CONFIG MANAGEMENT FUNCTIONS

// ParseConfig reads config from path and returns a Config struct
func ParseConfig(path string) (Config, error) {
	var tmpConfig Config

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error Reading Config: %v ", err)
		return tmpConfig, err
	}

	err = yaml.Unmarshal(yamlFile, &tmpConfig)
	if err != nil {
		log.Fatalf("Error Parsing  Config: %v", err)
		return tmpConfig, err
	}

	if tmpConfig.Version != 1 {
		log.Fatalf("Config Version Mismatch: %v", tmpConfig.Version)
		return tmpConfig, err
	}

	log.Trace("Read config, now setting defaults")
	// defaults
	if tmpConfig.General.LogLevel == 0 {
		tmpConfig.General.LogLevel = log.InfoLevel
	}
	if tmpConfig.General.StoragePath == "" {
		tmpConfig.General.StoragePath = filepath.Dir(path)
	}
	if !strings.HasSuffix(tmpConfig.General.StoragePath, "/") {
		tmpConfig.General.StoragePath += "/"
	}
	if tmpConfig.General.Name == "" {
		tmpConfig.General.Name = "Calendar"
	}

	if !helpers.DirectoryExists(tmpConfig.General.StoragePath + "notifystore/") {
		log.Info("Creating notifystore directory")
		err = os.MkdirAll(tmpConfig.General.StoragePath+"notifystore/", 0750)
		if err != nil {
			log.Fatalf("Error creating notifystore: %v", err)
			return tmpConfig, err
		}
	}

	return tmpConfig, nil
}
