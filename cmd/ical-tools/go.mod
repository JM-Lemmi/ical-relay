module github.com/jm-lemmi/ical-relay/tools

go 1.18

require (
	github.com/alexflint/go-arg v1.5.1
	github.com/arran4/golang-ical v0.2.4
	github.com/jm-lemmi/ical-relay/compare v0.0.0-00010101000000-000000000000
	github.com/jm-lemmi/ical-relay/helpers v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/alexflint/go-scalar v1.2.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
)

replace (
	github.com/arran4/golang-ical => ../../pkg/golang-ical
	github.com/jm-lemmi/ical-relay/compare => ../../pkg/compare
	github.com/jm-lemmi/ical-relay/helpers => ../../pkg/helpers
	github.com/jm-lemmi/ical-relay/modules => ../../pkg/modules
)
