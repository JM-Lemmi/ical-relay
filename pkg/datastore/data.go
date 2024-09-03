package datastore

import "time"

type DataStore interface {
	GetPublicProfileNames() []string
	GetAllProfileNames() []string
	ProfileExists(name string) bool
	// Note: Must check if profileExists beforehand
	GetProfileByName(name string) Profile
	AddProfile(name string, sources []string, public bool, immutablePast bool) //TODO: make this take a profile type

	// editProfile edits a profile, not touching tokens and rules
	EditProfile(name string, sources []string, public bool, immutablePast bool) //TODO: either make this take a profile type or split it into explicit editing functions
	AddSource(profileName string, src string) error
	// removeSource removes all sources with the given src string
	RemoveSource(profileName string, src string) error
	AddRule(profileName string, rule Rule) error
	RemoveRule(profileName string, rule Rule) //editRule(string profileName, rule Rule)

	CreateToken(profileName string, note *string) error
	ModifyTokenNote(profileName string, token string, note *string) error
	DeleteToken(profileName string, token string) error

	DeleteProfile(name string)

	NotifierExists(name string) bool
	AddNotifier(notifier Notifier)
	AddNotifierFromProfile(profileName string, ownURL string) error // takes the conf.Server.URL as ownURL, since its not a global variable in this package
	// getNotifiers returns a map of notifier (potentially WITHOUT the recipients populated!)
	GetNotifiers() map[string]Notifier
	GetNotifier(notifierName string) Notifier
	AddNotifyRecipient(notifierName string, recipient Recipient) error
	RemoveNotifyRecipient(notifierName string, recipient Recipient) error
	AddNotifierHistory(notifierName string, recipient string, historyType string, eventDate time.Time, changedDate time.Time, data string) error
}

type Token struct {
	Token string  `db:"token"`
	Note  *string `db:"note"`
}

type Profile struct {
	Name          string   `yaml:"name,omitempty" db:"name"`
	Sources       []string `yaml:"sources,omitempty"`
	Public        bool     `yaml:"public" db:"public"`
	ImmutablePast bool     `yaml:"immutable-past,omitempty" db:"immutable_past"`
	Tokens        []Token  `yaml:"admin-tokens,omitempty"`
	Rules         []Rule   `yaml:"rules,omitempty"`
}

type Rule struct {
	Id       int
	Filters  []map[string]string `yaml:"filters" json:"filters"`
	Operator string              `yaml:"operator" json:"operator"`
	Action   map[string]string   `yaml:"action" json:"action"`
	Expiry   string              `yaml:"expiry,omitempty" json:"expiry,omitempty"`
}

type Notifier struct {
	Name       string      `db:"name"`
	Source     string      `yaml:"source" db:"source"`
	Recipients []Recipient `yaml:"recipients"`
}

type Recipient struct {
	Type      string `yaml:"type" db:"type"`
	Recipient string `yaml:"recipient" db:"recipient"`
}

// Data Integrity Functions

// checks if a rule is valid.
// returns true if rule is valid, false if not
func (rule Rule) CheckRuleIntegrity() bool {
	// TODO implement!
	return true
}
