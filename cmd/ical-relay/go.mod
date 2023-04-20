module github.com/jm-lemmi/ical-relay/relay

go 1.19

require (
	github.com/arran4/golang-ical v0.0.0-20230318005454-19abf92700cc
	github.com/gorilla/mux v1.8.0
	github.com/juliangruber/go-intersect/v2 v2.0.1
	github.com/sirupsen/logrus v1.9.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v3 v3.0.1
)

require (
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)

replace (
	github.com/jm-lemmi/ical-relay/compare => ../../pkg/compare
)
