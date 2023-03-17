ical-relay
==========
Relay ical urls and edit them on the fly with different modules.

# Usage

You can download an example configuration file from [here](https://raw.githubusercontent.com/JM-Lemmi/ical-relay/master/config.yml.example).

The edited ical can be accessed on `http://server/profiles/profilename`

## Install from Apt Repository

If you're running a Debian based system, you can install the latest release from the my Package Repository.

This allows automatic updates to the latest version with `apt upgrade`.

```bash
echo "deb [arch=amd64] http://pkg.julian-lemmerich.de/deb stable main" | tee /etc/apt/sources.list.d/jm-lemmi.list
curl http://pkg.julian-lemmerich.de/deb/gpg.key | apt-key add -
apt update
apt install ical-relay
```

If you want to run the beta version, you can use the `testing` repository. Replace `stable` with `testing` in the above commands.

This installs the ical-relay as a systemd service. Change the configuration in `/etc/ical-relay/config.yml` and start the service with `systemctl start ical-relay`.

## Run standalone Binary

## Debian package

Download the package from the latest release.

Install with your package manager:

```
apt install ./ical-relay_1.3.0_amd64.deb
```

This will create a systemd service called `ical-relay.service` which can be started with `systemctl start ical-relay.service`. The defualt configuration file is located at `/etc/ical-relay/config.yml`.

Run a notifier manually:

```
/usr/bin/ical-relay --notifier <name> --config config.yml
```

## Docker Container

```
docker run -d -p 8080:80 -v ~/ical-relay/:/etc/ical-relay/ ghcr.io/jm-lemmi/ical-relay
```

## Build from Source

Clone this repo then either:

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
      expires: "2022-12-06T00:00:00Z"
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

Adding `expires: <RFC3339>` to any module will remove it on the next cleanup cycle after the date has passed. Currently the Cleanup runs every 1h.

## immutable-past

Even though immutable past is not really a module, it is listed here, cause it fits.

Add `immutable-past: true` in the profile to enable it.

If you enable immutable past, the relay will save all events that have already happened in a file called `<profile>-past.ics` in the storage path. Next time the profile is called, the past events will be added to the ical.

## delete-bysummary-regex

* `regex`: The regex to match the summary against
* `from`, optional: Beginning of timeframe that should be deleted in, in RFC3339 format
* `until`, optional: End of timeframe that should be deleted in, in RFC3339 format

## delete-byid

* `id`: The id of the event to delete

## add-url

* `url`: Adds all events from the specified url.
* `header-<headername>`, optional: Adds a header to the request. Can be used to pass authentication cookies or X-Forwarded-Host headers.

## add-file

* `file`: Adds all events from the specified local file.

## delete-timeframe

Deletes all events in the specified timeframe.

* `after`:  Start of the timeframe to be deleted in RFC3339 format or "now" for current time as value. If only after is specified, all events after the date are deleted.
* `before`: End of the timeframe to be deleted in RFC3339 format or "now" for current time as value. If only before is specified, all events before the date are deleted.

## delete-duplicates

Deletes events, if there already is an event with the same start, end and summary.

No parameters.

## edit-byid

Edits an Event with the passed id.
Parameters:
* `id`: the id of the event to edit
* `overwrite`, default true: Possible values are 'true', 'false' and 'fillempty'. True: Overwrite the property if it already exists; False: Append, Fillempty: Only fills empty properties.  Does not apply to 'new-start' and 'new-end'.
* `new-summary`, optional: the new summary
* `new-description`, optional: the new description
* `new-start`, optional: the new start time in RFC3339 format "2006-01-02T15:04:05Z"
* `new-end`, optional: the new end time in RFC3339 format "2006-01-02T15:04:05Z"
* `new-location`, optional: the new location

## edit-bysummary-regex

Edits all Events with the matching regex title.
Parameters:
* `id`, mandatory: the id of the event to edit
* `overwrite`, default true: Possible values are 'true', 'false' and 'fillempty'. True: Overwrite the property if it already exists; False: Append, Fillempty: Only fills empty properties.  Does not apply to 'new-start' and 'new-end'.
* `after`, optional: beginning of search timeframe
* `before`, optional: end of search timeframe
* `new-summary`, optional: the new summary
* `new-description`, optional: the new description
* `new-start`, optional: the new start time in RFC3339 format "2006-01-02T15:04:05Z" or "now"
* `new-end`, optional: the new end time in RFC3339 format "2006-01-02T15:04:05Z" or "now"
* `new-location`, optional: the new location
* `move-time`, optional, not together with 'new-start' or 'new-end': add time to the whole entry, to move entry. uses Go ParseDuration: most useful units are "m", "h"

#### known issues:

`move-time`, when the original time does not have a timezone, sets the timezone to UTC, so it needs to be adjusted for that.

## save-to-file

This module saves the current calendar to a local file.

* `file`: full path of file to save

# API

For details about the API endpoints, see the swagger documentation at [./documentation/swagger.yml](./documentation/swagger.yml)

Autorization is done in three levels:

- Public: No token, can use all public endpoints.
- Profile-Admin: Token for a specific profile, can use most endpoints for this profile, but not all module types.
- Super-Admin: Rights for all profiles and can also use all modules. May include LFI or CSRF-capable config options. Should be used with caution.

# Notifier

The notifiers do not have to reference a local ical, you can also use this to only call external icals.

You can configure SMTP with authentication or without to use an external mailserver, or something local like boky/postfix.

If you start the calendar with the `--notifier` flag, it will start the notifier from config. This allows setting up cronjobs to run the notifier.

# Support

This Project was developed for my own use, and I do not offer support for this at all.

If you do want to use it and need help I will try my best to help, but I can't promise anyting. You can contact me here: help@julian-lemmerich.de
