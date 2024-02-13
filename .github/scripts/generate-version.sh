#! /bin/bash
# also embedded in Dockerfile!

if [[ $(git tag -l --contains HEAD) ]]; then
    echo -n $(git tag -l --contains HEAD) > ./cmd/ical-relay/VERSION
else
    echo -n $(git rev-parse --short HEAD) > ./cmd/ical-relay/VERSION
fi
