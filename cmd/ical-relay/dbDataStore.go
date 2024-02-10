package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/thanhpk/randstr"
)

type DatabaseDataStore struct {
}

func (c DatabaseDataStore) getProfileByName(name string) profile {
	return *dbReadProfile(name)
}

func (c DatabaseDataStore) getPublicCalendars() []string {
	return dbListPublicProfiles()
}

func (c DatabaseDataStore) getAllCalendars() []string {
	return dbListAllProfiles()
}

func (c DatabaseDataStore) profileExists(name string) bool {
	return dbProfileExists(name)
}

func (c DatabaseDataStore) addProfile(name string, sources []string, public bool, immutablePast bool) {
	dbWriteProfile(profile{
		Name:          name,
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablePast,
		Tokens:        []string{},
		Rules:         []Rule{},
	})
}

func (c DatabaseDataStore) editProfile(name string, sources []string, public bool, immutablePast bool) {
	tempProfile := profile{
		Name:          name,
		Sources:       sources,
		Public:        public,
		ImmutablePast: immutablePast,
	}

	dbRemoveAllProfileSources(tempProfile)
	dbWriteProfile(tempProfile)
	for _, source := range tempProfile.Sources {
		if !dbProfileSourceExists(tempProfile, source) {
			dbAddProfileSource(tempProfile, source)
		}
	}
}

func (c DatabaseDataStore) deleteProfile(name string) {
	panic("DB Delete Profile not implemented yet!") // TODO: implement
}

func (c DatabaseDataStore) notifierExists(name string) bool {
	return dbNotifierExists(name)
}

func (c DatabaseDataStore) addNotifier(notifier notifier) {
	dbWriteNotifier(notifier)
}

func (c DatabaseDataStore) getNotifiers() map[string]notifier {
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

func (c DatabaseDataStore) getNotifier(notifierName string) notifier {
	notifier, err := dbReadNotifier(notifierName, true)
	if err != nil {
		log.Warnf("`dbReadNotifier` faild with %s", err.Error())
	}
	return *notifier
}

func (c DatabaseDataStore) addNotifyRecipient(notifierName string, recipient string) error {
	if !dbNotifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	dbAddNotifierRecipient(notifier{Name: notifierName}, recipient)
	return nil
}

func (c DatabaseDataStore) removeNotifyRecipient(notifierName string, recipient string) error {
	if !dbNotifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	dbRemoveNotifierRecipient(notifier{Name: notifierName}, recipient)
	return nil
}

func (c DatabaseDataStore) addRule(profileName string, rule Rule) error {
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbAddProfileRule(profile{Name: profileName}, rule)
	return nil
}

func (c DatabaseDataStore) removeRuleFromProfile(profile string, index int) {
	log.Error("removeRuleFromProfile currently not supported on db") // TODO: implement
}

func (c DatabaseDataStore) createToken(profileName string, note string) error {
	token := randstr.Base62(64)
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbWriteProfileToken(profile{Name: profileName}, token, &note)
	return nil
}

func (c DatabaseDataStore) modifyTokenNote(profileName string, token string, note string) error {
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbWriteProfileToken(profile{Name: profileName}, token, &note)
	return nil
}

func (c DatabaseDataStore) deleteToken(profileName string, token string) error {
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbRemoveProfileToken(profile{Name: profileName}, token)
	return nil
}

func (c DatabaseDataStore) addSource(profileName string, src string) error {
	if !dbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	dbAddProfileSource(profile{Name: profileName}, src)
	return nil
}
