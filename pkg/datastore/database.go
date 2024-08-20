package datastore

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var db sqlx.DB

const CurrentDbVersion = 4

// startup connection function
func Connect(dbUser string, dbPassword string, dbHost string, dbName string) {
	userStr := ""
	if dbUser != "" {
		userStr = dbUser
		if dbPassword != "" {
			userStr += ":" + dbPassword
		}
		userStr += "@"
	}
	connStr := "postgresql://" + userStr + dbHost + "/" + dbName + "?sslmode=disable"
	log.Debug("Connecting to db using " + connStr)

	dbConn, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("Connection to db failed: %s", err)
		panic(err)
	}
	log.Debug("Connected to db")
	db = *dbConn

	var dbVersion int
	err = db.Get(&dbVersion, `SELECT MAX(version) FROM schema_upgrades`)
	if err != nil {
		log.Info("Initially creating tables...")
		initTables()
		setDbVersion(CurrentDbVersion)
		dbVersion = CurrentDbVersion
	}
	if dbVersion != CurrentDbVersion {
		doDbUpgrade(dbVersion)
	}
	log.Tracef("%#v", db)
}

//go:embed db.sql
var dbInitScript string

func initTables() {
	_, err := db.Exec(dbInitScript)
	if err != nil {
		log.Panic("Failed to execute db init script", err)
	}
}

func doDbUpgrade(fromDbVersion int) {
	log.Debugf("Upgrading db from version %d", fromDbVersion)
	if fromDbVersion < 2 {
		_, err := db.Exec("ALTER TABLE admin_tokens ADD COLUMN note text")
		if err != nil {
			panic("Failed to upgrade db")
		}
		setDbVersion(2)
	}
	if fromDbVersion < 3 {
		// We don't support a lossless upgrade from this db version due to no releases with versions older than 3
		log.Error("Unsupported db version, dropping module data")
		_, err := db.Exec("DROP TABLE module")
		if err != nil {
			panic("Failed to upgrade db")
		}
		_, err = db.Exec("ALTER TABLE profile DROP COLUMN source")
		if err != nil {
			panic("Failed to upgrade db")
		}
		initTables() //create new tables
		setDbVersion(3)
	}
	if fromDbVersion < 4 {
		log.Error("Unsupported db version, dropping profile source data and rules")
		_, err := db.Exec("DROP TABLE profile_sources; DROP TABLE source")
		if err != nil {
			//log.Panic("Failed to upgrade db", err)
		}
		_, err = db.Exec("DROP TABLE filter; DROP TABLE rule; DROP TABLE action")
		if err != nil {
			log.Panic("Failed to upgrade db", err)
		}
		initTables()
		setDbVersion(4)
	}
}

func setDbVersion(dbVersion int) {
	log.Infof("DB is now at version %d", dbVersion)
	_, err := db.Exec("INSERT INTO schema_upgrades(version) VALUES ($1)", dbVersion)
	if err != nil {
		log.Panicf(
			"Failed to set db version! If future restarts fail, you might have to manually set the version to %d",
			dbVersion)
	}
}

// these structs are only used for reading
type dbRule struct {
	Id               int        `db:"id"`
	Operator         string     `db:"operator"`
	ActionType       string     `db:"action_type"`
	ActionParameters string     `db:"action_parameters"`
	Expiry           *time.Time `db:"expiry"`
}

type dbFilter struct {
	FilterType string `db:"type"`
	Parameters string `db:"parameters"`
}

func dbProfileExists(profileName string) bool {
	var profileExists bool

	err := db.Get(&profileExists, `SELECT EXISTS (SELECT * FROM profile WHERE name = $1)`, profileName)
	log.Debug("Exec'd" + "SELECT EXISTS (SELECT * FROM profile WHERE name = " + profileName + ")")
	log.Tracef("%#v\n", profileExists)
	if err != nil {
		panic(err)
	}

	return profileExists
}

func dbListPublicProfiles() []string {
	var profiles []string

	err := db.Select(&profiles, `SELECT name FROM profile WHERE public=true`)
	if err != nil {
		panic(err)
	}

	return profiles
}

func dbListProfiles() []string {
	var profiles []string

	err := db.Select(&profiles, `SELECT name FROM profile`)
	if err != nil {
		panic(err)
	}

	return profiles
}

