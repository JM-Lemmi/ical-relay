package main

type DataStore interface {
	getPublicCalendars() []string //TODO: rename to getPublicProfileNames
	getAllCalendars() []string    //TODO: rename to getAllProfileNames
	profileExists(name string) bool
	// Note: Must check if profileExists beforehand
	getProfileByName(name string) profile
	addProfile(name string, sources []string, public bool, immutablePast bool) //TODO: make this take a profile type

	// editProfile edits a profile, not touching tokens and rules
	editProfile(name string, sources []string, public bool, immutablePast bool) //TODO: either make this take a profile type or split it into explicit editing functions
	addSource(profileName string, src string) error
	addRule(profileName string, rule Rule) error
	removeRuleFromProfile(profile string, index int)

	createToken(profileName string, note string) error
	modifyTokenNote(profileName string, token string, note string) error
	deleteToken(profileName string, token string) error

	deleteProfile(name string)

	notifierExists(name string) bool
	addNotifier(notifier notifier)
	// getNotifiers returns a map of notifier (potentially WITHOUT the recipients populated!)
	getNotifiers() map[string]notifier
	getNotifier(notifierName string) notifier
	addNotifyRecipient(notifierName string, recipient string) error
	removeNotifyRecipient(notifierName string, recipient string) error
}
