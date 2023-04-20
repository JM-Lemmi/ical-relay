module github.com/jm-lemmi/ical-relay/modules

go 1.19

require (
	github.com/arran4/golang-ical v0.0.0-20230318005454-19abf92700cc
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/jm-lemmi/ical-relay/helpers v0.0.0-00010101000000-000000000000
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)

replace github.com/jm-lemmi/ical-relay/helpers => ../helpers
