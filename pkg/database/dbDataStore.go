package database

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/thanhpk/randstr"
)

type DatabaseDataStore struct {
}

func (c DatabaseDataStore) GetPublicProfileNames() []string {
	return DbListPublicProfiles()
}

func (c DatabaseDataStore) GetAllProfileNames() []string {
	return DbListProfiles()
}

func (c DatabaseDataStore) ProfileExists(name string) bool {
	return DbProfileExists(name)
}

func (c DatabaseDataStore) GetProfileByName(name string) Profile {
	return *DbReadProfile(name)
}

func (c DatabaseDataStore) AddProfile(name string, sources []string, public bool, immutablePast bool) {
	DbWriteProfile(Profile{
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

	DbWriteProfile(tempProfile)
}

func (c DatabaseDataStore) AddSource(profileName string, src string) error {
	if !DbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	DbAddProfileSource(Profile{Name: profileName}, src)
	return nil
}

func (c DatabaseDataStore) RemoveSource(profileName string, src string) error {
	if !DbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	DbRemoveProfileSourceByUrl(Profile{Name: profileName}, src)
	return nil
}

func (c DatabaseDataStore) AddRule(profileName string, rule Rule) error {
	if !DbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	DbAddProfileRule(Profile{Name: profileName}, rule)
	return nil
}

func (c DatabaseDataStore) RemoveRule(profileName string, rule Rule) {
	DbRemoveRule(Profile{Name: profileName}, rule.Id)
}

func (c DatabaseDataStore) CreateToken(profileName string, note *string) error {
	token := randstr.Base62(64)
	if !DbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	DbWriteProfileToken(Profile{Name: profileName}, token, note)
	return nil
}

func (c DatabaseDataStore) ModifyTokenNote(profileName string, token string, note *string) error {
	if !DbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	DbWriteProfileToken(Profile{Name: profileName}, token, note)
	return nil
}

func (c DatabaseDataStore) DeleteToken(profileName string, token string) error {
	if !DbProfileExists(profileName) {
		return fmt.Errorf("profile " + profileName + " does not exist")
	}
	DbRemoveProfileToken(Profile{Name: profileName}, token)
	return nil
}

func (c DatabaseDataStore) DeleteProfile(name string) {
	panic("DB Delete Profile not implemented yet!") // TODO: implement
}

func (c DatabaseDataStore) NotifierExists(name string) bool {
	return DbNotifierExists(name)
}

func (c DatabaseDataStore) AddNotifier(notifier Notifier) {
	DbWriteNotifier(notifier)
}

func (c DatabaseDataStore) GetNotifiers() map[string]Notifier {
	notifiers := make(map[string]Notifier)
	for _, notifierName := range DbListNotifiers() {
		notifier, err := DbReadNotifier(notifierName, false)
		if err != nil {
			log.Warnf("`dbReadNotifier` failed with %s", err.Error())
		}
		notifiers[notifierName] = *notifier
	}
	return notifiers
}

func (c DatabaseDataStore) GetNotifier(notifierName string) Notifier {
	notifier, err := DbReadNotifier(notifierName, true)
	if err != nil {
		log.Warnf("`dbReadNotifier` faild with %s", err.Error())
	}
	return *notifier
}

func (c DatabaseDataStore) AddNotifyRecipient(notifierName string, recipient string) error {
	if !DbNotifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	DbAddNotifierRecipient(Notifier{Name: notifierName}, recipient)
	return nil
}

func (c DatabaseDataStore) RemoveNotifyRecipient(notifierName string, recipient string) error {
	if !DbNotifierExists(notifierName) {
		return fmt.Errorf("notifier does not exist")
	}
	DbRemoveNotifierRecipient(Notifier{Name: notifierName}, recipient)
	return nil
}
