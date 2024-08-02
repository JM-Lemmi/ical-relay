module github.com/jm-lemmi/ical-relay/notifier

go 1.19

require (
	github.com/alexflint/go-arg v1.4.3
	github.com/arran4/golang-ical v0.2.4
	github.com/gopherlibs/feedhub v1.1.0
	github.com/jm-lemmi/ical-relay/compare v0.0.0-00010101000000-000000000000
	github.com/jm-lemmi/ical-relay/helpers v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/alexflint/go-scalar v1.2.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)

replace (
	github.com/arran4/golang-ical => ../../pkg/golang-ical
	github.com/gopherlibs/feedhub/feedhub => ../../pkg/feedhub
	github.com/jm-lemmi/ical-relay/compare => ../../pkg/compare
	github.com/jm-lemmi/ical-relay/helpers => ../../pkg/helpers
	github.com/jm-lemmi/ical-relay/modules => ../../pkg/modules
)
