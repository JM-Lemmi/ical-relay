Collection of tools for using ical calendars:

- **ical-relay**: Webserver to relay ical urls and edit them on the fly with different modules.
- **ical-notifier**: Tool for detecting changes and notifying different targets about it.

# Usage

## ical-relay

ical-relay can be used with a simple feature set (lite-mode) or an extended feature set, including a frontend and api.

A calendar is called a "profile". It takes one or more sources and then edits them with a set of filters and actions. The edited ical can be accessed on `http://server/profiles/profilename`

For persistent configuration changes you will need a postgress database. See [Config](#config) for more information.

## ical-notifier

ical-notifier can be used standalone or in conjunction with ical-relay in full mode for dynamic subscription and frontend.

Notifiers have a source and a list of recipients. The Source can be a local ical-relay via localhost, but you can also use a remote source via http. Recipients can be E-Mail, Webhook or an RSS output.

The notifier is meant to be run by an external timer like systemd-timer or cron.

# Installation

## Apt Repository

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

The package also includes the ical-notifier. If you `systemctl enable ical-notifier.timer` the notifier will run every 15 minutes.

## Install Package from Github

Download the package from the latest release.

Install with your package manager:

```
apt install ./ical-relay_1.3.0_amd64.deb
```

This installs the ical-relay as a systemd service. Change the configuration in `/etc/ical-relay/config.yml` and start the service with `systemctl start ical-relay`.

The package also includes the ical-notifier. If you `systemctl enable ical-notifier.timer` the notifier will run every 15 minutes.

## Docker

Docker images are availible in arm and x64 variants.

```
docker run ghcr.io/jm-lemmi/ical-relay:latest -p 80:80 -v ./config:/etc/ical-relay/config
```

The default health check only works, when ical-relay is listening on port 80 (inside the container)

There is also a docker container for ical-notifier using crond to time the ical-notifier.

# Configuration

The configuration is split in two:

- `config.yml` contains all the general server/service configuration.
- `data.yml` contains profiles and notifier information. You can use the same file for ical-relay and ical-notifier.

Example files are included in the installation or in this repo.

You can list as many profiles or notifiers as you want. \
You can then add as many rules as you want. A rule can contain multiple filters and one action. \
A profile can have enable immutable past, the relay will save all events that have already happened in a file called `<profile>-past.ics` in the storage path. Next time the profile is called, the past events will be used from storage instead of upstream.

To import data into a DB when running full mode, use the `--import-data` flag.

### config.yml versioning

| ical-relay version | config version |
|--------------------|----------------|
| 2.0.0-beta.5       | 2              |
| ?                  | 3              |
| 2.0.0-beta.9       | 4 !            |

### data.yml versioning

| ical-relay version | data version   |
|--------------------|----------------|
| 2.0.0-beta.9       | 1              |

### database versioning

| ical-relay version | db version |
|--------------------|------------|
| 2.0.0-beta.6       | 4          |
| 2.0.0-beta.9       | 5          |

# Lite-Mode

Running in Lite-Mode disables the frontend and api and doesnt need a database. It reads the profiles and rules from the `data.yaml`.

You can use litemode either with the --lite-mode flag or with the config option "lite-mode: true".
By default ical-relay will start in full mode.

Immutable-Past Files are still written to file in lite mode.

# Rules

A Rule contains one or more filters and one action. The filters determine which events will be edited. The action then determines, what changes for the events.

Feel free do open a PR with filters and actions of your own.

Adding `expires: <RFC3339>` to any rule will remove it on the next cleanup cycle after the date has passed. Currently the Cleanup runs every 1h.

You can find detailed information on all the different rules at [./documentation/filters.yml](./documentation/filters.md)

# API

For details about the API endpoints, see the swagger documentation at [./documentation/swagger.yml](./documentation/swagger.yml)

Autorization is done in three levels:

- Public: No token, can use all public endpoints.
- Profile-Admin: Token for a specific profile, can use most endpoints for this profile, but not all module types.
- Super-Admin: Rights for all profiles and can also use all modules. May include LFI or CSRF-capable config options. Should be used with caution.

### Adding an Event from File via API:

```
curl -F eventfile=@./testfile.ics -H "Authorization: <token>" http://localhost/api/profiles/test/newentryfile
```

# Combining multiple Profiles into one ICS

http://localhost:8080/profiles-combi/profile1+profile2

# Development

I am happy for PRs of any features, like new Actions, Filters or Bug Fixes.

See some more information for development at [development.md](./development.md)

# Support

This Project was developed for my own use and it ran away from there. I do not offer active support for this.

If you do want to use it and need help I will try my best to help, but I can't promise anyting. You can contact me here: help@julian-lemmerich.de
