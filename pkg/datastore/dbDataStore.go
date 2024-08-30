package datastore

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/thanhpk/randstr"
)

type DatabaseDataStore struct {
}

func (c DatabaseDataStore) GetPublicProfileNames() []string {
	return dbListPublicProfiles()
}

func (c DatabaseDataStore) GetAllProfileNames() []string {
	return dbListProfiles()
}

func (c DatabaseDataStore) ProfileExists(name string) bool {
	return dbProfileExists(name)
}

func (c DatabaseDataStore) GetProfileByName(name string) Profile {
	return *dbReadProfile(name)
}

func (c DatabaseDataStore) AddProfile(name string, sources []string, public bool, immutablePast bool) {
	dbWriteProfile(Profile{
		Name:          name,
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablePast,
		Tokens:        []Token{},
		Rules:         []Rule{},
	})
}

func (c DatabaseDataStore) EditProfile(name string, sources []string, public bool, immutablePast bool) {
	tempProfile := Profile{
		Name:          name,
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablePast,
	}

	dbWriteProfile(tempProfile)
}

func (c DatabaseDataStore) AddSource(profileName string, src string) error {
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbAddProfileSource(Profile{Name: profileName}, src)
	return nil
}

func (c DatabaseDataStore) RemoveSource(profileName string, src string) error {
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbRemoveProfileSourceByUrl(Profile{Name: profileName}, src)
	return nil
}

func (c DatabaseDataStore) AddRule(profileName string, rule Rule) error {
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbAddProfileRule(Profile{Name: profileName}, rule)
	return nil
}

func (c DatabaseDataStore) RemoveRule(profileName string, rule Rule) {
	dbRemoveRule(Profile{Name: profileName}, rule.Id)
}

func (c DatabaseDataStore) CreateToken(profileName string, note *string) error {
	token := randstr.Base62(64)
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbWriteProfileToken(Profile{Name: profileName}, token, note)
	return nil
}

func (c DatabaseDataStore) ModifyTokenNote(profileName string, token string, note *string) error {
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbWriteProfileToken(Profile{Name: profileName}, token, note)
	return nil
}

func (c DatabaseDataStore) DeleteToken(profileName string, token string) error {
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbRemoveProfileToken(Profile{Name: profileName}, token)
	return nil
}

func (c DatabaseDataStore) DeleteProfile(name string) {
	if dbProfileExists(name) {
		profile := *dbReadProfile(name)
		dbDeleteProfile(profile)
	}
}

func (c DatabaseDataStore) NotifierExists(name string) bool {
	return dbNotifierExists(name)
}

func (c DatabaseDataStore) AddNotifier(notifier Notifier) {
	dbWriteNotifier(notifier)
}

func (c DatabaseDataStore) AddNotifierFromProfile(profileName string, ownURL string) error {
	if !c.ProfileExists(profileName) {
		return fmt.Errorf("profile %s does not exist", profileName)
	}

	notifier := Notifier{
		Name:   profileName,
		Source: ownURL + "/profile/" + profileName,
	}
	dbWriteNotifier(notifier)
	return nil
}

func (c DatabaseDataStore) GetNotifiers() map[string]Notifier {
	notifiers := make(map[string]Notifier)
	for _, notifierName := range dbListNotifiers() {
		notifier, err := dbReadNotifier(notifierName, false)
		if err != nil {
			log.Warnf("`dbReadNotifier` failed with %s", err.Error())
		}
		notifiers[notifierName] = *notifier
	}
	return notifiers
}

func (c DatabaseDataStore) GetNotifier(notifierName string) Notifier {
	notifier, err := dbReadNotifier(notifierName, true)
	if err != nil {
		log.Warnf("`dbReadNotifier` faild with %s", err.Error())
	}
	return *notifier
}

func (c DatabaseDataStore) AddNotifyRecipient(notifierName string, recipient Recipient) error {
	if !dbNotifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	dbAddNotifierRecipient(Notifier{Name: notifierName}, recipient)
	return nil
}

func (c DatabaseDataStore) RemoveNotifyRecipient(notifierName string, recipient Recipient) error {
	if !dbNotifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	dbRemoveNotifierRecipient(Notifier{Name: notifierName}, recipient)
	return nil
}
