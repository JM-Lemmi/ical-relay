package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jm-lemmi/ical-relay/database"

	"github.com/thanhpk/randstr"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// STRUCTS
// !! breaking changes need to keep the old version in legacyconfig.go !!

// Config represents configuration for the application
type Config struct {
	Version   int                          `yaml:"version"`
	Server    serverConfig                 `yaml:"server"`
	Profiles  map[string]database.Profile  `yaml:"profiles,omitempty"`
	Notifiers map[string]database.Notifier `yaml:"notifiers,omitempty"`
}

type serverConfig struct {
	Addr            string     `yaml:"addr"`
	URL             string     `yaml:"url"`
	LogLevel        log.Level  `yaml:"loglevel"`
	StoragePath     string     `yaml:"storagepath"`
	LiteMode        bool       `yaml:"litemode,omitempty"`
	DisableFrontend bool       `yaml:"disable-frontend,omitempty"`
	TemplatePath    string     `yaml:"templatepath"`
	FaviconPath     string     `yaml:"faviconpath,omitempty"`
	Name            string     `yaml:"name,omitempty"`
	Imprint         string     `yaml:"imprintlink"`
	PrivacyPolicy   string     `yaml:"privacypolicylink"`
	DB              dbConfig   `yaml:"db,omitempty"`
	Mail            mailConfig `yaml:"mail,omitempty"`
	SuperTokens     []string   `yaml:"super-tokens,omitempty"`
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
		database.DbWriteProfile(profile)
		for _, source := range profile.Sources {
			if !database.DbProfileSourceExists(profile, source) {
				database.DbAddProfileSource(profile, source)
			}
		}
		for _, token := range profile.Tokens {
			database.DbWriteProfileToken(profile, token.Token, token.Note)
		}
		for _, rule := range profile.Rules {
			if !database.DbProfileRuleExists(profile, rule) {
				log.Debug("Adding rule " + rule.Action["type"])
				database.DbAddProfileRule(profile, rule)
			}
		}
	}
	for name, notifier := range c.Notifiers {
		// Write notifier name into object
		notifier.Name = name
		log.Debug("Importing notifier " + name)
		// Write the notifier to the db, adding recipients afterwards
		database.DbWriteNotifier(notifier)
		for _, recipient := range notifier.Recipients {
			database.DbAddNotifierRecipient(notifier, recipient)
		}
	}
}

// CONFIG DataStore FUNCTIONS

func (c Config) GetPublicProfileNames() []string {
	var cal []string
	for p := range c.Profiles {
		if c.Profiles[p].Public {
			cal = append(cal, p)
		}
	}
	return cal
}

func (c Config) GetAllProfileNames() []string {
	var cal []string
	for p := range c.Profiles {
		log.Debug("Adding profile " + p + " to list")
		cal = append(cal, p)
	}
	return cal
}

func (c Config) ProfileExists(name string) bool {
	_, ok := c.Profiles[name]
	return ok
}

func (c Config) GetProfileByName(name string) database.Profile {
	c.populateRuleIds(name)
	return c.Profiles[name]
}

// add a profile without tokens and without rules
func (c Config) AddProfile(name string, sources []string, public bool, immutablepast bool) {
	c.Profiles[name] = database.Profile{
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablepast,
		Tokens:        []database.Token{},
		Rules:         []database.Rule{},
	}
}

func (c Config) EditProfile(name string, sources []string, public bool, immutablepast bool) {
	c.Profiles[name] = database.Profile{
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablepast,
		Tokens:        c.Profiles[name].Tokens,
		Rules:         c.Profiles[name].Rules,
	}
}

func (c Config) AddSource(profileName string, src string) error {
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	p.Sources = append(c.Profiles[profileName].Sources, src)
	c.Profiles[profileName] = p
	return nil
}

func (c Config) RemoveSource(profileName string, src string) error {
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	for i, cSource := range p.Sources {
		if cSource == src {
			p.Sources = append(p.Sources[:i], p.Sources[i+1:]...)
		}
	}
	c.Profiles[profileName] = p
	return nil
}

func (c Config) AddRule(profileName string, rule database.Rule) error {
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	rule.Id = len(c.Profiles[profileName].Rules)
	p.Rules = append(c.Profiles[profileName].Rules, rule)
	c.Profiles[profileName] = p
	return nil
}

func (c Config) RemoveRule(profileName string, rule database.Rule) {
	log.Info("Removing rule at position " + fmt.Sprint(rule.Id+1) + " from profile " + profileName)
	p := c.Profiles[profileName]
	p.Rules = append(p.Rules[:rule.Id], p.Rules[rule.Id+1:]...)
	c.Profiles[profileName] = p
	c.populateRuleIds(profileName)
}

func (c Config) CreateToken(profileName string, note *string) error {
	tokenString := randstr.Base62(64)
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	p.Tokens = append(c.Profiles[profileName].Tokens, database.Token{
		Token: tokenString,
		Note:  note,
	})
	c.Profiles[profileName] = p
	return nil
}

func (c Config) ModifyTokenNote(profileName string, tokenString string, note *string) error {
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	for i := range c.Profiles[profileName].Tokens {
		c.Profiles[profileName].Tokens[i] = database.Token{Token: tokenString, Note: note}
	}
	return nil
}

func (c Config) DeleteToken(profileName string, token string) error {
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	for i, cToken := range p.Tokens {
		if cToken.Token == token {
			p.Tokens = append(p.Tokens[:i], p.Tokens[i+1:]...)
		}
	}
	c.Profiles[profileName] = p
	return nil
}

func (c Config) DeleteProfile(name string) {
	delete(c.Profiles, name)
}

func (c Config) NotifierExists(name string) bool {
	_, ok := c.Notifiers[name]
	return ok
}

func (c Config) AddNotifier(notifier database.Notifier) {
	c.Notifiers[notifier.Name] = notifier
}

func (c Config) GetNotifiers() map[string]database.Notifier {
	return c.Notifiers
}

func (c Config) GetNotifier(notifierName string) database.Notifier {
	return c.Notifiers[notifierName]
}

func (c Config) AddNotifyRecipient(notifierName string, recipient string) error {
	if !c.NotifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	n := c.Notifiers[notifierName]
	n.Recipients = append(n.Recipients, recipient)
	c.Notifiers[notifierName] = n
	return nil
}

func (c Config) RemoveNotifyRecipient(notifierName string, recipient string) error {
	if !c.NotifierExists(notifierName) {
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
		p.Rules[id].Id = id
	}
}

// LEGACY functions (to be moved elsewhere, or not supported yet by DataStore)

// TODO: move this function into the application code
func (c Config) addNotifierFromProfile(name string) {
	c.AddNotifier(database.Notifier{
		Name:       name,
		Source:     c.Server.URL + "/profiles/" + name,
		Interval:   "1h",
		Recipients: []string{},
	})
}

func (c Config) RunCleanup() {
	if database.Db.DB != nil {
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
						c.RemoveRule(p, database.Rule{Id: i})
					}
				}
			}
		}
	}
}

// checks if a rule is valid.
// returns true if rule is valid, false if not
func checkRuleIntegrity(rule database.Rule) bool {
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
