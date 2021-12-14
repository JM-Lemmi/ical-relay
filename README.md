ical-relay
==========
Relay ical event url and exclude events based on a regex.

# Usage
* Run from source: `go run .`
* Build and run: `rice embed-go && go build . && ./ical-relay`

Access filtered ical file on `server:8080/profiles/profilename`

Add `config.toml` to executing directory for configuration options.

All events in `addical.ics` will be added to the filtered ical.

# Config
```toml
url = "https://example.com/events.ical"

[server]
addr = ":8080"
loglevel = "info"

[profiles]
    [profiles.profilename]
    regex = ["pattern1", "pattern2"]
    public = true
    from = "1970-01-01T00:00:00Z"
    until = "2100-01-01T00:00:00Z"
    passid = true
```

### Regex

The Regex Patterns are matched against both the Summary as well as the ID. This can be used to exclude one specific entry.

### From & Until

The From and Until value allow for excluding the Pattern only in the selected Timeframe.
Time has to be provided in compliance with RFC3339.

### PassID

Bool Value to allow passing the original EventIDs to the new calendar.
