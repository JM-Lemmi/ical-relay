package main

import (
	"fmt"
	"io/ioutil"
	"regexp"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type regex struct {
	regexp.Regexp
}

func (r *regex) UnmarshalText(text []byte) error {
	tmpRe, err := regexp.Compile("(?i)" + string(text))
	r.Regexp = *tmpRe
	return err
}

// STRUCTS

type profile struct {
	Source  string              `yaml:"source"`
	Public  bool                `yaml:"public"`
	Tokens  []string            `yaml:"admin-tokens"`
	Modules []map[string]string `yaml:"modules,omitempty"`
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
	Mail        mailConfig `yaml:"mail,omitempty"`
	SuperTokens []string   `yaml:"super-tokens,omitempty"`
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

func (c Config) profileExists(name string) bool {
	_, ok := c.Profiles[name]
	return ok
}

func (c Config) notifierExists(name string) bool {
	_, ok := c.Notifiers[name]
	return ok
}

func (c Config) addNotifyRecipient(notifier string, recipient string) error {
	if _, ok := c.Notifiers[notifier]; ok {
		n := c.Notifiers[notifier]
		n.Recipients = append(n.Recipients, recipient)
		c.Notifiers[notifier] = n
		return c.saveConfig(configPath)
	} else {
		return fmt.Errorf("Notifier does not exist")
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
		return fmt.Errorf("Recipient not found")
	} else {
		return fmt.Errorf("Notifier does not exist")
	}
}
