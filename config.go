package main

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// STRUCTS

type profile struct {
	Source        string              `yaml:"source"`
	Public        bool                `yaml:"public"`
	ImmutablePast bool                `yaml:"immutable-past,omitempty"`
	Modules       []map[string]string `yaml:"modules,omitempty"`
}

type mailConfig struct {
	SMTPServer string `yaml:"smtp_server"`
	SMTPPort   int    `yaml:"smtp_port"`
	Sender     string `yaml:"sender"`
	SMTPUser   string `yaml:"smtp_user,omitempty"`
	SMTPPass   string `yaml:"smtp_pass,omitempty"`
}

type serverConfig struct {
	Addr        string     `yaml:"addr"`
	URL         string     `yaml:"url"`
	LogLevel    log.Level  `yaml:"loglevel"`
	StoragePath string     `yaml:"storagepath"`
	Mail        mailConfig `yaml:"mail,omitempty"`
}

type notifier struct {
	Source     string   `yaml:"source"`
	Interval   string   `yaml:"interval"`
	Recipients []string `yaml:"recipients"`
}

// Config represents configuration for the application
type Config struct {
	Server    serverConfig        `yaml:"server"`
	Profiles  map[string]profile  `yaml:"profiles,omitempty"`
	Notifiers map[string]notifier `yaml:"notifiers,omitempty"`
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
		log.Fatalf("Error Unmarshalling Config: %v", err)
		return tmpConfig, err
	}

	// defaults
	if tmpConfig.Server.Addr == "" {
		tmpConfig.Server.Addr = ":8080"
	}
	if tmpConfig.Server.LogLevel == 0 {
		tmpConfig.Server.LogLevel = log.InfoLevel
	}
	if tmpConfig.Server.StoragePath == "" {
		tmpConfig.Server.StoragePath = "/etc/ical-relay/"
	}

	return tmpConfig, nil
}

func reloadConfig() error {
	// load config
	var err error
	conf, err = ParseConfig(configPath)
	if err != nil {
		return err
	} else {
		log.Info("Config reloaded")
		return nil
	}
}

func (c Config) saveConfig(path string) error {
	d, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, d, 0644)
}

// CONFIG EDITING FUNCTIONS

func (c Config) getPublicCalendars() []string {
	var cal []string
	for p := range c.Profiles {
		if c.Profiles[p].Public {
			cal = append(cal, p)
		}
	}
	return cal
}

func (c Config) addNotifyRecipient(notifier string, recipient string) error {
	if _, ok := c.Notifiers[notifier]; ok {
		n := c.Notifiers[notifier]
		n.Recipients = append(n.Recipients, recipient)
		c.Notifiers[notifier] = n
		return c.saveConfig(configPath)
	} else {
		return fmt.Errorf("notifier does not exist")
	}
}

func (c Config) removeNotifyRecipient(notifier string, recipient string) error {
	if _, ok := c.Notifiers[notifier]; ok {
		n := c.Notifiers[notifier]
		for i, r := range n.Recipients {
			if r == recipient {
				n.Recipients = append(n.Recipients[:i], n.Recipients[i+1:]...)
				c.Notifiers[notifier] = n
				return c.saveConfig(configPath)
			}
		}
		return fmt.Errorf("recipient not found")
	} else {
		return fmt.Errorf("notifier does not exist")
	}
}
