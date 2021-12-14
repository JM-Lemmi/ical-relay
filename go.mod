module github.com/thor77/ical-relay

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/GeertJohan/go.rice v1.0.0
	github.com/arran4/golang-ical v0.0.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.0
	github.com/sirupsen/logrus v1.4.0
)

replace github.com/arran4/golang-ical => /app/golang-ical
