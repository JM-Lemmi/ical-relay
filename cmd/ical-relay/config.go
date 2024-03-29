package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thanhpk/randstr"

	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// STRUCTS
// !! breaking changes need to keep the old version in legacyconfig.go !!

// Config TODO: Eventually split into two parts: Config (possibly directly serverConfig) and Data (Profiles, Notifiers)
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

type token struct {
	Token string  `db:"token"`
	Note  *string `db:"note"`
}

type profile struct {
	Name          string   `db:"name"`
	Sources       []string `yaml:"sources,omitempty"`
	Public        bool     `yaml:"public" db:"public"`
	ImmutablePast bool     `yaml:"immutable-past,omitempty" db:"immutable_past"`
	Tokens        []string `yaml:"admin-tokens"`
	NTokens       []token  `yaml:"admin-tokens-storage-v2,omitempty"`
	Rules         []Rule   `yaml:"rules,omitempty"`
}

type Rule struct {
	Filters  []map[string]string `yaml:"filters" json:"filters"`
	Operator string              `yaml:"operator" json:"operator"`
	Action   map[string]string   `yaml:"action" json:"action"`
	Expiry   string              `yaml:"expiry,omitempty" json:"expiry,omitempty"`
}

type notifier struct {
	Name       string   `db:"name"`
	Source     string   `yaml:"source" db:"source"`
	Interval   string   `yaml:"interval" db:"interval"`
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
		// parsing in legacy mode succeeded
	}

	// check if config is up to date, if not
	if tmpConfig.Version < 2 {
		log.Warn("Config is outdated, upgrading")
		tmpConfig, err = LegacyParseConfig(path)
		if err != nil {
			log.Fatalf("Error Parsing Legacy Config: %v", err)
			return tmpConfig, err
		}
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

	if !helpers.DirectoryExists(tmpConfig.Server.StoragePath + "notifystore/") {
		log.Info("Creating notifystore directory")
		err = os.MkdirAll(tmpConfig.Server.StoragePath+"notifystore/", 0750)
		if err != nil {
			log.Fatalf("Error creating notifystore: %v", err)
			return tmpConfig, err
		}
	}
	if !helpers.DirectoryExists(tmpConfig.Server.StoragePath + "calstore/") {
		log.Info("Creating calstore directory")
		err = os.MkdirAll(tmpConfig.Server.StoragePath+"calstore/", 0750)
		if err != nil {
			log.Fatalf("Error creating calstore: %v", err)
			return tmpConfig, err
		}
	}

	return tmpConfig, nil
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
			dbWriteProfileToken(profile, token, nil)
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

func reloadConfig() error {
	// load config
	var err error
	conf, err = ParseConfig(configPath)
	if err != nil {
		return err
	} else {
		conf.importToDB()
		log.Info("Config reloaded")
		return nil
	}
	//TODO: clear caches
}

// CONFIG EDITING FUNCTIONS

func (c Config) getPublicCalendars() []string {
	if db.DB != nil {
		return dbListPublicProfiles()
	}
	var cal []string
	for p := range c.Profiles {
		if c.Profiles[p].Public {
			cal = append(cal, p)
		}
	}
	return cal
}

func (c Config) getAllCalendars() []string {
	if db.DB != nil {
		return dbListAllProfiles()
	}
	var cal []string
	for p := range c.Profiles {
		log.Debug("Adding profile " + p + " to list")
		cal = append(cal, p)
	}
	return cal
}

func (c Config) profileExists(name string) bool {
	if db.DB != nil {
		return dbProfileExists(name)
	}
	_, ok := c.Profiles[name]
	return ok
}

// add a profile without tokens and without rules
func (c Config) addProfile(name string, sources []string, public bool, immutablepast bool) {
	c.Profiles[name] = profile{
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablepast,
		Tokens:        []string{},
		Rules:         []Rule{},
	}
}

// edit a profile, keeping tokens and rules
func (c Config) editProfile(name string, sources []string, public bool, immutablepast bool) {
	c.Profiles[name] = profile{
		Name:          name,
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablepast,
		Tokens:        c.Profiles[name].Tokens,
		NTokens:       c.Profiles[name].NTokens,
		Rules:         c.Profiles[name].Rules,
	}
	if db.DB != nil {
		dbRemoveAllProfileSources(c.Profiles[name])
		dbWriteProfile(c.Profiles[name])
		for _, source := range c.Profiles[name].Sources {
			if !dbProfileSourceExists(c.Profiles[name], source) {
				dbAddProfileSource(c.Profiles[name], source)
			}
		}
	}
}

func (c Config) deleteProfile(name string) {
	delete(c.Profiles, name)
}

func (c Config) notifierExists(name string) bool {
	if db.DB != nil {
		return dbNotifierExists(name)
	}
	_, ok := c.Notifiers[name]
	return ok
}

// This is the hack that makes everything work currently
// TODO: remove
func (c Config) ensureProfileLoaded(name string) {
	if db.DB == nil {
		return
	}
	profile := dbReadProfile(name)
	if profile == nil {
		log.Fatal("Attempted to ensureProfileLoaded on an nonexistent profile")
	}
	c.Profiles[profile.Name] = *profile
}

func (c Config) addNotifierFromProfile(name string) {
	c.addNotifier(notifier{
		Name:       name,
		Source:     c.Server.URL + "/profiles/" + name,
		Interval:   "1h",
		Recipients: []string{},
	})
}

func (c Config) addNotifier(notifier notifier) {
	if db.DB != nil {
		dbWriteNotifier(notifier)
		return
	}
	c.Notifiers[notifier.Name] = notifier
}

// getNotifiers returns the a map of notifier WITHOUT the recipients populated
func (c Config) getNotifiers() map[string]notifier {
	if db.DB != nil {
		notifiers := make(map[string]notifier)
		for _, notifierName := range dbListNotifiers() {
			notifier, err := dbReadNotifier(notifierName, false)
			if err != nil {
				log.Warnf("`dbReadNotifier` failed with %s", err.Error())
			}
			notifiers[notifierName] = *notifier
		}
		return notifiers
	}
	return c.Notifiers
}

func (c Config) getNotifier(notifierName string) notifier {
	if db.DB != nil {
		notifier, err := dbReadNotifier(notifierName, true)
		if err != nil {
			log.Warnf("`dbReadNotifier` faild with %s", err.Error())
		}
		return *notifier
	}
	return c.Notifiers[notifierName]
}

func (c Config) addNotifyRecipient(notifierName string, recipient string) error {
	if !c.notifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	if db.DB != nil {
		dbAddNotifierRecipient(notifier{Name: notifierName}, recipient)
		return nil
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
	if db.DB != nil {
		dbRemoveNotifierRecipient(notifier{Name: notifierName}, recipient)
		return nil
	}
	n := c.Notifiers[notifierName]
	for i, r := range n.Recipients {
		if r == recipient {
			n.Recipients = append(n.Recipients[:i], n.Recipients[i+1:]...)
			c.Notifiers[notifierName] = n
			return nil
		}
	}
	return fmt.Errorf("recipient not found") //TODO: not supported on db
}

func (c Config) addRule(profileName string, rule Rule) error {
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	if db.DB != nil {
		dbAddProfileRule(profile{Name: profileName}, rule)
		return nil
	}
	p := c.Profiles[profileName]
	p.Rules = append(c.Profiles[profileName].Rules, rule)
	c.Profiles[profileName] = p
	return nil
}

func (c Config) removeRuleFromProfile(profile string, index int) {
	if db.DB != nil {
		log.Error("removeRuleFromProfile currently not supported on db")
		return
	}
	log.Info("Removing rule at position " + fmt.Sprint(index+1) + " from profile " + profile)
	p := c.Profiles[profile]
	p.Rules = append(p.Rules[:index], p.Rules[index+1:]...)
	c.Profiles[profile] = p
}

func (c Config) createToken(profileName string, note string) error {
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	token := randstr.Base62(64)
	if db.DB != nil {
		dbWriteProfileToken(profile{Name: profileName}, token, &note)
		return nil
	}
	p := c.Profiles[profileName]
	p.Tokens = append(c.Profiles[profileName].Tokens, token)
	c.Profiles[profileName] = p
	return nil
}

func (c Config) modifyTokenNote(profileName string, token string, note string) error {
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	if db.DB != nil {
		dbWriteProfileToken(profile{Name: profileName}, token, &note)
	}
	return nil
}

func (c Config) deleteToken(profileName string, token string) error {
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	if db.DB != nil {
		dbRemoveProfileToken(profile{Name: profileName}, token)
		return nil
	}
	p := c.Profiles[profileName]
	for i, cToken := range p.Tokens {
		if cToken == token {
			p.Tokens = append(p.Tokens[:i], p.Tokens[i+1:]...)
			break
		}
	}
	c.Profiles[profileName] = p
	return nil
}

func (c Config) addSource(profileName string, src string) error {
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	if db.DB != nil {
		dbAddProfileSource(profile{Name: profileName}, src)
		return nil
	}
	p := c.Profiles[profileName]
	p.Sources = append(c.Profiles[profileName].Sources, src)
	c.Profiles[profileName] = p
	return nil
}

func (c Config) RunCleanup() {
	for p := range c.Profiles {
		for i, m := range c.Profiles[p].Rules {
			if m.Expiry != "" {
				exp, err := time.Parse(time.RFC3339, m.Expiry)
				if err != nil {
					log.Errorf("RunCleanup could not parse the expiry time: %s", err.Error())
				}
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
