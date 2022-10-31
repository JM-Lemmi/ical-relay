ical-relay
==========
Relay ical urls and edit them on the fly with different modules.

# Usage

You can download an example configuration file from [here](https://raw.githubusercontent.com/JM-Lemmi/ical-relay/master/config.yml.example).

The edited ical can be accessed on `http://server/profiles/profilename`

## Docker Container

```
docker run -d -p 8080:80 -v ~/ical-relay/:/etc/ical-relay/ ghcr.io/jm-lemmi/ical-relay
```

## Standalone

Download the binary from the newest release.

```
./ical-relay --config config.yml
```

Run a notifier manually:

```
./ical-relay --notifier <name> --config config.yml
```

## Build
* Run from source: `go run .`
* Build and run: `go build . && ./ical-relay`

# Config

```yaml
server:
  addr: ":80"
  loglevel: "info"
  storagepath: "/etc/ical-relay/"

profiles:
  relay:
    source: "https://example.com/calendar.ics"
    public: true
    immutable-past: true
    modules:
    - name: "delete-bysummary-regex"
      regex: "testentry"
      from: "2021-12-02T00:00:00Z"
      until: "2021-12-31T00:00:00Z"
    - name: "add-url"
      url: "https://othersource.com/othercalendar.ics"
      header-Cookie: "MY_AUTH_COOKIE=abcdefgh"

notifiers:
  relay:
    source: "http://localhost/relay"
    interval: "15m"
    smtp_server: "mailout.julian-lemmerich.de"
    smtp_port: "25"
    sender: "calnotification@julian-lemmerich.de"
    recipients:
    - email: "jm.lemmerich@gmail.com"
```

The `server` section contains the configuration for the HTTP server. You can change the loglevel to "debug" to get more information.
You can list as many profiles as you want. Each profile has to have a source.
You can then add as many modules as you want. They are identified by the `name:`. All other fields are dependent on the module.
The modules are executed in the order they are listed and you can call a module multiple times.

# Modules

Feel free do open a PR with modules of your own.

## immutable-past

Even though immutable past is not really a module, it is treated as such.

If you enable immutable past, the relay will save all events that have already happened in a file called `<profile>-past.ics` in the storage path. Next time the profile is called, the past events will be added to the ical.

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
- 'overwrite', default true: Possible values are 'true', 'false' and 'fillempty'. True: Overwrite the property if it already exists; False: Append, Fillempty: Only fills empty properties.  Does not apply to 'new-start' and 'new-end'.
- 'new-summary', optional: the new summary
- 'new-description', optional: the new description
- 'new-start', optional: the new start time in RFC3339 format "2006-01-02T15:04:05Z"
- 'new-end', optional: the new end time in RFC3339 format "2006-01-02T15:04:05Z"
- 'new-location', optional: the new location

## edit-bysummary-regex

Edits all Events with the matching regex title.
Parameters:
- 'id', mandatory: the id of the event to edit
- 'overwrite', default true: Possible values are 'true', 'false' and 'fillempty'. True: Overwrite the property if it already exists; False: Append, Fillempty: Only fills empty properties.  Does not apply to 'new-start' and 'new-end'.
- 'after', optional: beginning of search timeframe
- 'before', optional: end of search timeframe
- 'new-summary', optional: the new summary
- 'new-description', optional: the new description
- 'new-start', optional: the new start time in RFC3339 format "2006-01-02T15:04:05Z" or "now"
- 'new-end', optional: the new end time in RFC3339 format "2006-01-02T15:04:05Z" or "now"
- 'new-location', optional: the new location
- 'move-time', optional, not together with 'new-start' or 'new-end': add time to the whole entry, to move entry. uses Go ParseDuration: most useful units are "m", "h"

#### known issues:

'move-time', when the original time does not have a timezone, sets the timezone to UTC, so it needs to be adjusted for that.

## save-to-file

This module saves the current calendar to a file.
Parameters: "file" mandatory: full path of file to save

# API

- `/api/calendars`: Returns all Public Calendars as json-array.
- `/api/reloadconfig`: Reloads the config from disk.
- `/api/notifier/<notifier>/addrecipient`: with an E-Mail Address as body adds the recipient to the notifier.

# Notifier

The notifiers do not have to reference a local ical, you can also use this to only call external icals.

You can configure SMTP with authentication or without to use an external mailserver, or something local like boky/postfix.

If you start the calendar with the `--notifier` flag, it will start the notifier from config. This allows setting up cronjobs to run the notifier.

# WebUI

You can use the [frereit/react-calendar](https://github.com/frereit/react-calendar/) webui to view a calendar.

1. Download the react-calendar release and unpack into an nginx static directory.
2. Configure nginx to serve the react-calendar webui at `/` and the ical-relay at `/profiles`<br>Example:

```conf
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;

    server_name cal.julian-lemmerich.de;

    ssl_certificate /etc/nginx/ssl/live/cal.julian-lemmerich.de/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/live/cal.julian-lemmerich.de/privkey.pem;
    
    location /profiles/ {
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP         $remote_addr;
	    proxy_buffering                    off;
    	proxy_pass http://ical/profiles/;
    }

    location / {
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP         $remote_addr;
	    proxy_buffering                    off;
    	proxy_pass http://static/reactcal/;
    }
}
```

3. Edit the config.js file with the ical file you want to view in the WebUI.

# Support

This Project was developed for my own use, and I do not offer support for this at all.

If you do want to use it and need help I will try my best to help, but I can't promise anyting. You can contact me here: help@julian-lemmerich.de
