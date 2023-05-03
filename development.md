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

## Using the embedded database for testing

For testing you can set server.db.host to Special:EMBEDDED
```
    db:
        host: Special:EMBEDDED
        db-name: ical_relay
```