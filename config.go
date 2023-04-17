package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// STRUCTS
// !! breaking changes need to keep the old version in legacyconfig.go !!

// Config represents configuration for the application
type Config struct {
	Version   int                 `yaml:"version"`
	Server    serverConfig        `yaml:"server"`
	Profiles  map[string]profile  `yaml:"profiles,omitempty"`
	Notifiers map[string]notifier `yaml:"notifiers,omitempty"`
}

type serverConfig struct {
	Addr          string     `yaml:"addr"`
	URL           string     `yaml:"url"`
	LogLevel      log.Level  `yaml:"loglevel"`
	StoragePath   string     `yaml:"storagepath"`
	TemplatePath  string     `yaml:"templatepath"`
	Imprint       string     `yaml:"imprintlink"`
	PrivacyPolicy string     `yaml:"privacypolicylink"`
	Mail          mailConfig `yaml:"mail,omitempty"`
	SuperTokens   []string   `yaml:"super-tokens,omitempty"`
}

type mailConfig struct {
	SMTPServer string `yaml:"smtp_server"`
	SMTPPort   int    `yaml:"smtp_port"`
	Sender     string `yaml:"sender"`
	SMTPUser   string `yaml:"smtp_user,omitempty"`
	SMTPPass   string `yaml:"smtp_pass,omitempty"`
}

type profile struct {
	Source        string   `yaml:"source,omitempty"`
	Sources       []string `yaml:"sources,omitempty"`
	Public        bool     `yaml:"public"`
	ImmutablePast bool     `yaml:"immutable-past,omitempty"`
	Tokens        []string `yaml:"admin-tokens"`
	Rules         []Rule   `yaml:"rules,omitempty"`
}

type Rule struct {
	Filters  []map[string]string `yaml:"filters" json:"filters"`
	Operator string              `yaml:"operator" json:"operator"`
	Action   map[string]string   `yaml:"action" json:"action"`
	Expiry   string              `yaml:"expiry,omitempty" json:"expiry,omitempty"`
}

type notifier struct {
	Source     string   `yaml:"source"`
	Interval   string   `yaml:"interval"`
	Recipients []string `yaml:"recipients"`
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
		log.Warnf("Error Unmarshalling Config: %v; attempting to parse legacy config", err)
		tmpConfig, err = LegacyParseConfig(path)
		if err != nil {
			log.Fatalf("Error Parsing Legacy Config: %v", err)
			return tmpConfig, err
			// parsing in legacy mode failed
		}
		// parsing in legacy mode succeeded, saving new config
		tmpConfig.saveConfig(path)
	}

	// check if config is up to date, if not
	if tmpConfig.Version < 2 {
		log.Warn("Config is outdated, upgrading")
		tmpConfig, err = LegacyParseConfig(path)
		if err != nil {
			log.Fatalf("Error Parsing Legacy Config: %v", err)
			return tmpConfig, err
		}
		tmpConfig.saveConfig(path)
	}

	log.Trace("Read config, now setting defaults")
	// defaults
	if tmpConfig.Server.Addr == "" {
		tmpConfig.Server.Addr = ":8080"
	}
	if tmpConfig.Server.LogLevel == 0 {
		tmpConfig.Server.LogLevel = log.InfoLevel
	}
	if tmpConfig.Server.StoragePath == "" {
		tmpConfig.Server.StoragePath = filepath.Dir(path)
	}
	if !strings.HasSuffix(tmpConfig.Server.StoragePath, "/") {
		tmpConfig.Server.StoragePath += "/"
	}
	if tmpConfig.Server.TemplatePath == "" {
		tmpConfig.Server.TemplatePath = filepath.Dir("/opt/ical-relay/templates/")
	}
	if !strings.HasSuffix(tmpConfig.Server.TemplatePath, "/") {
		tmpConfig.Server.TemplatePath += "/"
	}

	if !directoryExists(tmpConfig.Server.StoragePath + "notifystore/") {
		log.Info("Creating notifystore directory")
		err = os.MkdirAll(tmpConfig.Server.StoragePath+"notifystore/", 0750)
		if err != nil {
			log.Fatalf("Error creating notifystore: %v", err)
			return tmpConfig, err
		}
	}
	if !directoryExists(tmpConfig.Server.StoragePath + "calstore/") {
		log.Info("Creating calstore directory")
		err = os.MkdirAll(tmpConfig.Server.StoragePath+"calstore/", 0750)
		if err != nil {
			log.Fatalf("Error creating calstore: %v", err)
			return tmpConfig, err
		}
	}

	// check if all profiles have a source, or maybe duplicated source
	for i, p := range tmpConfig.Profiles {
		if tmpConfig.Profiles[i].Source != "" {
			p.Sources = append([]string{tmpConfig.Profiles[i].Source}, tmpConfig.Profiles[i].Sources...)
			p.Source = ""
			tmpConfig.Profiles[i] = p
		}
		log.Trace("Loading Profile Sources complete. Profile " + i + " has sources: " + strings.Join(p.Sources, ", "))
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
	return ioutil.WriteFile(path, d, 0600)
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

func (c Config) addNotifierFromProfile(name string) {
	c.Notifiers[name] = notifier{
		Source:     c.Server.URL + "/profiles/" + name,
		Interval:   "1h",
		Recipients: []string{},
	}
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

func (c Config) addRule(profile string, rule Rule) error {
	if !c.profileExists(profile) {
		return fmt.Errorf("profile " + profile + " does not exist")
	}
	p := c.Profiles[profile]
	p.Rules = append(c.Profiles[profile].Rules, rule)
	c.Profiles[profile] = p
	return c.saveConfig(configPath)
}

func (c Config) removeRuleFromProfile(profile string, index int) {
	log.Info("Removing rule at position " + fmt.Sprint(index+1) + " from profile " + profile)
	p := c.Profiles[profile]
	p.Rules = append(p.Rules[:index], p.Rules[index+1:]...)
	c.Profiles[profile] = p
	c.saveConfig(configPath)
}

func (c Config) addSource(profile string, src string) error {
	if !c.profileExists(profile) {
		return fmt.Errorf("profile " + profile + " does not exist")
	}
	p := c.Profiles[profile]
	p.Sources = append(c.Profiles[profile].Sources, src)
	c.Profiles[profile] = p
	return c.saveConfig(configPath)
}

func (c Config) RunCleanup() {
	for p := range c.Profiles {
		for i, m := range c.Profiles[p].Rules {
			if m.Expiry != "" {
				exp, _ := time.Parse(time.RFC3339, m.Expiry)
				if time.Now().After(exp) {
					c.removeRuleFromProfile(p, i)
				}
			}
		}
	}
}

// checks if a rule is valid.
// returns true if rule is valid, false if not
func checkRuleIntegrity(rule Rule) bool {
	// TODO implement!
	return true
}

// starts a heartbeat notifier in a sub-routine
func CleanupStartup() {
	log.Info("Starting Cleanup Timer")
	go TimeCleanup()
}

func TimeCleanup() {
	interval, _ := time.ParseDuration("1h")
	if interval == 0 {
		// failsave for 0s interval, to make machine still responsive
		interval = 1 * time.Second
	}
	log.Debug("Cleanup Timer, Interval: " + interval.String())
	// endless loop
	for {
		time.Sleep(interval)
		conf.RunCleanup()
	}
}
