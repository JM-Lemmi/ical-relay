package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thanhpk/randstr"

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
	FaviconPath   string     `yaml:"faviconpath,omitempty"`
	Name          string     `yaml:"name,omitempty"`
	Imprint       string     `yaml:"imprintlink"`
	PrivacyPolicy string     `yaml:"privacypolicylink"`
	DB            dbConfig   `yaml:"db,omitempty"`
	Mail          mailConfig `yaml:"mail,omitempty"`
	SuperTokens   []string   `yaml:"super-tokens,omitempty"`
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

// CONFIG MANAGEMENT FUNCTIONS

// ParseConfig reads config from path and returns a Config struct
func ParseConfig(path string) (Config, error) {
	var tmpConfig Config

	yamlFile, err := os.ReadFile(path)
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
		}
		tmpConfig.saveConfig(path)
	}

	// check if config has current version, if not upgrade it
	if tmpConfig.Version < 3 {
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
	if tmpConfig.Server.Name == "" {
		tmpConfig.Server.Name = "Calendar"
	}
	if tmpConfig.Server.FaviconPath == "" {
		tmpConfig.Server.FaviconPath = "/static/media/favicon.svg"
	}

	return tmpConfig, nil
}

func (c Config) saveConfig(path string) error {
	currentConfig, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = os.WriteFile(
		path[:strings.LastIndexByte(path, '.')]+time.Now().UTC().Format("2006-01-02_150405")+".bak.yml",
		currentConfig,
		0600)
	if err != nil {
		return err
	}

	d, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, d, 0600)
}

func (c Config) importToDB() {
	// import data into db
	for name, profile := range c.Profiles {
		// Write profile name into object
		profile.Name = name
		// Write the profile to the db, adding tokens and modules afterwards
		log.Debug("Importing profile " + name)
		dbWriteProfile(profile)
		for _, source := range profile.Sources {
			if !dbProfileSourceExists(profile, source) {
				dbAddProfileSource(profile, source)
			}
		}
		for _, token := range profile.Tokens {
			dbWriteProfileToken(profile, token.Token, token.Note)
		}
		for _, rule := range profile.Rules {
			if !dbProfileRuleExists(profile, rule) {
				log.Debug("Adding rule " + rule.Action["type"])
				dbAddProfileRule(profile, rule)
			}
		}
	}
	for name, notifier := range c.Notifiers {
		// Write notifier name into object
		notifier.Name = name
		log.Debug("Importing notifier " + name)
		// Write the notifier to the db, adding recipients afterwards
		dbWriteNotifier(notifier)
		for _, recipient := range notifier.Recipients {
			dbAddNotifierRecipient(notifier, recipient)
		}
	}
}

// CONFIG DataStore FUNCTIONS

func (c Config) getPublicProfileNames() []string {
	var cal []string
	for p := range c.Profiles {
		if c.Profiles[p].Public {
			cal = append(cal, p)
		}
	}
	return cal
}

func (c Config) getAllProfileNames() []string {
	var cal []string
	for p := range c.Profiles {
		log.Debug("Adding profile " + p + " to list")
		cal = append(cal, p)
	}
	return cal
}

func (c Config) profileExists(name string) bool {
	_, ok := c.Profiles[name]
	return ok
}

func (c Config) getProfileByName(name string) profile {
	c.populateRuleIds(name)
	return c.Profiles[name]
}

// add a profile without tokens and without rules
func (c Config) addProfile(name string, sources []string, public bool, immutablepast bool) {
	c.Profiles[name] = profile{
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablepast,
		Tokens:        []token{},
		Rules:         []Rule{},
	}
}

func (c Config) editProfile(name string, sources []string, public bool, immutablepast bool) {
	c.Profiles[name] = profile{
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablepast,
		Tokens:        c.Profiles[name].Tokens,
		Rules:         c.Profiles[name].Rules,
	}
}

func (c Config) addSource(profileName string, src string) error {
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	p.Sources = append(c.Profiles[profileName].Sources, src)
	c.Profiles[profileName] = p
	return nil
}

