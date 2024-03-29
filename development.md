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
