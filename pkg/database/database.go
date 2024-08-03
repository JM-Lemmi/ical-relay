package database

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var db sqlx.DB

//go:embed migrations/*.sql
var migrations embed.FS

// startup connection function
func ConnectAndUpgradeDB(dbUser string, dbPassword string, dbHost string, dbName string) sqlx.DB {
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

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("error on creating postgres driver: %s", err)
	}

	migrationDriver, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Fatalf("error on creating migration embed driver: %s", err)
	}

	m, err := migrate.NewWithInstance("iofs", migrationDriver, "postgres", driver)

	if err != nil {
		log.Fatalf("error on creating migrate instance: %s", err)
	}

	var dbVersion int
	err = db.Get(&dbVersion, `SELECT MAX(version) FROM schema_upgrades`)
	if err == nil {
		log.Info("Found legacy database version and force setting migrate version...")
		m.Force(dbVersion)
	}

	log.Info("Running database migrations")
	err = m.Up()
	if err != nil {
		log.Fatalf("Error while doing database upgrade: %s", err)
	}
	return db
}