func dbReadProfile(profileName string) *Profile {
	profile := new(Profile)
	err := db.Get(profile, "SELECT name, public, immutable_past FROM profile WHERE name = $1", profileName)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Select(
		&profile.Sources,
		`SELECT url FROM source
JOIN profile_sources ps ON id = ps.source WHERE ps.profile = $1`,
		profileName)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Select(&profile.Tokens, "SELECT token, note FROM admin_tokens WHERE profile = $1", profileName)
	if err != nil {
		log.Fatal(err)
	}

	var dbRules []dbRule
	err = db.Select(
		&dbRules, "SELECT id, operator, action_type, action_parameters, expiry FROM rule WHERE profile = $1",
		profileName)
	if err != nil {
		log.Panic(err)
	}
	log.Tracef("%#v\n", dbRules)
	for _, dbRule := range dbRules {
		rule := new(Rule)
		rule.Id = dbRule.Id
		rule.Operator = dbRule.Operator
		if dbRule.Expiry != nil {
			rule.Expiry = dbRule.Expiry.Format(time.RFC3339)
		} else {
			rule.Expiry = ""
		}
		actionParameters := map[string]string{}
		err = json.Unmarshal([]byte(dbRule.ActionParameters), &actionParameters)
		if err != nil {
			log.Fatal(err)
		}
		actionParameters["type"] = dbRule.ActionType
		rule.Action = actionParameters

		var dbFilters []dbFilter
		err = db.Select(&dbFilters, "SELECT type, parameters FROM filter WHERE id = $1", dbRule.Id)
		if err != nil {
			log.Fatal(err)
		}
		for _, dbFilter := range dbFilters {
			filterParameters := map[string]string{}
			err = json.Unmarshal([]byte(dbFilter.Parameters), &filterParameters)
			if err != nil {
				log.Fatal(err)
			}
			filterParameters["type"] = dbFilter.FilterType
			rule.Filters = append(rule.Filters, filterParameters)
		}
		profile.Rules = append(profile.Rules, *rule)
	}
	log.Tracef("%#v\n", profile.Rules)
	return profile
}

