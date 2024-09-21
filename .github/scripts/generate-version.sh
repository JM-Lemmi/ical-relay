#! /bin/bash
DIR="$(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
cd "$DIR" || exit 1

if [[ $(git tag -l --contains HEAD) ]]; then
    echo -n $(git tag -l --contains HEAD) > ../../cmd/ical-relay/VERSION
    echo -n $(git tag -l --contains HEAD) > ../../cmd/ical-notifier/VERSION
else
    echo -n $(git describe --tags --abbrev=0) > ../../cmd/ical-relay/VERSION
    echo -n $(git describe --tags --abbrev=0) > ../../cmd/ical-notifier/VERSION
    echo -n "+" >> ../../cmd/ical-relay/VERSION
    echo -n "+" >> ../../cmd/ical-notifier/VERSION
    echo -n $(git rev-parse --short HEAD) >> ../../cmd/ical-relay/VERSION
    echo -n $(git rev-parse --short HEAD) >> ../../cmd/ical-notifier/VERSION
fi
if ! git diff --quiet; then
    echo -n "-dirty" >> ../../cmd/ical-relay/VERSION
    echo -n "-dirty" >> ../../cmd/ical-notifier/VERSION
fi