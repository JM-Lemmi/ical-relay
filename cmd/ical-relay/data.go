package main

type DataStore interface {
	getPublicProfileNames() []string
	getAllProfileNames() []string
	profileExists(name string) bool
	// Note: Must check if profileExists beforehand
	getProfileByName(name string) profile
	addProfile(name string, sources []string, public bool, immutablePast bool) //TODO: make this take a profile type

	// editProfile edits a profile, not touching tokens and rules
	editProfile(name string, sources []string, public bool, immutablePast bool) //TODO: either make this take a profile type or split it into explicit editing functions
	addSource(profileName string, src string) error
	addRule(profileName string, rule Rule) error
	removeRule(profileName string, rule Rule) //editRule(string profileName, rule Rule)

	createToken(profileName string, note *string) error
	modifyTokenNote(profileName string, token string, note *string) error
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

type token struct {
	Token string  `db:"token"`
	Note  *string `db:"note"`
}

type profile struct {
	Name          string   `yaml:"name,omitempty" db:"name"`
	Sources       []string `yaml:"sources,omitempty"`
	Public        bool     `yaml:"public" db:"public"`
	ImmutablePast bool     `yaml:"immutable-past,omitempty" db:"immutable_past"`
	Tokens        []token  `yaml:"admin-tokens,omitempty"`
	Rules         []Rule   `yaml:"rules,omitempty"`
}

type Rule struct {
	id       int
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