// dbWriteProfile writes the profile options and sources to the db,
// silently overwriting if a profile with the same name exists.
func dbWriteProfile(profile Profile) {
	_, err := db.NamedExec(
		`INSERT INTO profile (name, public, immutable_past) VALUES (:name, :public, :immutable_past)
ON CONFLICT (name) DO UPDATE SET public = excluded.public, immutable_past = excluded.immutable_past`,
		profile)

	dbRemoveAllProfileSources(profile)
	for _, source := range profile.Sources {
		dbAddProfileSource(profile, source)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func dbDeleteProfile(profile Profile) {
	_, err := db.Exec(`DELETE FROM profile WHERE name=$1`, profile.Name)
	if err != nil {
		panic(err)
	}
	dbCleanupOrphanSources()
}

func dbProfileSourceExists(profile Profile, source string) bool {
	var profileSourceExists bool

	err := db.Get(&profileSourceExists, `SELECT EXISTS (SELECT * FROM source
JOIN profile_sources ps ON id = ps.source WHERE profile = $1 AND url = $2)`, profile.Name, source)
	if err != nil {
		panic(err)
	}

	return profileSourceExists
}

func dbAddProfileSource(profile Profile, source string) {
	var sourceId int
	err := db.Get(&sourceId, `INSERT INTO source (url) VALUES ($1) RETURNING id`, source)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = db.Exec(
		`INSERT INTO profile_sources (profile, source) VALUES ($1, $2) ON CONFLICT (profile, source) DO NOTHING`,
		profile.Name, sourceId)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func dbRemoveAllProfileSources(profile Profile) {
	log.Info(profile.Name)
	// Note: The following SQL statement is needlessly complicated due to the db supporting a n:n relation between
	// source and profile over profile_sources.
	// This n:n connection was used before db schema version 4, but since we support base64 events and can no longer
	// reasonably index on source.url, this capability is unused.
	// The db schema should be simplified and directly reference the profile from the source table, dropping the
	// profile_sources.
	_, err := db.Exec(`DELETE FROM source WHERE id IN (SELECT source FROM profile_sources WHERE profile = $1)`,
		profile.Name)
	if err != nil {
		log.Fatal(err)
		return
	}
}

// TODO: Expose sourceId, like we do for rules, to avoid this potentially expensive search
func dbRemoveProfileSourceByUrl(profile Profile, sourceUrl string) {
	var sourceIds []int
	err := db.Select(&sourceIds, `SELECT id FROM source WHERE url = $1`,
		sourceUrl)
	if err != nil {
		log.Fatal(err)
		return
	}
	if len(sourceIds) == 0 {
		log.Trace("no sources found for searchString sU:'", sourceUrl, "' p:'", profile.Name, "'")
	}
	for _, sourceId := range sourceIds {
		dbRemoveProfileSource(profile, sourceId) //Note: This will only remove the source if it belongs to the profile
	}
}

func dbRemoveProfileSource(profile Profile, sourceId int) {
	_, err := db.Exec(`DELETE FROM profile_sources WHERE profile = $1 AND source = $2`,
		profile.Name, sourceId)
	if err != nil {
		log.Fatal(err)
		return
	}
	dbCleanupOrphanSources()
}

func dbCleanupOrphanSources() {
	_, err := db.Exec(
		`DELETE FROM source WHERE (SELECT COUNT(*) FROM profile_sources WHERE profile_sources.source=source.id) < 1`,
	)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func dbRemoveAllProfileRules(profile Profile) {
	_, err := db.Exec(`DELETE FROM rule WHERE profile = $1`, profile.Name)
	if err != nil {
		log.Fatal(err)
		return
	}
}

// used for importing
func dbProfileRuleExists(profile Profile, rule Rule) bool {
	actionCopy := make(map[string]string)
	for k, v := range rule.Action {
		actionCopy[k] = v
	}
	actionType := actionCopy["type"]
	delete(actionCopy, "type")
	parametersJson, err := json.Marshal(actionCopy)
	if err != nil {
		panic(err)
	}

	var ruleIds []int
	if rule.Expiry != "" { // stored as true NULL in db
		err = db.Select(
			&ruleIds, `SELECT id FROM rule WHERE profile = $1 AND operator = $2
AND action_type = $3 AND action_parameters = $4 AND expiry = $5`,
			profile.Name, rule.Operator, actionType, parametersJson, rule.Expiry)
	} else {
		err = db.Select(
			&ruleIds, `SELECT id FROM rule WHERE profile = $1 AND operator = $2
AND action_type = $3 AND action_parameters = $4 AND expiry IS NULL`,
			profile.Name, rule.Operator, actionType, parametersJson)
	}
	if len(ruleIds) == 0 {
		log.Trace("rule not found with pN:'", profile.Name, "' rOp:'", rule.Operator,
			"' aT:'", actionType, "' aP:", string(parametersJson), " rE:'", rule.Expiry, "'")
		return false
	}
	if err != nil {
		log.Panic(err)
	}

	for _, ruleId := range ruleIds {
		ok := true
		for _, filter := range rule.Filters {
			filterCopy := make(map[string]string)
			for k, v := range filter {
				filterCopy[k] = v
			}
			filterType := filterCopy["type"]
			delete(filterCopy, "type")
			parametersJson, err := json.Marshal(filterCopy)
			if err != nil {
				panic(err)
			}

			var filterIsSame bool
			err = db.Get(
				&filterIsSame, `SELECT EXISTS (SELECT * FROM filter WHERE rule = $1 AND type = $2 AND parameters = $3)`,
				ruleId, filterType, parametersJson)
			if err != nil {
				log.Panic(err)
			}
			if !filterIsSame {
				log.Trace("filter does not match rId:", ruleId, " fT:'", filterType, "' pJ:", string(parametersJson))
				ok = false
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func dbAddProfileRule(profile Profile, rule Rule) {
	actionType := rule.Action["type"]
	delete(rule.Action, "type") //TODO: possibly deep-copy
	parametersJson, err := json.Marshal(rule.Action)
	if err != nil {
		panic(err)
	}

	var expiry sql.NullString
	expiry.String = rule.Expiry
	var ruleId int
	err = db.QueryRow(
		`INSERT INTO rule (profile, operator, action_type, action_parameters, expiry) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		profile.Name, rule.Operator, actionType, parametersJson, expiry).Scan(&ruleId)
	if err != nil {
		log.Panic(err)
	}

	for _, filter := range rule.Filters {
		dbAddRuleFilter(ruleId, filter)
	}
}

func dbAddRuleFilter(ruleId int, filter map[string]string) {
	filterType := filter["type"]
	delete(filter, "type") //TODO: possibly deep-copy
	parametersJson, err := json.Marshal(filter)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(
		`INSERT INTO filter (rule, type, parameters) VALUES ($1, $2, $3) RETURNING id`,
		ruleId, filterType, parametersJson)
	if err != nil {
		panic(err)
	}
}

func dbRemoveRule(profile Profile, ruleId int) {
	//TODO: ignore profile passed here, ruleIds are unique
	_, err := db.Exec(
		`DELETE FROM rule WHERE profile=$1 AND id=$2`,
		profile.Name, ruleId)
	if err != nil {
		panic(err)
	}
}

func dbWriteProfileToken(profile Profile, token string, note *string) {
	if len(token) != 64 {
		log.Fatal("Only 64-byte tokens are allowed!")
	}
	_, err := db.Exec(
		`INSERT INTO admin_tokens (profile, token, note) VALUES ($1, $2, $3)
ON CONFLICT (token) DO UPDATE SET note = excluded.note`,
		profile.Name, token, note)
	if err != nil {
		panic(err)
	}
}

func dbRemoveProfileToken(profile Profile, token string) {
	//TODO: ignore profile passed here, tokens are unique
	_, err := db.Exec(
		`DELETE FROM admin_tokens WHERE profile=$1 AND token=$2`,
		profile.Name, token)
	if err != nil {
		panic(err)
	}
}

// Notifier

func dbNotifierExists(notifierName string) bool {
	var notifierExists bool

	err := db.Get(&notifierExists, `SELECT EXISTS (SELECT * FROM notifier WHERE name = $1)`, notifierName)
	if err != nil {
		panic(err)
	}

	return notifierExists
}

func dbListNotifiers() []string {
	var notifiers []string

	err := db.Select(&notifiers, `SELECT name FROM notifier`)
	if err != nil {
		panic(err)
	}

	return notifiers
}

func dbReadNotifier(notifierName string, fetchRecipients bool) (*Notifier, error) {
	readNotifier := new(Notifier)
	tx, err := db.Beginx()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	_, err = tx.Exec("SET intervalstyle = 'iso_8601'")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = tx.Get(readNotifier, "SELECT name, source, \"interval\" FROM notifier WHERE name = $1", notifierName)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	duration, err :=
		time.ParseDuration(strings.ToLower(strings.Split(readNotifier.Interval, "T")[1]))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	//log.Tracef("%#v\n", duration.String())
	readNotifier.Interval = duration.String()

	if fetchRecipients {
		err = db.Select(
			&readNotifier.Recipients,
			`SELECT email FROM recipients
JOIN notifier_recipients nr ON email = nr.recipient
JOIN notifier ON nr.notifier = notifier.name WHERE nr.notifier = $1`,
			notifierName)
	}
	return readNotifier, nil
}

// dbWriteNotifier writes the notifier to the db, silently overwriting if a notifier with the same name exists.
// Important: Does not write the notifier recipients to db! This has to be done manually through dbAddNotifierRecipient!
func dbWriteNotifier(notifier Notifier) {
	_, err := db.NamedExec(
		`INSERT INTO notifier (name, source, interval) VALUES (:name, :source, :interval)
ON CONFLICT (name) DO UPDATE SET source = excluded.source, interval = excluded.interval`,
		notifier)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func dbDeleteNotifier(notifier Notifier) {
	_, err := db.Exec(`DELETE FROM notifier WHERE name=$1`, notifier.Name)
	if err != nil {
		panic(err)
	}
	dbCleanupOrphanRecipients()
}

func dbCleanupOrphanRecipients() {
	_, err := db.Exec(
		`DELETE FROM recipients WHERE
        	(SELECT COUNT(*) FROM notifier_recipients WHERE notifier_recipients.recipient=recipients.email) < 1`,
	)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func dbAddNotifierRecipient(notifier Notifier, recipient string) {
	_, err := db.Exec(`INSERT INTO recipients (email) VALUES ($1) ON CONFLICT (email) DO NOTHING`, recipient)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = db.Exec(
		`INSERT INTO notifier_recipients (notifier, recipient) VALUES ($1, $2)
ON CONFLICT (notifier, recipient) DO NOTHING`,
		notifier.Name, recipient)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func dbRemoveNotifierRecipient(notifier Notifier, recipient string) {
	_, err := db.Exec(`DELETE FROM notifier_recipients WHERE notifier = $1 AND recipient = $2`,
		notifier.Name, recipient)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func dbRemoveRecipient(recipient string) {
	//It is too expensive to go through all cached notifiers and remove the recipient, we simply invalidate the cache.
	//(The db is much more efficient doing a cascading deletion)
	//TODO: invalidate cache
	_, err := db.Exec(`DELETE FROM recipients WHERE email = $1`, recipient)
	if err != nil {
		log.Fatal(err)
		return
	}
}

// TODO: Database Cleanup somewhere
