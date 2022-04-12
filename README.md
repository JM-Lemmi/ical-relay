ical-relay
==========
Relay ical urls and edit them on the fly with different modules.

# Usage

You can download an example configuration file from [here](https://raw.githubusercontent.com/JM-Lemmi/ical-relay/master/config.yml.example).

The edited ical can be accessed on `http://server/profiles/profilename`

## Docker Container

```
docker run -d -p 8080:80 -v ~/ical-relay/config.yml:/app/config.yml ghcr.io/jm-lemmi/ical-relay
```

## Standalone

Download the binary from the newest release.
The configuration file has to be in your current directory, when starting ical-relay.

```
./ical-relay
```

## Build
* Run from source: `go run .`
* Build and run: `go build . && ./ical-relay`

# Config

```yaml
server:
    addr: ":80"
    loglevel: "info"

profiles:
    <profilename>:
        source: "https://example.com/calendar.ics"
        public: true
        modules:
        - name: "delete-bysummary-regex"
          regex: "testentry"
          from: "2021-12-02T00:00:00Z"
          until: "2021-12-31T00:00:00Z"
        - name: "add-url"
          url: "https://othersource.com/othercalendar.ics"
          header-Cookie: "MY_AUTH_COOKIE=abcdefgh"
```

The `server` section contains the configuration for the HTTP server. You can change the loglevel to "debug" to get more information.
You can list as many profiles as you want. Each profile has to have a source.
You can then add as many modules as you want. They are identified by the `name:`. All other fields are dependent on the module.
The modules are executed in the order they are listed and you can call a module multiple times.

# Modules

Feel free do open a PR with modules of your own.

## delete-bysummary-regex

Delete all entries with a summary matching the regex.
The module can be called with a from and/or until date in RFC3339 format.

## delete-byid

Deletes an entry by its id.

## add-url

Adds all events from the specified url.
The module can be called with a header-<headername> option to pass Authentication cookies or X-Forwarded-Host headers.

## add-file

Adds all events from the specified local file.

## delete-timeframe

Deletes all events in the specified timeframe. The timeframe is specified with a after and/or before date in RFC3339 format.
If only after is specified, all events after the date are deleted.
If only before is specified, all events before the date are deleted.
Other possible value is "now" to use current time as value.

## delete-duplicates

Deletes events, if there already is an event with the same start, end and summary.

## edit-byid

Edits an Event with the passed id.
Parameters:
- 'id', mandatory: the id of the event to edit
- 'new-summary', optional: the new summary
- 'new-description', optional: the new description
- 'new-start', optional: the new start time in RFC3339 format "2006-01-02T15:04:05Z"
- 'new-end', optional: the new end time in RFC3339 format "2006-01-02T15:04:05Z"
- 'new-location', optional: the new location

## edit-bysummary-regex

Edits all Events with the matching regex title.
Parameters:
- 'id', mandatory: the id of the event to edit
- 'after', optional: beginning of search timeframe
- 'before', optional: end of search timeframe
- 'new-summary', optional: the new summary
- 'new-description', optional: the new description
- 'new-start', optional: the new start time in RFC3339 format "2006-01-02T15:04:05Z"
- 'new-end', optional: the new end time in RFC3339 format "2006-01-02T15:04:05Z"
- 'new-location', optional: the new location

## save-to-file

This module saves the current calendar to a file.
Parameters: "file" mandatory: full path of file to save

# Special Combinations

Using the save-to-file Module and the delete-timeframe module with "now" you can create a calendar with immutable past. This stops the calendar events in the past from being updated.

```yaml
profiles:
  abc:
    source: "http://example.com/calendar.ics"
    modules:
    - name: "delete-timeframe"
      before: "now"
    - name: "add-url"
      url: "http://localhost/profiles/abc-past"
    - name: "save-to-file"
      file: "/app/calstore/abc-archive.ics"

  abc-past:
    source: ""
    modules:
    - name: "add-file"
      filename: "/app/calstore/abc-archive.ics"
    - name: "delete-timeframe"
      after: "now"
```
