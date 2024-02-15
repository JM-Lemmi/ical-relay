module github.com/jm-lemmi/ical-relay/compare

go 1.22.0

require (
	github.com/arran4/golang-ical v0.2.4
	github.com/sirupsen/logrus v1.9.3
)

require (
	golang.org/x/sys v0.17.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/arran4/golang-ical => ../../pkg/golang-ical
