package main

import (
	"fmt"
	"github.com/thanhpk/randstr"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// STRUCTS

type profile struct {
	Name          string              `db:"name"`
	Source        string              `yaml:"source" db:"source"`
	Public        bool                `yaml:"public" db:"public"`
	ImmutablePast bool                `yaml:"immutable-past,omitempty" db:"immutable_past"`
	Tokens        []string            `yaml:"admin-tokens"`
	NTokens       []token             `yaml:"admin-tokens-storage-v2,omitempty"`
	Modules       []map[string]string `yaml:"modules,omitempty"`
}

type token struct {
	Token string  `db:"token"`
	Note  *string `db:"note"`
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

type serverConfig struct {
	Addr          string     `yaml:"addr"`
	URL           string     `yaml:"url"`
	LogLevel      log.Level  `yaml:"loglevel"`
	StoragePath   string     `yaml:"storagepath"`
	TemplatePath  string     `yaml:"templatepath"`
	Imprint       string     `yaml:"imprintlink"`
	PrivacyPolicy string     `yaml:"privacypolicylink"`
	DB            dbConfig   `yaml:"db,omitempty"`
	Mail          mailConfig `yaml:"mail,omitempty"`
	SuperTokens   []string   `yaml:"super-tokens,omitempty"`
}

type notifier struct {
	Name       string   `db:"name"`
	Source     string   `yaml:"source" db:"source"`
	Interval   string   `yaml:"interval" db:"interval"`
	Recipients []string `yaml:"recipients"`
}

// Config TODO: Eventually split into two parts: Config (possibly directly serverConfig) and Data (Profiles, Notifiers)
// Config represents configuration for the application
type Config struct {
	Server    serverConfig        `yaml:"server"`
	Profiles  map[string]profile  `yaml:"profiles,omitempty"`
	Notifiers map[string]notifier `yaml:"notifiers,omitempty"`
}

// CONFIG MANAGEMENT FUNCTIONS

// ParseConfig reads config from path and returns a Config struct
func ParseConfig(path string, importData bool) (Config, error) {
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

	return tmpConfig, nil
}

func (c Config) importToDB() {
	// import data into db
	for name, profile := range c.Profiles {
		// Write profile name into object
		profile.Name = name
		// Write the profile to the db, adding tokens and profiles afterwards
		log.Debug("Importing profile " + name)
		dbWriteProfile(profile)
		for _, token := range profile.Tokens {
			dbWriteProfileToken(profile, token, nil)
		}
		for _, module := range profile.Modules {
			dbAddProfileModule(profile, module)
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
	conf, err = ParseConfig(configPath, false)
	if err != nil {
		return err
	} else {
		log.Info("Config reloaded")
		return nil
	}
	//TODO: clear caches
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

func (c Config) profileExists(name string) bool {
	if db.DB != nil {
		return dbProfileExists(name)
	}
	_, ok := c.Profiles[name]
	return ok
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
		var notifiers map[string]notifier
		for _, notifierName := range dbListNotifiers() {
			notifier, _ := dbReadNotifier(notifierName, false)
			notifiers[notifierName] = *notifier
		}
		return notifiers
	}
	return c.Notifiers
}

func (c Config) getNotifier(notifierName string) notifier {
	if db.DB != nil {
		notifier, _ := dbReadNotifier(notifierName, true)
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
	return c.saveConfig(configPath)
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
			return c.saveConfig(configPath)
		}
	}
	return fmt.Errorf("recipient not found") //TODO: not supported on db
}

func (c Config) addModule(profileName string, module map[string]string) error {
	if !c.profileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	if db.DB != nil {
		dbAddProfileModule(profile{Name: profileName}, module)
		return nil
	}
	p := c.Profiles[profileName]
	p.Modules = append(c.Profiles[profileName].Modules, module)
	c.Profiles[profileName] = p
	return c.saveConfig(configPath)
}

func (c Config) removeModuleFromProfile(profile string, index int) {
	if db.DB != nil {
		log.Error("removeModuleFromProfile currently not supported on db")
		return
	}
	log.Info("Removing expired module at position " + fmt.Sprint(index+1) + " from profile " + profile)
	removeFromMapString(c.Profiles[profile].Modules, index)
	c.saveConfig(configPath)
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
	return c.saveConfig(configPath)
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
	return c.saveConfig(configPath)
}

func (c Config) RunCleanup() {
	for p := range c.Profiles {
		for i, m := range c.Profiles[p].Modules {
			if m["expires"] != "" {
				exp, _ := time.Parse(time.RFC3339, m["expiration"])
				if time.Now().After(exp) {
					c.removeModuleFromProfile(p, i)
				}
			}
		}
	}
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
