# Informations for Development

## Cloning the Repo

Since there are submodules, you need to clone recursively:

```
git clone --recursive https://www.github.com/jm-lemmi/ical-relay
```

## Go linter with multiple modules

also see: https://github.com/golang/tools/blob/master/gopls/doc/workspace.md

adding a new pkg you also need to add it to `go work` to use it.

```
go work use ./pkg/<newname>
```

## compiling

please first generate the version number: `./.github/scripts/generate-version.sh` then build with `go build -o ./bin/ical-relay ./cmd/ical-relay/`

## Development Docker Compose

```
docker compose -f docker-compose.dev.yml up --build --force-recreate
```

## VTIMEZONE data

the VTIMEZONE compatibility modes need ics vtimezone data.
a github worklfow generates them, so easiest is to download them from the workflow. the workflow can be called via dispatch, so should be easily availible. Then download it to XXX to be included by go embed.
If not you can use the following very short instructions inside a debian docker container to generate them

```
git clone https://github.com/libical/vzic
apt install gcc libgtk-3-dev
```

in Makefile

```
OLSON_DIR = tzdata2024a
PRODUCT_ID = //jm-lemmi//ical-relay-timezones//EN
TZID_PREFIX =
```

olson dir can be set to a new olson download, but the current repo includes the 2024a already
tzid_prefix should be empty, because we dont need versioning of the tz

```
make -B
./vzic
```

this creates the folder zoneinfo with all ics files for the zones.
wen can then combine all the single ics files into a giant file to embed with the scirpt in the workflow (copy it out and run with bash)
