module github.com/jm-lemmi/ical-relay/compare

go 1.19

require (
	github.com/arran4/golang-ical v0.0.0-20230318005454-19abf92700cc
	github.com/sirupsen/logrus v1.9.0
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/arran4/golang-ical => ../../pkg/golang-ical
