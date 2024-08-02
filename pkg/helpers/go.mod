module github.com/jm-lemmi/ical-relay/helpers

go 1.19

require (
	github.com/arran4/golang-ical v0.2.4
	github.com/gopherlibs/feedhub v1.1.0
)

require gopkg.in/yaml.v3 v3.0.1 // indirect

replace (
	github.com/arran4/golang-ical => ../../pkg/golang-ical
	github.com/gopherlibs/feedhub/feedhub => ../../pkg/feedhub
)
