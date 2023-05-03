module github.com/jm-lemmi/ical-relay/relay

go 1.19

require (
	github.com/arran4/golang-ical v0.0.0-20230318005454-19abf92700cc
	github.com/fergusstrange/embedded-postgres v1.21.0
	github.com/gorilla/mux v1.8.0
	github.com/jm-lemmi/ical-relay/compare v0.0.0-00010101000000-000000000000
	github.com/jm-lemmi/ical-relay/helpers v0.0.0-00010101000000-000000000000
	github.com/jm-lemmi/ical-relay/modules v0.0.0-00010101000000-000000000000
	github.com/jmoiron/sqlx v1.3.5
	github.com/lib/pq v1.10.8
	github.com/sirupsen/logrus v1.9.0
	github.com/thanhpk/randstr v1.0.5
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/juliangruber/go-intersect/v2 v2.0.1

require (
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/sys v0.7.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)

replace (
	github.com/arran4/golang-ical => ../../pkg/golang-ical
	github.com/jm-lemmi/ical-relay/compare => ../../pkg/compare
	github.com/jm-lemmi/ical-relay/helpers => ../../pkg/helpers
	github.com/jm-lemmi/ical-relay/modules => ../../pkg/modules
)