func (c Config) addRule(profileName string, rule Rule) error {
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	rule.id = len(c.Profiles[profileName].Rules)
	p.Rules = append(c.Profiles[profileName].Rules, rule)
	c.Profiles[profileName] = p
	return nil
}

func (c Config) removeRule(profileName string, rule Rule) {
	log.Info("Removing rule at position " + fmt.Sprint(rule.id+1) + " from profile " + profileName)
	p := c.Profiles[profileName]
	p.Rules = append(p.Rules[:rule.id], p.Rules[rule.id+1:]...)
	c.Profiles[profileName] = p
	c.populateRuleIds(profileName)
}

func (c Config) createToken(profileName string, note *string) error {
	tokenString := randstr.Base62(64)
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	p.Tokens = append(c.Profiles[profileName].Tokens, token{
		Token: tokenString,
		Note:  note,
	})
	c.Profiles[profileName] = p
	return nil
}

func (c Config) modifyTokenNote(profileName string, tokenString string, note *string) error {
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	for i := range c.Profiles[profileName].Tokens {
		c.Profiles[profileName].Tokens[i] = token{Token: tokenString, Note: note}
	}
	return nil
}

func (c Config) deleteToken(profileName string, token string) error {
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	for i, cToken := range p.Tokens {
		if cToken.Token == token {
			p.Tokens = append(p.Tokens[:i], p.Tokens[i+1:]...)
			break
		}
	}
	c.Profiles[profileName] = p
	return nil
}

func (c Config) deleteProfile(name string) {
	delete(c.Profiles, name)
}

func (c Config) notifierExists(name string) bool {
	_, ok := c.Notifiers[name]
	return ok
}

func (c Config) addNotifier(notifier notifier) {
	c.Notifiers[notifier.Name] = notifier
}

func (c Config) getNotifiers() map[string]notifier {
	return c.Notifiers
}

func (c Config) getNotifier(notifierName string) notifier {
	return c.Notifiers[notifierName]
}

func (c Config) addNotifyRecipient(notifierName string, recipient string) error {
	if !c.notifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	n := c.Notifiers[notifierName]
	n.Recipients = append(n.Recipients, recipient)
	c.Notifiers[notifierName] = n
	return nil
}

func (c Config) removeNotifyRecipient(notifierName string, recipient string) error {
	if !c.notifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	n := c.Notifiers[notifierName]
	for i, r := range n.Recipients {
		if r == recipient {
			n.Recipients = append(n.Recipients[:i], n.Recipients[i+1:]...)
			c.Notifiers[notifierName] = n
			return nil
		}
	}
	return fmt.Errorf("recipient not found")
}

// internal helper functions

func (c Config) populateRuleIds(profileName string) {
	p := c.Profiles[profileName]
	for id := range p.Rules {
		p.Rules[id].id = id
	}
}

// LEGACY functions (to be moved elsewhere, or not supported yet by DataStore)

// TODO: move this function into the application code
func (c Config) addNotifierFromProfile(name string) {
	c.addNotifier(notifier{
		Name:       name,
		Source:     c.Server.URL + "/profiles/" + name,
		Interval:   "1h",
		Recipients: []string{},
	})
}

func (c Config) RunCleanup() {
	if db.DB != nil {
		log.Error("RunCleanup currently not supported on db") // TODO: implement
	} else {
		for p := range c.Profiles {
			for i, m := range c.Profiles[p].Rules {
				if m.Expiry != "" {
					exp, err := time.Parse(time.RFC3339, m.Expiry)
					if err != nil {
						log.Errorf("RunCleanup could not parse the expiry time: %s", err.Error())
					}
					if time.Now().After(exp) {
						c.removeRule(p, Rule{id: i})
					}
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

// CleanupStartup starts a heartbeat notifier in a sub-routine
func CleanupStartup() {
	log.Info("Starting Cleanup Timer")
	go TimeCleanup()
}

func TimeCleanup() {
	interval := time.Hour
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
