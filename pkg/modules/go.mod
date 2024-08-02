module github.com/jm-lemmi/ical-relay/modules

go 1.19

require (
	github.com/arran4/golang-ical v0.2.4
	github.com/sirupsen/logrus v1.9.3
)

require github.com/gopherlibs/feedhub v1.1.0 // indirect

require (
	github.com/jm-lemmi/ical-relay/helpers v0.0.0-00010101000000-000000000000
	golang.org/x/sys v0.17.0 // indirect
)

replace (
	github.com/arran4/golang-ical => ../../pkg/golang-ical
	github.com/gopherlibs/feedhub/feedhub => ../../pkg/feedhub
	github.com/jm-lemmi/ical-relay/helpers => ../helpers
)
