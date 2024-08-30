package datastore

import (
	"fmt"
	"os"

	"github.com/thanhpk/randstr"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

type DataFile struct {
	Profiles  map[string]Profile  `yaml:"profiles,omitempty"`
	Notifiers map[string]Notifier `yaml:"notifiers,omitempty"`
}

// ParseDataFile reads data file from path and returns a DataFile struct
func ParseDataFile(path string) (DataFile, error) {
	var tmpConfig DataFile

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error Reading DataFile: %v ", err)
		return tmpConfig, err
	}

	err = yaml.Unmarshal(yamlFile, &tmpConfig)
	if err != nil {
		log.Fatalf("Error Parsing Data File: %v", err)
		return tmpConfig, err
	}

	return tmpConfig, nil
}

func ImportToDB(path string) error {
	// Parse data file
	data, err := ParseDataFile(path)
	if err != nil {
		log.Fatalf("Error Parsing Data File: %v", err)
		return err
	}

	// Check if db is initialized
	if db.DB == nil {
		log.Fatalf("DB not initialized")
		return fmt.Errorf("DB not initialized")
	}

	// import data into db
	for name, profile := range data.Profiles {
		// Write profile name into object
		profile.Name = name
		// Write the profile to the db, adding tokens and modules afterwards
		log.Info("Importing profile " + name)
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
	for name, notifier := range data.Notifiers {
		// Write notifier name into object
		notifier.Name = name
		log.Info("Importing notifier " + name)
		// Write the notifier to the db, adding recipients afterwards
		dbWriteNotifier(notifier)
		for _, recipient := range notifier.Recipients {
			dbAddNotifierRecipient(notifier, recipient)
		}
	}
	return nil
}

// CONFIG DataStore FUNCTIONS

func (c DataFile) GetPublicProfileNames() []string {
	var cal []string
	for p := range c.Profiles {
		if c.Profiles[p].Public {
			cal = append(cal, p)
		}
	}
	return cal
}

func (c DataFile) GetAllProfileNames() []string {
	var cal []string
	for p := range c.Profiles {
		log.Debug("Adding profile " + p + " to list")
		cal = append(cal, p)
	}
	return cal
}

func (c DataFile) ProfileExists(name string) bool {
	_, ok := c.Profiles[name]
	return ok
}

func (c DataFile) GetProfileByName(name string) Profile {
	c.populateRuleIds(name)
	return c.Profiles[name]
}

// add a profile without tokens and without rules
func (c DataFile) AddProfile(name string, sources []string, public bool, immutablepast bool) {
	c.Profiles[name] = Profile{
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablepast,
		Tokens:        []Token{},
		Rules:         []Rule{},
	}
}

func (c DataFile) EditProfile(name string, sources []string, public bool, immutablepast bool) {
	c.Profiles[name] = Profile{
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablepast,
		Tokens:        c.Profiles[name].Tokens,
		Rules:         c.Profiles[name].Rules,
	}
}

func (c DataFile) AddSource(profileName string, src string) error {
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	p.Sources = append(c.Profiles[profileName].Sources, src)
	c.Profiles[profileName] = p
	return nil
}

func (c DataFile) RemoveSource(profileName string, src string) error {
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

func (c DataFile) AddRule(profileName string, rule Rule) error {
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	rule.Id = len(c.Profiles[profileName].Rules)
	p.Rules = append(c.Profiles[profileName].Rules, rule)
	c.Profiles[profileName] = p
	return nil
}

func (c DataFile) RemoveRule(profileName string, rule Rule) {
	log.Info("Removing rule at position " + fmt.Sprint(rule.Id+1) + " from profile " + profileName)
	p := c.Profiles[profileName]
	p.Rules = append(p.Rules[:rule.Id], p.Rules[rule.Id+1:]...)
	c.Profiles[profileName] = p
	c.populateRuleIds(profileName)
}

func (c DataFile) CreateToken(profileName string, note *string) error {
	tokenString := randstr.Base62(64)
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	p := c.Profiles[profileName]
	p.Tokens = append(c.Profiles[profileName].Tokens, Token{
		Token: tokenString,
		Note:  note,
	})
	c.Profiles[profileName] = p
	return nil
}

func (c DataFile) ModifyTokenNote(profileName string, tokenString string, note *string) error {
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	for i := range c.Profiles[profileName].Tokens {
		c.Profiles[profileName].Tokens[i] = Token{Token: tokenString, Note: note}
	}
	return nil
}

func (c DataFile) DeleteToken(profileName string, token string) error {
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

func (c DataFile) DeleteProfile(name string) {
	delete(c.Profiles, name)
}

func (c DataFile) NotifierExists(name string) bool {
	_, ok := c.Notifiers[name]
	return ok
}

func (c DataFile) AddNotifier(notifier Notifier) {
	c.Notifiers[notifier.Name] = notifier
}

func (c DataFile) AddNotifierFromProfile(profileName string, ownURL string) error {
	// NOT IMPLEMENTED
	return fmt.Errorf("not implemented")
}

func (c DataFile) GetNotifiers() map[string]Notifier {
	return c.Notifiers
}

func (c DataFile) GetNotifier(notifierName string) Notifier {
	return c.Notifiers[notifierName]
}

func (c DataFile) AddNotifyRecipient(notifierName string, recipient Recipient) error {
	if !c.NotifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	n := c.Notifiers[notifierName]
	n.Recipients = append(n.Recipients, recipient)
	c.Notifiers[notifierName] = n
	return nil
}

func (c DataFile) RemoveNotifyRecipient(notifierName string, recipient Recipient) error {
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

func (c DataFile) populateRuleIds(profileName string) {
	p := c.Profiles[profileName]
	for id := range p.Rules {
		p.Rules[id].Id = id
	}
}
