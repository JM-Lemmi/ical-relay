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

You can mount mount templates from `/opt/ical-relay/templates` to edit them. Please be aware of updates to the templates when choosing this method.

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

## Build

* Run from source: `go run .`
* Build and run: `go build . && ./ical-relay`

# Config

```yaml
server:
  addr: ":80"
  loglevel: "info"
  url: "https://cal.julian-lemmerich.de"
  templatepath: /opt/ical-relay/templates
  imprintlink: "https://your-imprint"
  privacypolicylink: "http://your-data-privacy-policy"
  mail:
    smtp_server: "mailout.julian-lemmerich.de"
    smtp_port: 25
    sender: "calnotification@julian-lemmerich.de"
  super-tokens:
  - rA4nhdhmr34lL6x6bLyGoJSHE9o9cA2BwjsMOeqV5SEzm61apcRRzWybtGVjLKiB

profiles:
  relay:
    source: "https://example.com/calendar.ics"
    public: true
    immutable-past: true
    admin-tokens:
    - eAn97Sa0BKHKk02O12lNsa1O5wXmqXAKrBYxRcTNsvZoU9tU4OVS6FH7EP4yFbEt
    rules:
      filters:
      - type: "regex"
        regex: "testentry"
        target: "summary"
      - type: "timeframe"
        from: "2021-12-02T00:00:00Z"
        until: "2021-12-31T00:00:00Z"
        operator: "and"
      action:
        type: "delete"
      expires: "2022-12-31T00:00:00Z"

notifiers:
  relay:
    source: "http://localhost/relay"
    interval: "15m"
    recipients:
    - email: "jm.lemmerich@gmail.com"
```

The `server` section contains the configuration for the HTTP server. You can change the loglevel to "debug" to get more information.
You can list as many profiles as you want. Each profile has to have a source.
You can then add as many rules as you want. The `name:` filed specifies the module, the rule references. All other fields are dependent on the module.
The rule are executed in the order they are listed. You can create multiple rules from one module.

## immutable-past

Add `immutable-past: true` in the profile to enable it.

If you enable immutable past, the relay will save all events that have already happened in a file called `<profile>-past.ics` in the storage path. Next time the profile is called, the past events will be added to the ical.

## Rules

A Rule contains one or more filters and one action. The filters determine which events will be edited. The action then determines, what changes for the events.

Feel free do open a PR with filters and actions of your own.

Adding `expires: <RFC3339>` to any rule will remove it on the next cleanup cycle after the date has passed. Currently the Cleanup runs every 1h.

### Filters

Currently the Filters can not handle Repeating Events. See Issue #77

#### regex

* `regex`: The regex to match against
* `target`: Parameter to match against the regex. Default Summary, options: Summary, Description, Location

#### id

* `id`: Event ID.

This can match multiple events, for example with repeating events.

#### timeframe

* `after`, `before`. At least one is mandatory. Uses max time, if none is given. Can also be set to "now".

#### duplicates

No parameters. Filters the second and following events that are identified as duplicate. Looks at start, end, summary. If all three are equal, the Event is deemed duplicate.

#### all

No parameters. Filters all.

#### duration

* `duration` in timeDuration format (most relevant: `m`, `h`)´
* `operator`. Either "longer" or "shorter", default "longer".

### Actions

#### delete

No parameters. Deletes the Filtered Events.

#### edit

* `new-summary`, optional: the new summary
* `new-description`, optional: the new description
* `new-start`, optional: the new start time in RFC3339 format "2006-01-02T15:04:05Z"
* `new-end`, optional: the new end time in RFC3339 format "2006-01-02T15:04:05Z"
* `new-location`, optional: the new location
* `overwrite`, default true: Possible values are 'true', 'false' and 'fillempty'. True: Overwrite the property if it already exists; False: Append, Fillempty: Only fills empty properties.  Does not apply to 'new-start' and 'new-end'.
* `move-time`, optional, not together with 'new-start' or 'new-end': add time to the whole entry, to move entry. uses Go ParseDuration: most useful units are "m", "h"
  * when the original time does not have a timezone, sets the timezone to UTC, so it needs to be adjusted for that.

#### add-reminder

* `time`: time in timeDuration Format the alarm will go off before the event.

This usually doesnt work when used in server mode. Most Calendar Applications ignore reminders of external calendars.

#### strip-info

* `mode`: "availibility" (puts busy status as summary, and removes all other information), or "limited" (only keeps summary and busy status)

Inspired by Outlooks export options.

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
